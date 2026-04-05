package publicdotcom

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const baseURL = "https://api.public.com"

const version = "0.1.0"

var userAgent = "publicdotcom-go/" + version

// Client is an HTTP client for the Public.com trading API.
type Client struct {
	http      *http.Client
	secretKey string
	accountID string

	accessToken  string
	tokenExpires time.Time
}

// ClientOption configures a [Client].
type ClientOption func(*Client)

// WithRetry enables automatic retries for 429 and 5xx responses using the
// given [RetryPolicy]. If policy is nil, [DefaultRetryPolicy] is used.
func WithRetry(policy *RetryPolicy) ClientOption {
	return func(c *Client) {
		if policy == nil {
			policy = DefaultRetryPolicy()
		}
		base := innerTransport(c.http.Transport)
		c.http.Transport = &retryTransport{
			base:   base,
			policy: policy,
		}
	}
}

// WithHTTPClient replaces the default HTTP client.
func WithHTTPClient(hc *http.Client) ClientOption {
	return func(c *Client) { c.http = hc }
}

// NewClient creates a Public.com API client. By default the client blocks
// plaintext HTTP requests and sets a User-Agent header identifying the SDK.
func NewClient(secretKey, accountID string, opts ...ClientOption) *Client {
	c := &Client{
		http: &http.Client{
			Timeout:   30 * time.Second,
			Transport: &httpsOnlyTransport{base: http.DefaultTransport},
		},
		secretKey: secretKey,
		accountID: accountID,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// httpsOnlyTransport rejects any request whose URL scheme is not "https".
type httpsOnlyTransport struct {
	base http.RoundTripper
}

// ErrInsecureRequest is returned when a request uses plaintext HTTP.
var ErrInsecureRequest = errors.New("insecure HTTP requests are not allowed; use HTTPS")

func (t *httpsOnlyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if !strings.EqualFold(req.URL.Scheme, "https") {
		return nil, ErrInsecureRequest
	}
	return t.base.RoundTrip(req)
}

// innerTransport unwraps the current transport chain to the innermost
// real transport, skipping httpsOnlyTransport. This lets WithRetry wrap
// the HTTPS guard rather than doubling it.
func innerTransport(t http.RoundTripper) http.RoundTripper {
	if t == nil {
		return http.DefaultTransport
	}
	if h, ok := t.(*httpsOnlyTransport); ok {
		return h
	}
	return t
}

// Authenticate obtains a personal access token. Tokens are cached and
// re-used until they expire. This is called automatically before each
// request; you only need to call it explicitly if you want to fail fast
// on bad credentials.
func (c *Client) Authenticate(ctx context.Context) error {
	if c.accessToken != "" && time.Now().Before(c.tokenExpires) {
		return nil
	}

	body := map[string]any{
		"validityInMinutes": 60,
		"secret":            c.secretKey,
	}
	resp, err := c.doJSON(ctx, http.MethodPost, "/userapiauthservice/personal/access-tokens", body, false)
	if err != nil {
		return fmt.Errorf("authenticate: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken string `json:"accessToken"`
	}
	if err := decodeResponse(resp, &result); err != nil {
		return fmt.Errorf("authenticate: %w", err)
	}

	c.accessToken = result.AccessToken
	c.tokenExpires = time.Now().Add(55 * time.Minute) // refresh 5 min early
	return nil
}

// --- internal helpers ---

func (c *Client) authedGet(ctx context.Context, path string) (json.RawMessage, error) {
	if err := c.Authenticate(ctx); err != nil {
		return nil, err
	}
	req, err := c.newAuthedRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GET %s: %w", path, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, readAPIError(resp)
	}
	return io.ReadAll(resp.Body)
}

func (c *Client) authedPost(ctx context.Context, path string, body any) (json.RawMessage, error) {
	if err := c.Authenticate(ctx); err != nil {
		return nil, err
	}
	req, err := c.newAuthedRequest(ctx, http.MethodPost, path, body)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("POST %s: %w", path, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, readAPIError(resp)
	}
	return io.ReadAll(resp.Body)
}

func (c *Client) authedPut(ctx context.Context, path string, body any) (json.RawMessage, error) {
	if err := c.Authenticate(ctx); err != nil {
		return nil, err
	}
	req, err := c.newAuthedRequest(ctx, http.MethodPut, path, body)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("PUT %s: %w", path, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, readAPIError(resp)
	}
	return io.ReadAll(resp.Body)
}

func (c *Client) authedDelete(ctx context.Context, path string) error {
	if err := c.Authenticate(ctx); err != nil {
		return err
	}
	req, err := c.newAuthedRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("DELETE %s: %w", path, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return readAPIError(resp)
	}
	return nil
}

func (c *Client) newAuthedRequest(ctx context.Context, method, path string, body any) (*http.Request, error) {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, baseURL+path, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("User-Agent", userAgent)
	return req, nil
}

func (c *Client) doJSON(ctx context.Context, method, path string, body any, auth bool) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, baseURL+path, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent)
	if auth && c.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s %s: %w", method, path, err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		return nil, readAPIError(resp)
	}
	return resp, nil
}

func decodeResponse(resp *http.Response, v any) error {
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return readAPIError(resp)
	}
	return json.NewDecoder(resp.Body).Decode(v)
}

// decode unmarshals a json.RawMessage into the given type.
func decode[T any](raw json.RawMessage, err error) (*T, error) {
	if err != nil {
		return nil, err
	}
	var v T
	if err := json.Unmarshal(raw, &v); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &v, nil
}

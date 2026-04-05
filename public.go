package publicdotcom

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const baseURL = "https://api.public.com"

// Client is an HTTP client for the Public.com trading API.
type Client struct {
	http      *http.Client
	secretKey string
	accountID string

	accessToken  string
	tokenExpires time.Time
}

// NewClient creates a Public.com API client.
func NewClient(secretKey, accountID string) *Client {
	return &Client{
		http:      &http.Client{Timeout: 30 * time.Second},
		secretKey: secretKey,
		accountID: accountID,
	}
}

// Authenticate obtains a personal access token. Tokens are cached and
// re-used until they expire.
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

// --- Accounts ---

func (c *Client) GetAccounts(ctx context.Context) (json.RawMessage, error) {
	return c.authedGet(ctx, "/userapigateway/trading/account")
}

// --- Portfolio ---

func (c *Client) GetPortfolio(ctx context.Context) (json.RawMessage, error) {
	return c.authedGet(ctx, fmt.Sprintf("/userapigateway/trading/%s/portfolio/v2", c.accountID))
}

// --- History ---

func (c *Client) GetHistory(ctx context.Context, params *HistoryParams) (json.RawMessage, error) {
	path := fmt.Sprintf("/userapigateway/trading/%s/history", c.accountID)
	req, err := c.newAuthedRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	if params != nil {
		q := req.URL.Query()
		if params.Start != "" {
			q.Set("start", params.Start)
		}
		if params.End != "" {
			q.Set("end", params.End)
		}
		if params.PageSize > 0 {
			q.Set("pageSize", fmt.Sprintf("%d", params.PageSize))
		}
		if params.NextToken != "" {
			q.Set("nextToken", params.NextToken)
		}
		req.URL.RawQuery = q.Encode()
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get history: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, readAPIError(resp)
	}
	return io.ReadAll(resp.Body)
}

// --- Instruments ---

func (c *Client) GetInstruments(ctx context.Context) (json.RawMessage, error) {
	return c.authedGet(ctx, "/userapigateway/trading/instruments")
}

func (c *Client) GetInstrument(ctx context.Context, symbol, instrumentType string) (json.RawMessage, error) {
	return c.authedGet(ctx, fmt.Sprintf("/userapigateway/trading/instruments/%s/%s", symbol, instrumentType))
}

// --- Market Data ---

func (c *Client) GetQuotes(ctx context.Context, instruments []Instrument) (json.RawMessage, error) {
	path := fmt.Sprintf("/userapigateway/marketdata/%s/quotes", c.accountID)
	return c.authedPost(ctx, path, QuoteRequest{Instruments: instruments})
}

func (c *Client) GetOptionExpirations(ctx context.Context, instrument Instrument) (json.RawMessage, error) {
	path := fmt.Sprintf("/userapigateway/marketdata/%s/option-expirations", c.accountID)
	return c.authedPost(ctx, path, map[string]any{"instrument": instrument})
}

func (c *Client) GetOptionChain(ctx context.Context, req OptionChainRequest) (json.RawMessage, error) {
	path := fmt.Sprintf("/userapigateway/marketdata/%s/option-chain", c.accountID)
	return c.authedPost(ctx, path, req)
}

// --- Orders ---

// PreflightOrder validates a single-leg order without placing it.
func (c *Client) PreflightOrder(ctx context.Context, order OrderRequest) (json.RawMessage, error) {
	path := fmt.Sprintf("/userapigateway/trading/%s/preflight/single-leg", c.accountID)
	return c.authedPost(ctx, path, order)
}

// PreflightMultiLegOrder validates a multi-leg order without placing it.
func (c *Client) PreflightMultiLegOrder(ctx context.Context, order MultiLegOrderRequest) (json.RawMessage, error) {
	path := fmt.Sprintf("/userapigateway/trading/%s/preflight/multi-leg", c.accountID)
	return c.authedPost(ctx, path, order)
}

// PlaceOrder submits a single-leg order for execution.
func (c *Client) PlaceOrder(ctx context.Context, order OrderRequest) (json.RawMessage, error) {
	path := fmt.Sprintf("/userapigateway/trading/%s/order", c.accountID)
	return c.authedPost(ctx, path, order)
}

// PlaceMultiLegOrder submits a multi-leg order for execution.
func (c *Client) PlaceMultiLegOrder(ctx context.Context, order MultiLegOrderRequest) (json.RawMessage, error) {
	path := fmt.Sprintf("/userapigateway/trading/%s/order/multileg", c.accountID)
	return c.authedPost(ctx, path, order)
}

// GetOrder retrieves the status of an order.
func (c *Client) GetOrder(ctx context.Context, orderID string) (json.RawMessage, error) {
	return c.authedGet(ctx, fmt.Sprintf("/userapigateway/trading/%s/order/%s", c.accountID, orderID))
}

// CancelOrder cancels a pending order.
func (c *Client) CancelOrder(ctx context.Context, orderID string) error {
	path := fmt.Sprintf("/userapigateway/trading/%s/order/%s", c.accountID, orderID)
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

// --- Option Details ---

func (c *Client) GetOptionGreeks(ctx context.Context, osiSymbol string) (json.RawMessage, error) {
	return c.authedGet(ctx, fmt.Sprintf("/userapigateway/option-details/%s/%s/greeks", c.accountID, osiSymbol))
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

// APIError represents an error response from the Public.com API.
type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("public.com API error %d: %s", e.StatusCode, e.Body)
}

func readAPIError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)
	return &APIError{StatusCode: resp.StatusCode, Body: string(body)}
}

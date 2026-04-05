package publicdotcom

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// APIError is the base error type for all Public.com API errors. More specific
// error types embed it: [AuthenticationError], [RateLimitError],
// [ValidationError], [NotFoundError], and [ServerError].
// Use errors.As to check for specific error types.
type APIError struct {
	StatusCode int
	Body       string
	Code       string // API error code, if present (e.g. "user_api_auth.personal.invalid_secret")
}

func (e *APIError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("public.com API error %d (%s): %s", e.StatusCode, e.Code, e.Body)
	}
	return fmt.Sprintf("public.com API error %d: %s", e.StatusCode, e.Body)
}

// AuthenticationError is returned for 401 Unauthorized responses.
type AuthenticationError struct {
	APIError
}

func (e *AuthenticationError) Error() string {
	return fmt.Sprintf("authentication failed: %s", e.APIError.Error())
}

func (e *AuthenticationError) Unwrap() error { return &e.APIError }

// RateLimitError is returned for 429 Too Many Requests responses.
// RetryAfter indicates when the request can be retried, if the server
// provided a Retry-After header.
type RateLimitError struct {
	APIError
	RetryAfter time.Duration
}

func (e *RateLimitError) Error() string {
	if e.RetryAfter > 0 {
		return fmt.Sprintf("rate limited (retry after %s): %s", e.RetryAfter, e.APIError.Error())
	}
	return fmt.Sprintf("rate limited: %s", e.APIError.Error())
}

func (e *RateLimitError) Unwrap() error { return &e.APIError }

// ValidationError is returned for 400 Bad Request responses.
type ValidationError struct {
	APIError
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed: %s", e.APIError.Error())
}

func (e *ValidationError) Unwrap() error { return &e.APIError }

// NotFoundError is returned for 404 Not Found responses.
type NotFoundError struct {
	APIError
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("not found: %s", e.APIError.Error())
}

func (e *NotFoundError) Unwrap() error { return &e.APIError }

// ServerError is returned for 5xx responses.
type ServerError struct {
	APIError
}

func (e *ServerError) Error() string {
	return fmt.Sprintf("server error: %s", e.APIError.Error())
}

func (e *ServerError) Unwrap() error { return &e.APIError }

// readAPIError reads the response body and returns a typed error based on
// the HTTP status code.
func readAPIError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	// Try to extract an error code from the response body.
	var code string
	var parsed struct {
		Code string `json:"code"`
	}
	if json.Unmarshal(body, &parsed) == nil && parsed.Code != "" {
		code = parsed.Code
	}

	base := APIError{
		StatusCode: resp.StatusCode,
		Body:       bodyStr,
		Code:       code,
	}

	switch {
	case resp.StatusCode == http.StatusUnauthorized:
		return &AuthenticationError{APIError: base}

	case resp.StatusCode == http.StatusTooManyRequests:
		retryAfter := parseRetryAfter(resp.Header.Get("Retry-After"))
		return &RateLimitError{APIError: base, RetryAfter: retryAfter}

	case resp.StatusCode == http.StatusBadRequest:
		return &ValidationError{APIError: base}

	case resp.StatusCode == http.StatusNotFound:
		return &NotFoundError{APIError: base}

	case resp.StatusCode >= 500:
		return &ServerError{APIError: base}

	default:
		return &base
	}
}

// parseRetryAfter parses the Retry-After header value, which can be
// either a number of seconds or an HTTP-date.
func parseRetryAfter(value string) time.Duration {
	if value == "" {
		return 0
	}
	if secs, err := strconv.Atoi(value); err == nil {
		return time.Duration(secs) * time.Second
	}
	if t, err := http.ParseTime(value); err == nil {
		d := time.Until(t)
		if d > 0 {
			return d
		}
	}
	return 0
}

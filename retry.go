package publicdotcom

import (
	"errors"
	"math"
	"math/rand/v2"
	"net/http"
	"time"
)

// RetryPolicy configures automatic retry behavior for failed requests.
type RetryPolicy struct {
	// MaxRetries is the maximum number of retry attempts (default 3).
	MaxRetries int
	// BaseDelay is the initial delay before the first retry (default 500ms).
	// Subsequent retries use exponential backoff: baseDelay * 2^attempt.
	BaseDelay time.Duration
	// MaxDelay caps the backoff delay (default 30s).
	MaxDelay time.Duration
}

// DefaultRetryPolicy returns a retry policy with sensible defaults:
// 3 retries, 500ms base delay, 30s max delay.
func DefaultRetryPolicy() *RetryPolicy {
	return &RetryPolicy{
		MaxRetries: 3,
		BaseDelay:  500 * time.Millisecond,
		MaxDelay:   30 * time.Second,
	}
}

// retryTransport wraps an [http.RoundTripper] and retries requests that
// fail with 429 or 5xx status codes.
type retryTransport struct {
	base   http.RoundTripper
	policy *RetryPolicy
}

func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	for attempt := range t.policy.MaxRetries + 1 {
		resp, err = t.base.RoundTrip(req)
		if err != nil {
			return nil, err
		}

		if !t.shouldRetry(resp.StatusCode) || attempt == t.policy.MaxRetries {
			return resp, nil
		}

		// Determine wait duration: prefer Retry-After header, else backoff.
		wait := t.backoff(attempt)
		if resp.StatusCode == http.StatusTooManyRequests {
			if ra := parseRetryAfter(resp.Header.Get("Retry-After")); ra > 0 {
				wait = ra
			}
		}

		// Drain and close the body so the connection can be reused.
		resp.Body.Close()

		// Wait with context cancellation support.
		timer := time.NewTimer(wait)
		select {
		case <-req.Context().Done():
			timer.Stop()
			return nil, req.Context().Err()
		case <-timer.C:
		}
	}

	return resp, err
}

func (t *retryTransport) shouldRetry(statusCode int) bool {
	return statusCode == http.StatusTooManyRequests || statusCode >= 500
}

func (t *retryTransport) backoff(attempt int) time.Duration {
	delay := t.policy.BaseDelay * time.Duration(math.Pow(2, float64(attempt)))
	if delay > t.policy.MaxDelay {
		delay = t.policy.MaxDelay
	}
	// Add jitter: 75%-125% of delay.
	jitter := 0.75 + rand.Float64()*0.5
	return time.Duration(float64(delay) * jitter)
}

// IsRetryable reports whether an error from the client is retryable
// (i.e. a rate limit or server error that may succeed on retry).
func IsRetryable(err error) bool {
	var rl *RateLimitError
	if errors.As(err, &rl) {
		return true
	}
	var se *ServerError
	return errors.As(err, &se)
}

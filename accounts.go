package publicdotcom

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
)

// GetAccounts retrieves the list of financial accounts associated with the
// authenticated user, including brokerage, retirement, and high-yield cash
// accounts. Unmarshal the result into [AccountsResponse].
func (c *Client) GetAccounts(ctx context.Context) (json.RawMessage, error) {
	return c.authedGet(ctx, "/userapigateway/trading/account")
}

// GetPortfolio retrieves positions, open orders, buying power, and equity
// summaries for the configured account. Unmarshal the result into
// [PortfolioResponse].
func (c *Client) GetPortfolio(ctx context.Context) (json.RawMessage, error) {
	return c.authedGet(ctx, fmt.Sprintf("/userapigateway/trading/%s/portfolio/v2", c.accountID))
}

// GetHistory retrieves the transaction history for the configured account.
// Pass nil for params to use server defaults. Results are paginated; use
// [HistoryParams.NextToken] for subsequent pages. Unmarshal the result into
// [HistoryResponse].
func (c *Client) GetHistory(ctx context.Context, params *HistoryParams) (json.RawMessage, error) {
	path := fmt.Sprintf("/userapigateway/trading/%s/history", c.accountID)
	req, err := c.newAuthedRequest(ctx, "GET", path, nil)
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

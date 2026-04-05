package publicdotcom

import (
	"context"
	"fmt"
	"io"
)

// GetAccounts retrieves the list of financial accounts associated with the
// authenticated user, including brokerage, retirement, and high-yield cash
// accounts.
func (c *Client) GetAccounts(ctx context.Context) (*AccountsResponse, error) {
	return decode[AccountsResponse](c.authedGet(ctx, "/userapigateway/trading/account"))
}

// GetPortfolio retrieves positions, open orders, buying power, and equity
// summaries for the configured account.
func (c *Client) GetPortfolio(ctx context.Context) (*PortfolioResponse, error) {
	return decode[PortfolioResponse](c.authedGet(ctx, fmt.Sprintf("/userapigateway/trading/%s/portfolio/v2", c.accountID)))
}

// GetHistory retrieves the transaction history for the configured account.
// Pass nil for params to use server defaults. Results are paginated; use
// [HistoryParams.NextToken] for subsequent pages.
func (c *Client) GetHistory(ctx context.Context, params *HistoryParams) (*HistoryResponse, error) {
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
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return decode[HistoryResponse](raw, nil)
}

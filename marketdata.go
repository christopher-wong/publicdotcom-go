package publicdotcom

import (
	"context"
	"encoding/json"
	"fmt"
)

// GetQuotes returns market quotes for the given instruments. Unmarshal the
// result into [QuoteResponse].
func (c *Client) GetQuotes(ctx context.Context, instruments []Instrument) (json.RawMessage, error) {
	path := fmt.Sprintf("/userapigateway/marketdata/%s/quotes", c.accountID)
	return c.authedPost(ctx, path, QuoteRequest{Instruments: instruments})
}

// GetOptionExpirations returns available expiration dates for the given
// instrument. Unmarshal the result into [OptionExpirationsResponse].
func (c *Client) GetOptionExpirations(ctx context.Context, instrument Instrument) (json.RawMessage, error) {
	path := fmt.Sprintf("/userapigateway/marketdata/%s/option-expirations", c.accountID)
	return c.authedPost(ctx, path, OptionExpirationsRequest{Instrument: instrument})
}

// GetOptionChain returns call and put quotes for an instrument at a specific
// expiration date. Unmarshal the result into [OptionChainResponse].
func (c *Client) GetOptionChain(ctx context.Context, req OptionChainRequest) (json.RawMessage, error) {
	path := fmt.Sprintf("/userapigateway/marketdata/%s/option-chain", c.accountID)
	return c.authedPost(ctx, path, req)
}

// GetOptionGreeks retrieves the greeks (delta, gamma, theta, vega, rho, IV)
// for an option identified by its OSI symbol. Unmarshal the result into
// [GreeksResponse].
func (c *Client) GetOptionGreeks(ctx context.Context, osiSymbol string) (json.RawMessage, error) {
	return c.authedGet(ctx, fmt.Sprintf("/userapigateway/option-details/%s/%s/greeks", c.accountID, osiSymbol))
}

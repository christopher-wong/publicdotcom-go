package publicdotcom

import (
	"context"
	"fmt"
)

// GetQuotes returns market quotes for the given instruments.
func (c *Client) GetQuotes(ctx context.Context, instruments []Instrument) (*QuoteResponse, error) {
	path := fmt.Sprintf("/userapigateway/marketdata/%s/quotes", c.accountID)
	return decode[QuoteResponse](c.authedPost(ctx, path, QuoteRequest{Instruments: instruments}))
}

// GetOptionExpirations returns available expiration dates for the given
// instrument.
func (c *Client) GetOptionExpirations(ctx context.Context, instrument Instrument) (*OptionExpirationsResponse, error) {
	path := fmt.Sprintf("/userapigateway/marketdata/%s/option-expirations", c.accountID)
	return decode[OptionExpirationsResponse](c.authedPost(ctx, path, OptionExpirationsRequest{Instrument: instrument}))
}

// GetOptionChain returns call and put quotes for an instrument at a specific
// expiration date.
func (c *Client) GetOptionChain(ctx context.Context, req OptionChainRequest) (*OptionChainResponse, error) {
	path := fmt.Sprintf("/userapigateway/marketdata/%s/option-chain", c.accountID)
	return decode[OptionChainResponse](c.authedPost(ctx, path, req))
}

// GetOptionGreeks retrieves the greeks (delta, gamma, theta, vega, rho, IV)
// for an option identified by its OSI symbol.
func (c *Client) GetOptionGreeks(ctx context.Context, osiSymbol string) (*GreeksResponse, error) {
	return decode[GreeksResponse](c.authedGet(ctx, fmt.Sprintf("/userapigateway/option-details/%s/%s/greeks", c.accountID, osiSymbol)))
}

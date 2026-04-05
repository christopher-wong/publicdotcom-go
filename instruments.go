package publicdotcom

import (
	"context"
	"fmt"
)

// GetInstruments retrieves all available instruments and their trading
// permissions.
func (c *Client) GetInstruments(ctx context.Context) (*InstrumentsResponse, error) {
	return decode[InstrumentsResponse](c.authedGet(ctx, "/userapigateway/trading/instruments"))
}

// GetInstrument retrieves details for a single instrument identified by symbol
// and type (e.g. "AAPL", "EQUITY").
func (c *Client) GetInstrument(ctx context.Context, symbol, instrumentType string) (*InstrumentDetail, error) {
	return decode[InstrumentDetail](c.authedGet(ctx, fmt.Sprintf("/userapigateway/trading/instruments/%s/%s", symbol, instrumentType)))
}

package publicdotcom

import (
	"context"
	"encoding/json"
	"fmt"
)

// GetInstruments retrieves all available instruments and their trading
// permissions. Unmarshal the result into [InstrumentsResponse].
func (c *Client) GetInstruments(ctx context.Context) (json.RawMessage, error) {
	return c.authedGet(ctx, "/userapigateway/trading/instruments")
}

// GetInstrument retrieves details for a single instrument identified by symbol
// and type (e.g. "AAPL", "EQUITY"). Unmarshal the result into [InstrumentDetail].
func (c *Client) GetInstrument(ctx context.Context, symbol, instrumentType string) (json.RawMessage, error) {
	return c.authedGet(ctx, fmt.Sprintf("/userapigateway/trading/instruments/%s/%s", symbol, instrumentType))
}

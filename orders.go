package publicdotcom

import (
	"context"
	"encoding/json"
	"fmt"
)

// PreflightOrder validates a single-leg order without placing it. Use this to
// check estimated cost, fees, margin requirements, and buying power impact
// before submitting. Unmarshal the result into [PreflightResponse].
func (c *Client) PreflightOrder(ctx context.Context, order OrderRequest) (json.RawMessage, error) {
	path := fmt.Sprintf("/userapigateway/trading/%s/preflight/single-leg", c.accountID)
	return c.authedPost(ctx, path, order)
}

// PreflightMultiLegOrder validates a multi-leg order without placing it.
// Unmarshal the result into [PreflightMultiLegResponse].
func (c *Client) PreflightMultiLegOrder(ctx context.Context, order MultiLegOrderRequest) (json.RawMessage, error) {
	path := fmt.Sprintf("/userapigateway/trading/%s/preflight/multi-leg", c.accountID)
	return c.authedPost(ctx, path, order)
}

// PlaceOrder submits a single-leg order for execution. Order placement is
// asynchronous — the response confirms submission, not execution. Use
// [Client.GetOrder] to poll for status. The OrderID field serves as a
// deduplication key for idempotency. Unmarshal the result into
// [PlaceOrderResponse].
func (c *Client) PlaceOrder(ctx context.Context, order OrderRequest) (json.RawMessage, error) {
	path := fmt.Sprintf("/userapigateway/trading/%s/order", c.accountID)
	return c.authedPost(ctx, path, order)
}

// PlaceMultiLegOrder submits a multi-leg order for execution. Like
// [Client.PlaceOrder], placement is asynchronous. Unmarshal the result into
// [PlaceOrderResponse].
func (c *Client) PlaceMultiLegOrder(ctx context.Context, order MultiLegOrderRequest) (json.RawMessage, error) {
	path := fmt.Sprintf("/userapigateway/trading/%s/order/multileg", c.accountID)
	return c.authedPost(ctx, path, order)
}

// ReplaceOrder replaces an existing order with new parameters. Replacement is
// asynchronous. The [ReplaceOrderRequest.OrderID] identifies the order to
// replace, and [ReplaceOrderRequest.RequestID] is the new order's UUID.
// Unmarshal the result into [PlaceOrderResponse].
func (c *Client) ReplaceOrder(ctx context.Context, order ReplaceOrderRequest) (json.RawMessage, error) {
	path := fmt.Sprintf("/userapigateway/trading/%s/order", c.accountID)
	return c.authedPut(ctx, path, order)
}

// GetOrder retrieves the current status of an order. The order may not be
// immediately available after placement due to asynchronous processing.
// Unmarshal the result into [Order].
func (c *Client) GetOrder(ctx context.Context, orderID string) (json.RawMessage, error) {
	return c.authedGet(ctx, fmt.Sprintf("/userapigateway/trading/%s/order/%s", c.accountID, orderID))
}

// CancelOrder cancels a pending order.
func (c *Client) CancelOrder(ctx context.Context, orderID string) error {
	return c.authedDelete(ctx, fmt.Sprintf("/userapigateway/trading/%s/order/%s", c.accountID, orderID))
}

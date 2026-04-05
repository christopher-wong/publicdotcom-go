// Package publicdotcom is an unofficial Go client for the Public.com trading API.
//
// It covers authentication, account management, portfolio data, market quotes,
// order placement, and options data. All endpoints return [json.RawMessage];
// typed response structs are provided for unmarshalling.
//
// This package is not affiliated with, maintained by, or endorsed by Public Holdings, Inc.
//
// # Authentication
//
// The client authenticates using a personal secret key, which is exchanged for
// a short-lived access token via [Client.Authenticate]. Authentication is called
// automatically before each request, and tokens are cached until 5 minutes
// before expiry.
//
//	client := publicdotcom.NewClient(secretKey, accountID)
//	// Authenticate is called automatically, or explicitly:
//	client.Authenticate(ctx)
//
// # Placing orders
//
// Orders are placed asynchronously. [Client.PlaceOrder] confirms submission, not
// execution. Use [Client.GetOrder] to poll for status. Use [Client.PreflightOrder]
// to validate an order before submitting.
//
//	resp, err := client.PlaceOrder(ctx, publicdotcom.OrderRequest{
//	    OrderID:    "uuid-here",
//	    Instrument: publicdotcom.Instrument{Symbol: "AAPL", Type: "EQUITY"},
//	    OrderSide:  "BUY",
//	    OrderType:  "LIMIT",
//	    Quantity:   "1",
//	    LimitPrice: "100.00",
//	    Expiration: publicdotcom.OrderExpiration{TimeInForce: "DAY"},
//	})
//
// # Error handling
//
// API errors are returned as [*APIError], which includes the HTTP status code and
// response body. Use errors.As to inspect them:
//
//	var apiErr *publicdotcom.APIError
//	if errors.As(err, &apiErr) {
//	    log.Printf("status %d: %s", apiErr.StatusCode, apiErr.Body)
//	}
//
// # References
//
//   - API docs: https://public.com/api/docs
//   - Postman collection: https://github.com/PublicDotCom/postman-collection
package publicdotcom

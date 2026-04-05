# publicdotcom-go

[![CI](https://github.com/christopher-wong/publicdotcom-go/actions/workflows/ci.yml/badge.svg)](https://github.com/christopher-wong/publicdotcom-go/actions/workflows/ci.yml)

Unofficial Go client for the [Public.com](https://public.com) trading API. Zero dependencies beyond the Go standard library.

> **This library is not affiliated with, maintained by, or endorsed by Public Holdings, Inc.** It interacts with live brokerage accounts capable of real trades. Use at your own risk.

## Install

```
go get github.com/christopher-wong/publicdotcom-go
```

Requires Go 1.25+.

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    publicdotcom "github.com/christopher-wong/publicdotcom-go"
)

func main() {
    client := publicdotcom.NewClient(
        os.Getenv("PUBLIC_SECRET_KEY"),
        os.Getenv("PUBLIC_ACCOUNT_ID"),
        publicdotcom.WithRetry(nil), // enable retries with default policy
    )

    ctx := context.Background()

    // Get portfolio — returns *PortfolioResponse directly
    portfolio, err := client.GetPortfolio(ctx)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Buying power: %s\n", portfolio.BuyingPower.BuyingPower)
    for _, p := range portfolio.Positions {
        fmt.Printf("  %s: %s shares\n", p.Instrument.Symbol, p.Quantity)
    }
}
```

A runnable example is in [`example/main.go`](example/main.go).

## Authentication

The client uses Public.com's personal access token flow. You need a **secret key** (generated from your Public.com account settings) and your **account ID**.

```go
client := publicdotcom.NewClient(secretKey, accountID)
```

Authentication is handled automatically before each request. Tokens are cached and refreshed 5 minutes before expiry. You can also call `client.Authenticate(ctx)` explicitly to fail fast on bad credentials.

## Client Options

```go
// Enable automatic retries (3 retries, exponential backoff, respects Retry-After)
client := publicdotcom.NewClient(key, id, publicdotcom.WithRetry(nil))

// Custom retry policy
client := publicdotcom.NewClient(key, id, publicdotcom.WithRetry(&publicdotcom.RetryPolicy{
    MaxRetries: 5,
    BaseDelay:  1 * time.Second,
    MaxDelay:   60 * time.Second,
}))

// Custom HTTP client
client := publicdotcom.NewClient(key, id, publicdotcom.WithHTTPClient(myHTTPClient))
```

All requests include a `User-Agent: publicdotcom-go/0.1.0` header. Plaintext HTTP is blocked at the transport level — only HTTPS requests are allowed.

## API Coverage

Based on the [official Postman collection](https://github.com/PublicDotCom/postman-collection) and [API docs](https://public.com/api/docs):

### Accounts & Portfolio

| Method | Description | Response Type |
|---|---|---|
| `GetAccounts` | List all accounts (brokerage, IRA, etc.) | `AccountsResponse` |
| `GetPortfolio` | Positions, orders, buying power, equity | `PortfolioResponse` |
| `GetHistory` | Paginated transaction history | `HistoryResponse` |

### Instruments

| Method | Description | Response Type |
|---|---|---|
| `GetInstruments` | All tradeable instruments with permissions | `InstrumentsResponse` |
| `GetInstrument` | Single instrument by symbol and type | `InstrumentDetail` |

### Market Data

| Method | Description | Response Type |
|---|---|---|
| `GetQuotes` | Bid/ask/last for multiple instruments | `QuoteResponse` |
| `GetOptionExpirations` | Available expiration dates | `OptionExpirationsResponse` |
| `GetOptionChain` | Calls and puts at a given expiration | `OptionChainResponse` |
| `GetOptionGreeks` | Delta, gamma, theta, vega, rho, IV | `GreeksResponse` |

### Orders

| Method | Description | Response Type |
|---|---|---|
| `PreflightOrder` | Validate order, get cost/fee/margin estimates | `PreflightResponse` |
| `PlaceOrder` | Submit a single-leg order | `PlaceOrderResponse` |
| `ReplaceOrder` | Atomically cancel and replace an order | `PlaceOrderResponse` |
| `GetOrder` | Get order status and fill details | `Order` |
| `CancelOrder` | Cancel a pending order | — |
| `PreflightMultiLegOrder` | Validate a multi-leg options order | `PreflightMultiLegResponse` |
| `PlaceMultiLegOrder` | Submit a multi-leg options order | `PlaceOrderResponse` |

All methods return typed response structs directly (e.g. `*PortfolioResponse`, `*Order`). No manual unmarshalling needed.

## Error Handling

API errors are returned as typed errors. Use `errors.As` to inspect them:

```go
import "errors"

resp, err := client.PlaceOrder(ctx, order)
if err != nil {
    var rateLimited *publicdotcom.RateLimitError
    var authErr *publicdotcom.AuthenticationError
    var validationErr *publicdotcom.ValidationError
    var notFound *publicdotcom.NotFoundError
    var serverErr *publicdotcom.ServerError

    switch {
    case errors.As(err, &rateLimited):
        fmt.Printf("rate limited, retry after %s\n", rateLimited.RetryAfter)
    case errors.As(err, &authErr):
        fmt.Println("bad credentials")
    case errors.As(err, &validationErr):
        fmt.Printf("invalid request: %s\n", validationErr.Body)
    case errors.As(err, &notFound):
        fmt.Println("resource not found")
    case errors.As(err, &serverErr):
        fmt.Println("server error, try again later")
    default:
        fmt.Println(err)
    }
}
```

All typed errors embed `*APIError` and can be unwrapped to it. The `APIError` includes the HTTP status code, response body, and API error code (if present).

You can check if an error is retryable:

```go
if publicdotcom.IsRetryable(err) {
    // safe to retry (rate limit or server error)
}
```

## Automatic Retries

Enable with `WithRetry`:

```go
client := publicdotcom.NewClient(key, id, publicdotcom.WithRetry(nil))
```

The default policy retries up to 3 times with exponential backoff (500ms base, 30s max) and jitter. Only 429 (rate limit) and 5xx (server error) responses are retried. The `Retry-After` header is respected on 429 responses.

Retries are safe for order placement because Public.com uses the `orderId` field as an idempotency key.

## Subscriptions

### Price Subscription

Poll for quote changes and get notified when prices move:

```go
instruments := []publicdotcom.Instrument{
    {Symbol: "AAPL", Type: "EQUITY"},
    {Symbol: "MSFT", Type: "EQUITY"},
}

sub := client.SubscribePrices(ctx, instruments,
    publicdotcom.WithPriceInterval(2*time.Second),
)
defer sub.Close()

for update := range sub.C() {
    if update.Err != nil {
        log.Println("error:", update.Err)
        continue
    }
    fmt.Printf("%s: %s\n", update.Instrument.Symbol, update.Quote.Last)
}
```

Updates are only delivered when a price actually changes. The subscription runs until `Close()` is called or the context is cancelled.

### Order Subscription

Track an order through its lifecycle:

```go
sub := client.SubscribeOrder(ctx, orderID,
    publicdotcom.WithOrderInterval(1*time.Second),
)
defer sub.Close()

for update := range sub.C() {
    if update.Err != nil {
        log.Println("error:", update.Err)
        continue
    }
    fmt.Printf("order %s: %s -> %s\n",
        update.Order.OrderID, update.Previous, update.Order.Status)
}
// Channel closes automatically when the order reaches a terminal state
// (FILLED, CANCELLED, REJECTED, EXPIRED, REPLACED)
```

## Project Structure

```
publicdotcom-go/
  client.go        # Client, auth, transport, options
  accounts.go      # GetAccounts, GetPortfolio, GetHistory
  instruments.go   # GetInstruments, GetInstrument
  marketdata.go    # GetQuotes, GetOptionExpirations, GetOptionChain, GetOptionGreeks
  orders.go        # PlaceOrder, ReplaceOrder, GetOrder, CancelOrder, preflight, multi-leg
  types.go         # All request and response types
  errors.go        # Typed error hierarchy
  retry.go         # Retry transport and policy
  subscribe.go     # Price and order subscriptions
  doc.go           # Package-level godoc
  example/main.go  # Runnable example
```

## References

- [Public.com API Docs](https://public.com/api/docs)
- [Public.com API Changelog](https://public.com/api/docs/changelog)
- [Official Postman Collection](https://github.com/PublicDotCom/postman-collection)
- [Official Python SDK](https://github.com/PublicDotCom/publicdotcom-py)

## Disclaimer

This is an unofficial, community-maintained project. It is not affiliated with, endorsed by, or supported by Public Holdings, Inc. This library interacts with live brokerage accounts and can execute real trades. Trading involves risk of financial loss. Always review orders before submission.

# publicdotcom-go

Unofficial Go client for the [Public.com](https://public.com) trading API.

This library is **not affiliated with, maintained by, or endorsed by Public Holdings, Inc.**

## Install

```
go get github.com/christopher-wong/publicdotcom-go
```

## Usage

```go
client := publicdotcom.NewClient(secretKey, accountID)

// Authentication is handled automatically, or call it explicitly:
client.Authenticate(ctx)

// Get portfolio
portfolio, err := client.GetPortfolio(ctx)

// Get quotes
quotes, err := client.GetQuotes(ctx, []publicdotcom.Instrument{
    {Symbol: "AAPL", Type: "EQUITY"},
})

// Place an order
resp, err := client.PlaceOrder(ctx, publicdotcom.OrderRequest{
    OrderID:    "uuid-here",
    Instrument: publicdotcom.Instrument{Symbol: "AAPL", Type: "EQUITY"},
    OrderSide:  "BUY",
    OrderType:  "LIMIT",
    Quantity:   "1",
    LimitPrice: "100.00",
    Expiration: publicdotcom.OrderExpiration{TimeInForce: "DAY"},
})
```

All responses are returned as `json.RawMessage` since the API does not publish response schemas.

## API Coverage

Based on the [official Postman collection](https://github.com/PublicDotCom/postman-collection):

| Category | Methods |
|---|---|
| Auth | `Authenticate` |
| Accounts | `GetAccounts` |
| Portfolio | `GetPortfolio` |
| History | `GetHistory` |
| Instruments | `GetInstruments`, `GetInstrument` |
| Market Data | `GetQuotes`, `GetOptionExpirations`, `GetOptionChain` |
| Orders | `PreflightOrder`, `PlaceOrder`, `GetOrder`, `CancelOrder` |
| Multi-leg Orders | `PreflightMultiLegOrder`, `PlaceMultiLegOrder` |
| Options | `GetOptionGreeks` |

## References

- [Public.com API Docs](https://public.com/api/docs)
- [Postman Collection](https://github.com/PublicDotCom/postman-collection)

## Disclaimer

This is an unofficial, community-maintained project. Use at your own risk. Trading involves risk of financial loss. This library interacts with live brokerage accounts.

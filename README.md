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
raw, err := client.GetPortfolio(ctx)

// Unmarshal into typed response
var portfolio publicdotcom.PortfolioResponse
json.Unmarshal(raw, &portfolio)

// Get quotes
raw, err = client.GetQuotes(ctx, []publicdotcom.Instrument{
    {Symbol: "AAPL", Type: "EQUITY"},
})
var quotes publicdotcom.QuoteResponse
json.Unmarshal(raw, &quotes)

// Place an order
raw, err = client.PlaceOrder(ctx, publicdotcom.OrderRequest{
    OrderID:    "uuid-here",
    Instrument: publicdotcom.Instrument{Symbol: "AAPL", Type: "EQUITY"},
    OrderSide:  "BUY",
    OrderType:  "LIMIT",
    Quantity:   "1",
    LimitPrice: "100.00",
    Expiration: publicdotcom.OrderExpiration{TimeInForce: "DAY"},
})
```

Responses are returned as `json.RawMessage`. Typed response structs are provided for all endpoints — unmarshal into them as shown above.

## API Coverage

Based on the [official Postman collection](https://github.com/PublicDotCom/postman-collection) and [API docs](https://public.com/api/docs):

| Category | Methods | Response Type |
|---|---|---|
| Auth | `Authenticate` | — |
| Accounts | `GetAccounts` | `AccountsResponse` |
| Portfolio | `GetPortfolio` | `PortfolioResponse` |
| History | `GetHistory` | `HistoryResponse` |
| Instruments | `GetInstruments`, `GetInstrument` | `InstrumentsResponse`, `InstrumentDetail` |
| Market Data | `GetQuotes`, `GetOptionExpirations`, `GetOptionChain` | `QuoteResponse`, `OptionExpirationsResponse`, `OptionChainResponse` |
| Orders | `PreflightOrder`, `PlaceOrder`, `ReplaceOrder`, `GetOrder`, `CancelOrder` | `PreflightResponse`, `PlaceOrderResponse`, `Order` |
| Multi-leg Orders | `PreflightMultiLegOrder`, `PlaceMultiLegOrder` | `PreflightMultiLegResponse`, `PlaceOrderResponse` |
| Options | `GetOptionGreeks` | `GreeksResponse` |

## References

- [Public.com API Docs](https://public.com/api/docs)
- [Postman Collection](https://github.com/PublicDotCom/postman-collection)

## Disclaimer

This is an unofficial, community-maintained project. Use at your own risk. Trading involves risk of financial loss. This library interacts with live brokerage accounts.

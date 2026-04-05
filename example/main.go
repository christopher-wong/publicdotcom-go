package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	publicdotcom "github.com/cwong/publicdotcom-go"
)

func main() {
	secretKey := os.Getenv("PUBLIC_SECRET_KEY")
	accountID := os.Getenv("PUBLIC_ACCOUNT_ID")
	if secretKey == "" || accountID == "" {
		log.Fatal("PUBLIC_SECRET_KEY and PUBLIC_ACCOUNT_ID must be set")
	}

	ctx := context.Background()
	client := publicdotcom.NewClient(secretKey, accountID, publicdotcom.WithRetry(nil))

	// Authenticate
	if err := client.Authenticate(ctx); err != nil {
		log.Fatalf("auth: %v", err)
	}
	fmt.Println("Authenticated.")

	// Get accounts
	raw, err := client.GetAccounts(ctx)
	if err != nil {
		log.Fatalf("get accounts: %v", err)
	}
	var accounts publicdotcom.AccountsResponse
	json.Unmarshal(raw, &accounts)
	for _, a := range accounts.Accounts {
		fmt.Printf("Account: %s (%s)\n", a.AccountID, a.AccountType)
	}

	// Get portfolio
	raw, err = client.GetPortfolio(ctx)
	if err != nil {
		log.Fatalf("get portfolio: %v", err)
	}
	var portfolio publicdotcom.PortfolioResponse
	json.Unmarshal(raw, &portfolio)
	fmt.Printf("Buying power: %s\n", portfolio.BuyingPower.BuyingPower)
	for _, p := range portfolio.Positions {
		fmt.Printf("  %s: %s shares @ %s\n", p.Instrument.Symbol, p.Quantity, p.LastPrice.LastPrice)
	}

	// Get a quote
	raw, err = client.GetQuotes(ctx, []publicdotcom.Instrument{
		{Symbol: "AAPL", Type: "EQUITY"},
	})
	if err != nil {
		log.Fatalf("get quotes: %v", err)
	}
	var quotes publicdotcom.QuoteResponse
	json.Unmarshal(raw, &quotes)
	for _, q := range quotes.Quotes {
		fmt.Printf("%s: last=%s bid=%s ask=%s\n", q.Instrument.Symbol, q.Last, q.Bid, q.Ask)
	}
}

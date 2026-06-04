package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	publicdotcom "github.com/christopher-wong/publicdotcom-go"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "usage: %s <secret-key> <account-id>\n", os.Args[0])
		os.Exit(1)
	}
	secretKey := os.Args[1]
	accountID := os.Args[2]

	log.Printf("starting Public.com example")
	log.Printf("using account id: %s", accountID)
	log.Printf("creating client with retry policy enabled")

	ctx := context.Background()
	client := publicdotcom.NewClient(secretKey, accountID, publicdotcom.WithRetry(nil))

	// --- Authenticate ---
	log.Printf("authenticating")
	start := time.Now()
	if err := client.Authenticate(ctx); err != nil {
		log.Fatalf("auth: %v", err)
	}
	log.Printf("authenticated successfully in %s", time.Since(start))
	fmt.Println("Authenticated.")
	fmt.Println()

	// --- Account info ---
	log.Printf("fetching accounts")
	start = time.Now()
	accounts, err := client.GetAccounts(ctx)
	if err != nil {
		log.Fatalf("get accounts: %v", err)
	}
	log.Printf("fetched %d account(s) in %s", len(accounts.Accounts), time.Since(start))
	for _, a := range accounts.Accounts {
		fmt.Printf("Account: %s  type=%s  options=%s  permissions=%s\n",
			a.AccountID, a.AccountType, a.OptionsLevel, a.TradePermissions)
	}

	// --- Portfolio: balance + positions ---
	fmt.Println("\n--- Portfolio ---")
	log.Printf("fetching portfolio for account %s", accountID)
	start = time.Now()
	portfolio, err := client.GetPortfolio(ctx)
	if err != nil {
		log.Fatalf("get portfolio: %v", err)
	}
	log.Printf("fetched portfolio in %s: %d equity item(s), %d position(s), %d open order(s)",
		time.Since(start), len(portfolio.Equity), len(portfolio.Positions), len(portfolio.Orders))

	fmt.Printf("Buying Power:      %s\n", portfolio.BuyingPower.BuyingPower)
	fmt.Printf("Cash Buying Power: %s\n", portfolio.BuyingPower.CashOnlyBuyingPower)

	for _, eq := range portfolio.Equity {
		fmt.Printf("  %-20s %s\n", eq.Label+":", eq.Value)
	}

	if len(portfolio.Positions) == 0 {
		fmt.Println("\nNo open positions.")
	} else {
		fmt.Printf("\nPositions (%d):\n", len(portfolio.Positions))
		fmt.Printf("  %-8s %10s %12s %10s %12s %10s\n",
			"Symbol", "Qty", "Value", "Last", "Cost Basis", "P&L")
		fmt.Println("  ------   ---------- ------------ ---------- ------------ ----------")
		for _, p := range portfolio.Positions {
			fmt.Printf("  %-8s %10s %12s %10s %12s %10s\n",
				p.Instrument.Symbol,
				p.Quantity,
				p.CurrentValue,
				p.LastPrice.LastPrice,
				p.CostBasis.TotalCost,
				p.CostBasis.GainValue)
		}
	}

	if len(portfolio.Orders) > 0 {
		fmt.Printf("\nOpen Orders (%d):\n", len(portfolio.Orders))
		for _, o := range portfolio.Orders {
			fmt.Printf("  %s %s %s %s qty=%s status=%s\n",
				o.OrderID, o.Side, o.Type, o.Instrument.Symbol, o.Quantity, o.Status)
		}
	}

	// --- RKLB quote ---
	fmt.Println("\n--- RKLB Market Data ---")
	log.Printf("fetching RKLB quote")
	start = time.Now()
	quotes, err := client.GetQuotes(ctx, []publicdotcom.Instrument{
		{Symbol: "RKLB", Type: "EQUITY"},
	})
	if err != nil {
		log.Fatalf("get quotes: %v", err)
	}
	log.Printf("fetched %d quote(s) in %s", len(quotes.Quotes), time.Since(start))
	for _, q := range quotes.Quotes {
		fmt.Printf("Symbol:         %s\n", q.Instrument.Symbol)
		fmt.Printf("Last:           %s\n", q.Last)
		fmt.Printf("Bid:            %s (size %d)\n", q.Bid, q.BidSize)
		fmt.Printf("Ask:            %s (size %d)\n", q.Ask, q.AskSize)
		fmt.Printf("Volume:         %d\n", q.Volume)
		fmt.Printf("Prev Close:     %s\n", q.PreviousClose)
		fmt.Printf("Day Change:     %s (%s%%)\n", q.OneDayChange.Change, q.OneDayChange.PercentChange)
	}

	// --- RKLB instrument details ---
	log.Printf("fetching RKLB instrument details")
	start = time.Now()
	instrument, err := client.GetInstrument(ctx, "RKLB", "EQUITY")
	if err != nil {
		log.Fatalf("get instrument: %v", err)
	}
	log.Printf("fetched RKLB instrument details in %s", time.Since(start))
	fmt.Printf("\nInstrument:     %s (%s)\n", instrument.Instrument.Symbol, instrument.Instrument.Type)
	fmt.Printf("Trading:        %s\n", instrument.Trading)
	fmt.Printf("Fractional:     %s\n", instrument.FractionalTrading)
	fmt.Printf("Options:        %s\n", instrument.OptionTrading)
	if instrument.ShortingAvailability != "" {
		fmt.Printf("Shorting:       %s\n", instrument.ShortingAvailability)
	}

	// --- RKLB option expirations ---
	log.Printf("fetching RKLB option expirations")
	start = time.Now()
	expirations, err := client.GetOptionExpirations(ctx, publicdotcom.Instrument{
		Symbol: "RKLB", Type: "EQUITY",
	})
	if err != nil {
		log.Fatalf("get option expirations: %v", err)
	}
	log.Printf("fetched %d RKLB option expiration(s) in %s", len(expirations.Expirations), time.Since(start))
	fmt.Printf("\nOption Expirations (%d):", len(expirations.Expirations))
	for i, exp := range expirations.Expirations {
		if i >= 5 {
			fmt.Printf(" ... and %d more", len(expirations.Expirations)-5)
			break
		}
		fmt.Printf(" %s", exp)
	}
	fmt.Println()
	log.Printf("example completed successfully")
}

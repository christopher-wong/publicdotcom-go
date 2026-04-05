package publicdotcom

import "time"

// --- Shared ---

// Instrument identifies a tradeable instrument by symbol and type.
// Type is one of: EQUITY, OPTION, CRYPTO, ALT, TREASURY, BOND, INDEX,
// or MULTI_LEG_INSTRUMENT.
type Instrument struct {
	Symbol string `json:"symbol"`
	Type   string `json:"type"`
}

// --- Request types ---

// OrderRequest is the body for placing or preflighting a single-leg order.
// Quantity and Amount are mutually exclusive: use Quantity for share-based
// orders and Amount for notional (dollar) orders.
type OrderRequest struct {
	OrderID             string          `json:"orderId,omitempty"`
	Instrument          Instrument      `json:"instrument"`
	OrderSide           string          `json:"orderSide"`                     // BUY, SELL
	OrderType           string          `json:"orderType"`                     // MARKET, LIMIT, STOP, STOP_LIMIT
	Quantity            string          `json:"quantity,omitempty"`
	Amount              string          `json:"amount,omitempty"`
	LimitPrice          string          `json:"limitPrice,omitempty"`
	StopPrice           string          `json:"stopPrice,omitempty"`
	Expiration          OrderExpiration `json:"expiration"`
	EquityMarketSession string          `json:"equityMarketSession,omitempty"` // CORE, EXTENDED
	OpenCloseIndicator  string          `json:"openCloseIndicator,omitempty"`  // OPEN, CLOSE
}

// OrderExpiration controls time-in-force for an order.
// Use TimeInForce "DAY" for day orders, or "GTD" with an ExpirationTime
// for good-til-date orders.
type OrderExpiration struct {
	TimeInForce    string `json:"timeInForce"`
	ExpirationTime string `json:"expirationTime,omitempty"`
}

// MultiLegOrderRequest is the body for placing or preflighting a multi-leg
// (options spread) order. Legs define the individual option contracts and
// their sides.
type MultiLegOrderRequest struct {
	OrderID    string          `json:"orderId,omitempty"`
	OrderType  string          `json:"type"`
	LimitPrice string          `json:"limitPrice,omitempty"`
	Expiration OrderExpiration `json:"expiration"`
	Quantity   int             `json:"quantity"`
	Legs       []OrderLeg      `json:"legs"`
}

// OrderLeg is a single leg within a multi-leg order.
type OrderLeg struct {
	Instrument         Instrument `json:"instrument"`
	Side               string     `json:"side"`               // BUY, SELL
	OpenCloseIndicator string     `json:"openCloseIndicator"` // OPEN, CLOSE
	RatioQuantity      int        `json:"ratioQuantity"`
}

// ReplaceOrderRequest is the body for replacing an existing order.
// OrderID identifies the order to replace; RequestID is the UUID for the
// new replacement order and serves as a deduplication key.
type ReplaceOrderRequest struct {
	OrderID    string          `json:"orderId"`
	RequestID  string          `json:"requestId"`
	OrderType  string          `json:"orderType"` // MARKET, LIMIT, STOP, STOP_LIMIT
	Expiration OrderExpiration `json:"expiration"`
	Quantity   string          `json:"quantity,omitempty"`
	LimitPrice string          `json:"limitPrice,omitempty"`
	StopPrice  string          `json:"stopPrice,omitempty"`
}

// QuoteRequest is the body for requesting market quotes.
type QuoteRequest struct {
	Instruments []Instrument `json:"instruments"`
}

// OptionChainRequest is the body for requesting an option chain at a
// specific expiration date (formatted as YYYY-MM-DD).
type OptionChainRequest struct {
	Instrument     Instrument `json:"instrument"`
	ExpirationDate string     `json:"expirationDate"`
}

// OptionExpirationsRequest is the body for requesting available option
// expiration dates for an instrument.
type OptionExpirationsRequest struct {
	Instrument Instrument `json:"instrument"`
}

// HistoryParams are optional query parameters for [Client.GetHistory].
// All fields are optional; pass nil to use server defaults.
type HistoryParams struct {
	Start     string // ISO 8601 datetime with timezone
	End       string // ISO 8601 datetime with timezone
	PageSize  int
	NextToken string // pagination token from a previous [HistoryResponse]
}

// --- Response types ---

// AccountsResponse is the response from [Client.GetAccounts].
type AccountsResponse struct {
	Accounts []Account `json:"accounts"`
}

// Account represents a financial account associated with the authenticated user.
type Account struct {
	AccountID            string `json:"accountId"`
	AccountType          string `json:"accountType"`          // BROKERAGE, HIGH_YIELD, BOND_ACCOUNT, RIA_ASSET, TREASURY, TRADITIONAL_IRA, ROTH_IRA
	OptionsLevel         string `json:"optionsLevel"`         // NONE, LEVEL_1, LEVEL_2, LEVEL_3, LEVEL_4
	BrokerageAccountType string `json:"brokerageAccountType"` // CASH, MARGIN
	TradePermissions     string `json:"tradePermissions"`     // BUY_AND_SELL, RESTRICTED_SETTLED_FUNDS_ONLY, RESTRICTED_CLOSE_ONLY, RESTRICTED_NO_TRADING
}

// PortfolioResponse is the response from [Client.GetPortfolio].
// It contains positions, open orders, buying power, and equity summaries.
type PortfolioResponse struct {
	AccountID   string       `json:"accountId"`
	AccountType string       `json:"accountType"`
	BuyingPower BuyingPower  `json:"buyingPower"`
	Equity      []EquityItem `json:"equity"`
	Positions   []Position   `json:"positions"`
	Orders      []Order      `json:"orders"`
}

// BuyingPower represents available buying power for the account.
type BuyingPower struct {
	CashOnlyBuyingPower string `json:"cashOnlyBuyingPower"`
	BuyingPower         string `json:"buyingPower"`
	OptionsBuyingPower  string `json:"optionsBuyingPower"`
}

// EquityItem is a labeled equity summary value (e.g. "Total Value", "Day P&L").
type EquityItem struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// Position represents a portfolio position with current value, cost basis,
// and gain/loss information.
type Position struct {
	Instrument         PositionInstrument `json:"instrument"`
	Quantity           string             `json:"quantity"`
	OpenedAt           time.Time          `json:"openedAt"`
	CurrentValue       string             `json:"currentValue"`
	PercentOfPortfolio string             `json:"percentOfPortfolio"`
	LastPrice          PricePoint         `json:"lastPrice"`
	InstrumentGain     GainInfo           `json:"instrumentGain"`
	PositionDailyGain  GainInfo           `json:"positionDailyGain"`
	CostBasis          CostBasis          `json:"costBasis"`
}

// PositionInstrument extends [Instrument] with a human-readable name.
type PositionInstrument struct {
	Symbol string `json:"symbol"`
	Name   string `json:"name"`
	Type   string `json:"type"`
}

// PricePoint is a price observation at a point in time.
type PricePoint struct {
	LastPrice string    `json:"lastPrice"`
	Timestamp time.Time `json:"timestamp"`
}

// GainInfo represents a gain or loss as both a dollar value and percentage.
type GainInfo struct {
	GainValue      string    `json:"gainValue"`
	GainPercentage string    `json:"gainPercentage"`
	Timestamp      time.Time `json:"timestamp"`
}

// CostBasis represents the cost basis and unrealized gain/loss for a position.
type CostBasis struct {
	TotalCost      string `json:"totalCost"`
	UnitCost       string `json:"unitCost"`
	GainValue      string `json:"gainValue"`
	GainPercentage string `json:"gainPercentage"`
}

// Order represents an order, returned by [Client.GetOrder] and within
// [PortfolioResponse]. Status is one of: NEW, PARTIALLY_FILLED, FILLED,
// CANCELLED, QUEUED_CANCELLED, REJECTED, PENDING_REPLACE, PENDING_CANCEL,
// EXPIRED, or REPLACED.
type Order struct {
	OrderID            string          `json:"orderId"`
	Instrument         Instrument      `json:"instrument"`
	CreatedAt          time.Time       `json:"createdAt"`
	Type               string          `json:"type"`
	Side               string          `json:"side"`
	Status             string          `json:"status"`
	Quantity           string          `json:"quantity,omitempty"`
	NotionalValue      string          `json:"notionalValue,omitempty"`
	Expiration         OrderExpiration `json:"expiration"`
	LimitPrice         string          `json:"limitPrice,omitempty"`
	StopPrice          string          `json:"stopPrice,omitempty"`
	ClosedAt           *time.Time      `json:"closedAt,omitempty"`
	OpenCloseIndicator string          `json:"openCloseIndicator,omitempty"`
	FilledQuantity     string          `json:"filledQuantity,omitempty"`
	AveragePrice       string          `json:"averagePrice,omitempty"`
	Legs               []OrderLeg      `json:"legs,omitempty"`
	RejectReason       string          `json:"rejectReason,omitempty"`
}

// PlaceOrderResponse is returned by [Client.PlaceOrder],
// [Client.PlaceMultiLegOrder], and [Client.ReplaceOrder]. It confirms
// submission only — use [Client.GetOrder] to check execution status.
type PlaceOrderResponse struct {
	OrderID string `json:"orderId"`
}

// HistoryResponse is the response from [Client.GetHistory].
// Use NextToken with a subsequent request to fetch the next page.
type HistoryResponse struct {
	Transactions []Transaction `json:"transactions"`
	NextToken    string        `json:"nextToken,omitempty"`
	Start        time.Time     `json:"start"`
	End          time.Time     `json:"end"`
	PageSize     int           `json:"pageSize"`
}

// Transaction represents a single account transaction (trade, deposit,
// dividend, etc.).
type Transaction struct {
	Timestamp       time.Time `json:"timestamp"`
	ID              string    `json:"id"`
	Type            string    `json:"type"`    // TRADE, MONEY_MOVEMENT, POSITION_ADJUSTMENT
	SubType         string    `json:"subType"` // DEPOSIT, WITHDRAWAL, DIVIDEND, FEE, REWARD, INTEREST, TRADE, TRANSFER, MISC
	Symbol          string    `json:"symbol,omitempty"`
	SecurityType    string    `json:"securityType,omitempty"` // EQUITY, OPTION, CRYPTO, ALT, TREASURY, BOND
	Side            string    `json:"side,omitempty"`         // BUY, SELL
	Direction       string    `json:"direction,omitempty"`    // INCOMING, OUTGOING
	NetAmount       string    `json:"netAmount,omitempty"`
	PrincipalAmount string    `json:"principalAmount,omitempty"`
	Quantity        string    `json:"quantity,omitempty"`
	Fees            string    `json:"fees,omitempty"`
	Description     string    `json:"description,omitempty"`
}

// QuoteResponse is the response from [Client.GetQuotes].
type QuoteResponse struct {
	Quotes []Quote `json:"quotes"`
}

// Quote represents a market quote for a single instrument, including
// bid/ask, last price, volume, and daily change.
type Quote struct {
	Instrument    Instrument          `json:"instrument"`
	Outcome       string              `json:"outcome"` // SUCCESS, UNKNOWN
	Last          string              `json:"last"`
	LastTimestamp  time.Time           `json:"lastTimestamp"`
	Bid           string              `json:"bid"`
	BidSize       int                 `json:"bidSize"`
	BidTimestamp  time.Time           `json:"bidTimestamp"`
	Ask           string              `json:"ask"`
	AskSize       int                 `json:"askSize"`
	AskTimestamp  time.Time           `json:"askTimestamp"`
	Volume        int64               `json:"volume"`
	OpenInterest  int64               `json:"openInterest,omitempty"`
	PreviousClose string              `json:"previousClose"`
	OneDayChange  OneDayChange        `json:"oneDayChange"`
	OptionDetails *QuoteOptionDetails `json:"optionDetails,omitempty"`
}

// OneDayChange represents the daily price change in both dollar and
// percentage terms.
type OneDayChange struct {
	Change        string `json:"change"`
	PercentChange string `json:"percentChange"`
}

// QuoteOptionDetails contains option-specific fields present on option quotes.
type QuoteOptionDetails struct {
	StrikePrice string        `json:"strikePrice"`
	MidPrice    string        `json:"midPrice,omitempty"`
	Greeks      *OptionGreeks `json:"greeks,omitempty"`
}

// InstrumentsResponse is the response from [Client.GetInstruments].
type InstrumentsResponse struct {
	Instruments []InstrumentDetail `json:"instruments"`
}

// InstrumentDetail describes a tradeable instrument, its trading permissions,
// shorting availability, and optional asset-class-specific details.
type InstrumentDetail struct {
	Instrument                 Instrument              `json:"instrument"`
	Trading                    string                  `json:"trading"`              // BUY_AND_SELL, LIQUIDATION_ONLY, DISABLED
	FractionalTrading          string                  `json:"fractionalTrading"`    // BUY_AND_SELL, LIQUIDATION_ONLY, DISABLED
	OptionTrading              string                  `json:"optionTrading"`        // BUY_AND_SELL, LIQUIDATION_ONLY, DISABLED
	OptionSpreadTrading        string                  `json:"optionSpreadTrading"`  // BUY_AND_SELL, LIQUIDATION_ONLY, DISABLED
	ShortingAvailability       string                  `json:"shortingAvailability,omitempty"` // NOT_SHORTABLE, EASY_TO_BORROW, HARD_TO_BORROW
	HardToBorrowPercentageRate string                  `json:"hardToBorrowPercentageRate,omitempty"`
	InstrumentDetails          *InstrumentExtraDetails `json:"instrumentDetails,omitempty"`
}

// InstrumentExtraDetails holds asset-class-specific details. For crypto
// instruments, the precision fields are populated. For bonds, HasOutstanding
// is populated. The API uses a oneOf schema; both are flattened here.
type InstrumentExtraDetails struct {
	CryptoQuantityPrecision int  `json:"cryptoQuantityPrecision,omitempty"`
	CryptoPricePrecision    int  `json:"cryptoPricePrecision,omitempty"`
	TradableInNewYork       bool `json:"tradableInNewYork,omitempty"`
	HasOutstanding          bool `json:"hasOutstanding,omitempty"`
}

// PreflightResponse is the response from [Client.PreflightOrder]. It contains
// estimated costs, fees, margin requirements, and short-selling availability
// for a proposed order.
type PreflightResponse struct {
	Instrument              Instrument              `json:"instrument"`
	Cusip                   string                  `json:"cusip,omitempty"`
	RootSymbol              string                  `json:"rootSymbol,omitempty"`
	RootOptionSymbol        string                  `json:"rootOptionSymbol,omitempty"`
	OrderValue              string                  `json:"orderValue"`
	EstimatedQuantity       string                  `json:"estimatedQuantity,omitempty"`
	EstimatedCost           string                  `json:"estimatedCost,omitempty"`
	EstimatedProceeds       string                  `json:"estimatedProceeds,omitempty"`
	EstimatedCommission     string                  `json:"estimatedCommission,omitempty"`
	EstimatedIndexOptionFee string                  `json:"estimatedIndexOptionFee,omitempty"`
	EstimatedExecutionFee   string                  `json:"estimatedExecutionFee,omitempty"`
	BuyingPowerRequirement  string                  `json:"buyingPowerRequirement,omitempty"`
	RegulatoryFees          *RegulatoryFees         `json:"regulatoryFees,omitempty"`
	OptionDetails           *PreflightOptionDetails `json:"optionDetails,omitempty"`
	EstimatedOrderRebate    *OptionRebate           `json:"estimatedOrderRebate,omitempty"`
	MarginRequirement       *MarginRequirement      `json:"marginRequirement,omitempty"`
	MarginImpact            *MarginImpact           `json:"marginImpact,omitempty"`
	ShortSelling            *ShortSelling           `json:"shortSelling,omitempty"`
	PriceIncrement          *PriceIncrement         `json:"priceIncrement,omitempty"`
}

// RegulatoryFees contains the breakdown of regulatory fees for an order.
type RegulatoryFees struct {
	SecFee      string `json:"secFee,omitempty"`
	TafFee      string `json:"tafFee,omitempty"`
	OrfFee      string `json:"orfFee,omitempty"`
	ExchangeFee string `json:"exchangeFee,omitempty"`
	OccFee      string `json:"occFee,omitempty"`
	CatFee      string `json:"catFee,omitempty"`
}

// PreflightOptionDetails contains option-specific details from a preflight response.
type PreflightOptionDetails struct {
	BaseSymbol       string `json:"baseSymbol"`
	Type             string `json:"type"` // CALL, PUT
	StrikePrice      string `json:"strikePrice"`
	OptionExpireDate string `json:"optionExpireDate"`
}

// OptionRebate contains estimated rebate information for option orders.
type OptionRebate struct {
	EstimatedOptionRebate string `json:"estimatedOptionRebate,omitempty"`
	PerContractRebate     string `json:"perContractRebate,omitempty"`
	OptionRebatePercent   int    `json:"optionRebatePercent,omitempty"`
}

// MarginRequirement contains the margin requirements for an order.
type MarginRequirement struct {
	LongMaintenanceRequirement string `json:"longMaintenanceRequirement,omitempty"`
	LongInitialRequirement     string `json:"longInitialRequirement,omitempty"`
}

// MarginImpact describes how an order would affect the account's margin usage.
type MarginImpact struct {
	MarginUsageImpact        string `json:"marginUsageImpact,omitempty"`
	InitialMarginRequirement string `json:"initialMarginRequirement,omitempty"`
}

// ShortSelling contains short-selling availability and margin details for
// an instrument.
type ShortSelling struct {
	Availability                           string `json:"availability,omitempty"` // NOT_SHORTABLE, EASY_TO_BORROW, HARD_TO_BORROW
	HardToBorrowPercentageRate             string `json:"hardToBorrowPercentageRate,omitempty"`
	InitialMarginRequirementPercentage     string `json:"initialMarginRequirementPercentage,omitempty"`
	MaintenanceMarginRequirementPercentage string `json:"maintenanceMarginRequirementPercentage,omitempty"`
	MaxQuantityForLocate                   int    `json:"maxQuantityForLocate,omitempty"`
	UptickRule                             string `json:"uptickRule,omitempty"` // TRIGGERED, NOT_TRIGGERED
}

// PriceIncrement contains tick-size information for an instrument.
type PriceIncrement struct {
	IncrementBelow3  string `json:"incrementBelow3,omitempty"`
	IncrementAbove3  string `json:"incrementAbove3,omitempty"`
	CurrentIncrement string `json:"currentIncrement,omitempty"`
}

// PreflightMultiLegResponse is the response from [Client.PreflightMultiLegOrder].
// It includes the detected strategy name, per-leg details, and aggregate cost/margin
// estimates.
type PreflightMultiLegResponse struct {
	BaseSymbol              string             `json:"baseSymbol,omitempty"`
	StrategyName            string             `json:"strategyName,omitempty"`
	Legs                    []PreflightLeg     `json:"legs,omitempty"`
	OrderValue              string             `json:"orderValue,omitempty"`
	EstimatedQuantity       string             `json:"estimatedQuantity,omitempty"`
	EstimatedCost           string             `json:"estimatedCost,omitempty"`
	EstimatedProceeds       string             `json:"estimatedProceeds,omitempty"`
	EstimatedCommission     string             `json:"estimatedCommission,omitempty"`
	EstimatedIndexOptionFee string             `json:"estimatedIndexOptionFee,omitempty"`
	BuyingPowerRequirement  string             `json:"buyingPowerRequirement,omitempty"`
	RegulatoryFees          *RegulatoryFees    `json:"regulatoryFees,omitempty"`
	MarginRequirement       *MarginRequirement `json:"marginRequirement,omitempty"`
	MarginImpact            *MarginImpact      `json:"marginImpact,omitempty"`
	PriceIncrement          *PriceIncrement    `json:"priceIncrement,omitempty"`
}

// PreflightLeg contains per-leg details from a multi-leg preflight,
// including the resolved option details.
type PreflightLeg struct {
	Instrument         Instrument              `json:"instrument"`
	Side               string                  `json:"side"`               // BUY, SELL
	OpenCloseIndicator string                  `json:"openCloseIndicator"` // OPEN, CLOSE
	RatioQuantity      int                     `json:"ratioQuantity"`
	OptionDetails      *PreflightOptionDetails `json:"optionDetails,omitempty"`
}

// OptionExpirationsResponse is the response from [Client.GetOptionExpirations].
type OptionExpirationsResponse struct {
	BaseSymbol  string   `json:"baseSymbol"`
	Expirations []string `json:"expirations"` // dates in YYYY-MM-DD format
}

// OptionChainResponse is the response from [Client.GetOptionChain].
// Calls and Puts each contain [Quote] entries with option-specific details.
type OptionChainResponse struct {
	BaseSymbol string  `json:"baseSymbol"`
	Calls      []Quote `json:"calls"`
	Puts       []Quote `json:"puts"`
}

// GreeksResponse is the response from [Client.GetOptionGreeks].
type GreeksResponse struct {
	Greeks []GreekEntry `json:"greeks"`
}

// GreekEntry contains the greeks for a single option, identified by its
// OSI-normalized symbol.
type GreekEntry struct {
	Symbol string       `json:"symbol"`
	Greeks OptionGreeks `json:"greeks"`
}

// OptionGreeks contains the greek values for an option contract.
// All values are returned as strings representing decimal numbers.
type OptionGreeks struct {
	Delta             string `json:"delta"`
	Gamma             string `json:"gamma"`
	Theta             string `json:"theta"`
	Vega              string `json:"vega"`
	Rho               string `json:"rho"`
	ImpliedVolatility string `json:"impliedVolatility"`
}

package publicdotcom

// Instrument identifies a tradeable instrument.
type Instrument struct {
	Symbol string `json:"symbol"`
	Type   string `json:"type"` // EQUITY, OPTION, etc.
}

// OrderRequest is the body for placing or preflighting a single-leg order.
type OrderRequest struct {
	OrderID             string          `json:"orderId,omitempty"`
	Instrument          Instrument      `json:"instrument"`
	OrderSide           string          `json:"orderSide"`                     // BUY, SELL
	OrderType           string          `json:"orderType"`                     // MARKET, LIMIT, STOP, STOP_LIMIT
	Quantity            string          `json:"quantity"`
	LimitPrice          string          `json:"limitPrice,omitempty"`
	StopPrice           string          `json:"stopPrice,omitempty"`
	Expiration          OrderExpiration `json:"expiration"`
	EquityMarketSession string          `json:"equityMarketSession,omitempty"` // CORE, EXTENDED
	OpenCloseIndicator  string          `json:"openCloseIndicator,omitempty"`  // OPEN, CLOSE
}

// OrderExpiration controls time-in-force for an order.
type OrderExpiration struct {
	TimeInForce    string `json:"timeInForce"`              // DAY, GTD
	ExpirationTime string `json:"expirationTime,omitempty"` // ISO 8601, for GTD
}

// MultiLegOrderRequest is the body for placing or preflighting a multi-leg order.
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
	Side               string     `json:"side"`                            // BUY, SELL
	OpenCloseIndicator string     `json:"openCloseIndicator"`              // OPEN, CLOSE
	RatioQuantity      int        `json:"ratioQuantity"`
}

// QuoteRequest is the body for requesting market quotes.
type QuoteRequest struct {
	Instruments []Instrument `json:"instruments"`
}

// OptionChainRequest is the body for requesting an option chain.
type OptionChainRequest struct {
	Instrument     Instrument `json:"instrument"`
	ExpirationDate string     `json:"expirationDate"` // YYYY-MM-DD
}

// HistoryParams are query parameters for the history endpoint.
type HistoryParams struct {
	Start     string // ISO 8601
	End       string // ISO 8601
	PageSize  int
	NextToken string
}

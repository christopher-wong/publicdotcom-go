package publicdotcom

import (
	"context"
	"sync"
	"time"
)

// --- Price Subscriptions ---

// PriceUpdate is delivered to [PriceSubscription] callbacks when a quote changes.
type PriceUpdate struct {
	Instrument Instrument
	Quote      Quote
	Err        error
}

// PriceSubscription polls for quote changes on a set of instruments and
// delivers updates via a channel. Create one with [Client.SubscribePrices].
type PriceSubscription struct {
	client      *Client
	instruments []Instrument
	interval    time.Duration
	updates     chan PriceUpdate
	done        chan struct{}
	closeOnce   sync.Once

	mu     sync.Mutex
	latest map[string]string // symbol -> last price
}

// PriceSubscriptionOption configures a [PriceSubscription].
type PriceSubscriptionOption func(*PriceSubscription)

// WithPriceInterval sets the polling interval (default 5s).
func WithPriceInterval(d time.Duration) PriceSubscriptionOption {
	return func(s *PriceSubscription) { s.interval = d }
}

// WithPriceBufferSize sets the channel buffer size (default 64).
func WithPriceBufferSize(n int) PriceSubscriptionOption {
	return func(s *PriceSubscription) { s.updates = make(chan PriceUpdate, n) }
}

// SubscribePrices creates a [PriceSubscription] that polls for quote changes
// and delivers [PriceUpdate] values on the returned channel. The subscription
// runs until [PriceSubscription.Close] is called or the context is cancelled.
//
//	sub := client.SubscribePrices(ctx, instruments, publicdotcom.WithPriceInterval(2*time.Second))
//	for update := range sub.C() {
//	    fmt.Printf("%s: %s\n", update.Instrument.Symbol, update.Quote.Last)
//	}
func (c *Client) SubscribePrices(ctx context.Context, instruments []Instrument, opts ...PriceSubscriptionOption) *PriceSubscription {
	s := &PriceSubscription{
		client:      c,
		instruments: instruments,
		interval:    5 * time.Second,
		updates:     make(chan PriceUpdate, 64),
		done:        make(chan struct{}),
		latest:      make(map[string]string),
	}
	for _, opt := range opts {
		opt(s)
	}
	go s.run(ctx)
	return s
}

// C returns the channel that receives [PriceUpdate] values.
func (s *PriceSubscription) C() <-chan PriceUpdate {
	return s.updates
}

// Close stops the subscription and closes the updates channel.
func (s *PriceSubscription) Close() {
	s.closeOnce.Do(func() { close(s.done) })
}

func (s *PriceSubscription) run(ctx context.Context) {
	defer close(s.updates)
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	// Do an initial poll immediately.
	s.poll(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.done:
			return
		case <-ticker.C:
			s.poll(ctx)
		}
	}
}

func (s *PriceSubscription) poll(ctx context.Context) {
	resp, err := s.client.GetQuotes(ctx, s.instruments)
	if err != nil {
		select {
		case s.updates <- PriceUpdate{Err: err}:
		default:
		}
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, q := range resp.Quotes {
		sym := q.Instrument.Symbol
		prev, exists := s.latest[sym]
		s.latest[sym] = q.Last

		if !exists || q.Last != prev {
			select {
			case s.updates <- PriceUpdate{Instrument: q.Instrument, Quote: q}:
			default:
			}
		}
	}
}

// --- Order Subscriptions ---

// OrderUpdate is delivered to [OrderSubscription] callbacks when an order's
// status changes.
type OrderUpdate struct {
	Order    Order
	Previous string // previous status, empty on first observation
	Err      error
}

// OrderSubscription polls for order status changes and delivers updates via
// a channel. Create one with [Client.SubscribeOrder].
type OrderSubscription struct {
	client    *Client
	orderID   string
	interval  time.Duration
	updates   chan OrderUpdate
	done      chan struct{}
	closeOnce sync.Once
}

// OrderSubscriptionOption configures an [OrderSubscription].
type OrderSubscriptionOption func(*OrderSubscription)

// WithOrderInterval sets the polling interval (default 1s).
func WithOrderInterval(d time.Duration) OrderSubscriptionOption {
	return func(s *OrderSubscription) { s.interval = d }
}

// WithOrderBufferSize sets the channel buffer size (default 16).
func WithOrderBufferSize(n int) OrderSubscriptionOption {
	return func(s *OrderSubscription) { s.updates = make(chan OrderUpdate, n) }
}

// SubscribeOrder creates an [OrderSubscription] that polls for status changes
// on a specific order and delivers [OrderUpdate] values on the returned channel.
// The subscription automatically stops when the order reaches a terminal state
// (FILLED, CANCELLED, REJECTED, EXPIRED, REPLACED) or when
// [OrderSubscription.Close] is called.
//
//	sub := client.SubscribeOrder(ctx, orderID)
//	for update := range sub.C() {
//	    fmt.Printf("order %s: %s\n", update.Order.OrderID, update.Order.Status)
//	}
func (c *Client) SubscribeOrder(ctx context.Context, orderID string, opts ...OrderSubscriptionOption) *OrderSubscription {
	s := &OrderSubscription{
		client:   c,
		orderID:  orderID,
		interval: 1 * time.Second,
		updates:  make(chan OrderUpdate, 16),
		done:     make(chan struct{}),
	}
	for _, opt := range opts {
		opt(s)
	}
	go s.run(ctx)
	return s
}

// C returns the channel that receives [OrderUpdate] values.
func (s *OrderSubscription) C() <-chan OrderUpdate {
	return s.updates
}

// Close stops the subscription and closes the updates channel.
func (s *OrderSubscription) Close() {
	s.closeOnce.Do(func() { close(s.done) })
}

var terminalStatuses = map[string]bool{
	"FILLED":    true,
	"CANCELLED": true,
	"REJECTED":  true,
	"EXPIRED":   true,
	"REPLACED":  true,
}

func (s *OrderSubscription) run(ctx context.Context) {
	defer close(s.updates)
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	var lastStatus string

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.done:
			return
		case <-ticker.C:
			order, err := s.client.GetOrder(ctx, s.orderID)
			if err != nil {
				select {
				case s.updates <- OrderUpdate{Err: err}:
				default:
				}
				continue
			}

			if order.Status != lastStatus {
				select {
				case s.updates <- OrderUpdate{Order: *order, Previous: lastStatus}:
				default:
				}
				lastStatus = order.Status
			}

			if terminalStatuses[order.Status] {
				return
			}
		}
	}
}

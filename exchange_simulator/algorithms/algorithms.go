package algorithms

import (
	"container/list"
	"fmt"
	"strings"
	"sync"
	"time"

	"../accounts"
	"../brokers"
	"../candles"
	"../indicators"
	"../orders"
	"../ticks"
	"../utils"

	"../../ga/trading_decisions"

	"github.com/robinsonstrategy/metastore"
)

func NewWithDeets(acc *accounts.Account, f func(*Algorithm, *ticks.MarketTick)) *Algorithm {
	algo := Algorithm{Account: acc, logic: f}

	return &algo
}

// ===== ALGORITHMS ================================================================================

type Algorithm struct {
	Account *accounts.Account
	Broker  brokers.Broker

	Charts map[string]map[string]*candles.CandleChart

	currencies []string

	firstTick bool
	lastTick  *ticks.MarketTick

	customData map[string]interface{}

	indis map[string]map[string]indicators.Indicator

	logic func(*Algorithm, *ticks.MarketTick)

	hasStartupDelay bool
	startupDelay    time.Duration
	firstTickAfter  time.Time

	TradingDecision trading_decisions.TradingDecision

	tickChannel   chan *ticks.MarketTick
	tickWaitGroup sync.WaitGroup

	leadingTicks map[string]*list.List

	Metadata metastore.Metastore
}

const LEADING_TICK_COUNT = 15

func (a *Algorithm) CloneLeadingTicksOntoOrder(symbol string, o *orders.Order) {
	var newTicks list.List
	leading := a.leadingTicks[symbol]

	for el := leading.Front(); el != nil; el = el.Next() {
		newTicks.PushBack(el.Value.(*ticks.MarketTick))
	}

	o.Metadata.Set("leading_ticks", newTicks)
}

func (a *Algorithm) recordLeadingTick(tick *ticks.MarketTick) {
	_, ok := a.leadingTicks[tick.Symbol]
	if !ok {
		a.leadingTicks[tick.Symbol] = new(list.List)
	}

	tt := a.leadingTicks[tick.Symbol]

	tt.PushBack(tick)

	if tt.Len() > LEADING_TICK_COUNT + 1 {
		tt.Remove(tt.Front())
	}
}

func (a *Algorithm) SetStartupDelay(t time.Duration) {
	a.startupDelay = t
	a.hasStartupDelay = true
}

func (a *Algorithm) Init(b brokers.Broker) {
	if 0 == len(a.currencies) {
		panic("you must subscribe to at least one currency")
	}

	a.Broker = b
	a.tickChannel = make(chan *ticks.MarketTick, 10000)
	a.firstTick = true

	for _, currency := range a.currencies {
		for _, indi := range a.indis[currency] {
			indi.Init()
		}
	}

	a.leadingTicks = make(map[string]*list.List)
}

func (a *Algorithm) OnTick(tick *ticks.MarketTick) {
	panic("must implement in child class")
}

func (a *Algorithm) PrintSummary() {
	a.Account.PrintSummary()
}

func (a *Algorithm) SendTick(tick *ticks.MarketTick) {
	a.tickChannel <- tick
}

func (a *Algorithm) StopReceiverLoop() {
	close(a.tickChannel)

	a.tickWaitGroup.Wait()
}

func (a *Algorithm) TickReceiverLoop() {
	a.tickWaitGroup.Add(1)

	latestTicks := make(map[string]*ticks.MarketTick)

	for {
		tick, ok := <- a.tickChannel
		if !ok {
			a.Broker.CloseAllOrders(a.Account, latestTicks)
			break
		}

		latestTicks[tick.Symbol] = tick

		if a.firstTick {
			if a.hasStartupDelay {
				// Note: This has to be set here because we don't know the time of
				//       the first tick until it's been received.
				a.firstTickAfter = tick.Time.Add(a.startupDelay)
			}

			a.firstTick = false
			a.lastTick = tick
		}

		// ----- PREPROCESSING -------------------------------------------------------------

		if a.Account.HasOpenOrders() {
			for el := a.Account.Orders.Front(); el != nil; el = el.Next() {
				order := el.Value.(*orders.Order)

				if order.Symbol != tick.Symbol || order.IsClosed() {
					continue
				}

				order.OnTick(tick)
				order.RecordTick(tick)

				if order.ProcessStops(tick) {
					a.Broker.CloseOrder(a.Account, order, tick)
				}
			}

			_       = a.Account.UpdateBalance()
			equity := a.Account.UpdateEquity()
			margin := a.Account.UpdateMarginAvailable()

			if a.Account.HasExceededDrawdown() {
				fmt.Printf(
					"Account has violated max drawdown, closing all orders (curr: %.2f%%, max: %.2f%%)\n",
					a.Account.GetDrawdown(),
					a.Account.GetDrawdownLimit(),
				)

				a.Broker.CloseAllOrders(a.Account, latestTicks)
			} else if equity <= accounts.MINIMUM_EQUITY {
				fmt.Printf(
					"Equity is below minimum, closing all orders (curr: %.2f, min: %.2f)\n",
					equity,
					accounts.MINIMUM_EQUITY,
				)

				a.Account.MarginCalled()
				a.Broker.CloseAllOrders(a.Account, latestTicks)
			} else if margin <= accounts.MINIMUM_MARGIN {
				fmt.Printf(
					"Margin available is below minimum, TODO: close biggest order (curr: %s, min: %s)\n",
					utils.FormatMoney(margin),
					utils.FormatMoney(accounts.MINIMUM_EQUITY),
				)

				a.Account.MarginCalled()
				a.Broker.CloseAllOrders(a.Account, latestTicks)
			}

			// If we just got margin called or hit the DD limit, empty our channel and exit.
			if !a.Account.CanTrade() {
				for {
					_, ok := <- a.tickChannel
					if !ok {
						break
					}
				}
			}
		}

		a.recordLeadingTick(tick)

		a.updateCharts(tick)
		a.runIndicators(tick)

		if a.hasStartupDelay && tick.Time.Before(a.firstTickAfter) {
			continue
		}

		// ----- PROCESSING ----------------------------------------------------------------

		a.logic(a, tick)

		a.lastTick = tick
	}

	a.tickWaitGroup.Done()
}

// ===== CURRENCY ==================================================================================

func (a *Algorithm) AddCurrency(currencies ...string) {
	for _, currency := range currencies {
		// TODO: panic if currency already exists
		a.currencies = append(a.currencies, currency)

		if !strings.HasSuffix(currency, "USD") {
			fmt.Println("")
			fmt.Println("******************************************************************")
			fmt.Println(" WARNING: non-USD base currencies will not calculate P/L properly")
			fmt.Println("******************************************************************")
			fmt.Println("")
		}
	}
}

func (a *Algorithm) WantsCurrency(s string) bool {
	return utils.StringArrayContainsString(a.currencies, s)
}

// ===== CANDLESTICK CHARTS ========================================================================

func (a *Algorithm) AttachCharts(periods ...string) {
	for i := range periods {
		a.attachChart(periods[i])
	}
}

func (a *Algorithm) attachChart(period string) {
	if nil == a.Charts {
		a.Charts = make(map[string]map[string]*candles.CandleChart)
	}

	exists := false

	for _, currency := range a.currencies {
		_, exists = a.Charts[currency]
		if !exists {
			a.Charts[currency] = make(map[string]*candles.CandleChart)
		}

		_, exists = a.Charts[currency][period]
		if exists {
			panic("Chart already attached: " + period)
		}

		a.Charts[currency][period] = candles.NewCandleChart(period, 61)
	}
}

func (a *Algorithm) updateCharts(tick *ticks.MarketTick) {
	for _, currency := range a.currencies {
		if currency != tick.Symbol {
			continue
		}

		for _, chart := range a.Charts[currency] {
			chart.Update(tick, a.lastTick)
		}
	}
}

// ===== INDICATORS ================================================================================

func (a *Algorithm) AddIndicator(name string, indi func() indicators.Indicator) {
	if nil == a.indis {
		a.indis = make(map[string]map[string]indicators.Indicator)
	}

	exists := false

	for _, currency := range a.currencies {
		_, exists = a.indis[currency]
		if !exists {
			a.indis[currency] = make(map[string]indicators.Indicator)
		}

		_, exists = a.indis[currency][name]
		if exists {
			panic("attempted to add indicator \"" + name + "\" which already exists!")
		}

		a.indis[currency][name] = indi()
	}
}

func (a *Algorithm) ReadIndicator(currency, name string) indicators.Indicator {
	indi, ok := a.indis[currency][name]
	if !ok {
		panic("Unknown indicator: " + name)
	}

	return indi
}

func (a *Algorithm) dumpIndis() {
	for _, currency := range a.currencies {
		for key, value := range a.indis[currency] {
			fmt.Printf("[%s][%s] -> %#v\n", currency, key, value)
		}
	}
}

func (a *Algorithm) runIndicators(tick *ticks.MarketTick) {
	for _, currency := range a.currencies {
		if currency != tick.Symbol {
			continue
		}

		for _, indi := range a.indis[currency] {
			indi.OnTick(a, tick)
		}
	}
}

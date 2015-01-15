package exchanges

import (
	"container/list"
	"fmt"
	"log"
	"math"
	"runtime"
	"time"

	"../accounts"
	"../algorithms"
	"../orders"
	"../stops"
	"../ticks"
	"../utils"
)

func NewWithDeets(mt ticks.MarketTicker) *Exchange {
	e := Exchange{tickSource: mt}

	return &e
}

type Exchange struct {
	algorithms list.List
	tickSource ticks.MarketTicker

	totalOrdersProcessed int64
	totalTicksProcessed  int64

	firstTick *ticks.MarketTick
	lastTick  *ticks.MarketTick

	runStartedAt time.Time
}

func (e *Exchange) AddAlgorithm(a *algorithms.Algorithm) {
	e.algorithms.PushBack(a)
	a.Init(e)
}

func (e *Exchange) CloseAllOrders(a *accounts.Account, finalTicks map[string]*ticks.MarketTick) {
	for el := a.Orders.Front(); el != nil; el = el.Next() {
		o := el.Value.(*orders.Order)

		if o.IsClosed() {
			continue
		}

		e.CloseOrder(a, o, finalTicks[o.Symbol])
	}
}

func bidOrAskToClose(isBuy bool, t *ticks.MarketTick) float64 {
	if isBuy {
		return t.OpenBid
	} else {
		return t.OpenAsk
	}
}

func bidOrAskToOpen(isBuy bool, t *ticks.MarketTick) float64 {
	if isBuy {
		return t.OpenAsk
	} else {
		return t.OpenBid
	}
}

func (e Exchange) CloseOrder(a *accounts.Account, o *orders.Order, tick *ticks.MarketTick) {
	if o.IsClosed() {
		panic("can't close already closed order")
	}

	closePrice := bidOrAskToClose(o.IsBuy(), tick)

	// log.Printf("Closing @ %.5f\n", closePrice)
	// log.Printf("Closing tick: %#v\nClosing Order: %#v\n\n=============================\n\n", tick, o)

	o.ClosePrice = closePrice
	o.ClosedAt   = tick.Time
	o.CloseBid   = tick.OpenBid
	o.CloseAsk   = tick.OpenAsk

	o.OrdersOpenAtClose = int64(len(a.OpenOrders()))

	a.RealizeProfit(o)

	o.BalanceAtClose = a.GetBalance()
	o.EquityAtClose = a.GetEquity()

	o.DrawdownAtClose = a.CurrentDrawdown()
}

func openOrder(direction orders.TradeDirection, a *accounts.Account, symbol string, tick *ticks.MarketTick, lots float64, sl stops.StopLoss, tp stops.TakeProfit) *orders.Order {
	openPrice := bidOrAskToOpen(direction == orders.BUY, tick)

	// TODO: validate symbol is a real currency pair
	utils.EnsureZeroOrGreater(openPrice)
	utils.EnsureZeroOrGreater(lots)

	// TODO: slippage from the exchange should be calculated here
	// TODO: determine if order can be opened due to slippage / AllowedSlippage

	o := orders.Order{
		Symbol: symbol,

		OpenedAt: tick.Time,
		OpenPrice: openPrice,
		Direction: direction,
		LotSize: lots,

		EquityAtOpen: a.GetEquity(),
		BalanceAtOpen: a.GetBalance(),

		OpenBid: tick.OpenBid,
		OpenAsk: tick.OpenAsk,
		LowestBid: tick.OpenBid,
		LowestAsk: tick.OpenAsk,
		HighestBid: tick.OpenBid,
		HighestAsk: tick.OpenAsk,

		OrdersOpenAtOpen: int64(len(a.OpenOrders())),
	}

	// TODO: refactor this out so it just uses pips
	if sl.Set {
		o.SetStopLoss(sl.Pips)
	}

	if tp.Set {
		o.SetTakeProfit(tp.Pips)
	}

	tick.Metadata.Set("percent_to_tp", 0.0)
	tick.Metadata.Set("percent_to_sl", 0.0)

	o.DrawdownAtOpen = a.CurrentDrawdown()
	o.Ticks.PushBack(tick)

	a.AddOrder(&o)

	return &o
}

func (e *Exchange) OpenBuyOrder(a *accounts.Account, symbol string, tick *ticks.MarketTick, lots float64, sl stops.StopLoss, tp stops.TakeProfit) *orders.Order {
	e.totalOrdersProcessed += 1
	return openOrder(orders.BUY, a, symbol, tick, lots, sl, tp)
}

func (e *Exchange) OpenSellOrder(a *accounts.Account, symbol string, tick *ticks.MarketTick, lots float64, sl stops.StopLoss, tp stops.TakeProfit) *orders.Order {
	e.totalOrdersProcessed += 1
	return openOrder(orders.SELL, a, symbol, tick, lots, sl, tp)
}

func (e *Exchange) Run() {
	numAlgos := int64(e.algorithms.Len())

	if 0 == numAlgos {
		fmt.Println("No algorithms added, so nothing to run")
		return
	}

	runtime.GOMAXPROCS(int(numAlgos) + 10)

	e.runStartedAt = time.Now()

	fmt.Printf("Simulating exchange for %d algorithms\n", numAlgos)
	fmt.Println("")

	for algo := e.algorithms.Front(); algo != nil; algo = algo.Next() {
		go algo.Value.(*algorithms.Algorithm).TickReceiverLoop()
	}

	symbolsSeen := make(map[string]bool)
	tickDeviance := make(map[string]*ticks.MarketTick)
	tickCount := int64(0)

	first := true

	// FXCM switched to an ECN model with lower spreads the week of 2014/10/05
	ecnChangeoverCutoff := time.Date(2014, time.October, 5, 0, 0, 0, 0, time.UTC)

	for tick := range e.tickSource.Ticks() {
		if !tick.Time.Before(ecnChangeoverCutoff) {
			continue
		}

		if first {
			first = false
			e.firstTick = tick
		}

		_, ok := symbolsSeen[tick.Symbol]
		if !ok {
			symbolsSeen[tick.Symbol] = true
			tickDeviance[tick.Symbol] = tick
		}

		previousTick, ok := tickDeviance[tick.Symbol]
		if !ok {
			panic("could not find previous tick for " + tick.Symbol)
		}

		validateTick(tick, previousTick)

		for algo := e.algorithms.Front(); algo != nil; algo = algo.Next() {
			currAlgo := algo.Value.(*algorithms.Algorithm)

			if currAlgo.WantsCurrency(tick.Symbol) {
				currAlgo.SendTick(tick)
				e.totalTicksProcessed += 1
			}
		}

		e.lastTick = tick
		tickDeviance[tick.Symbol] = tick
		tickCount += 1
	}

	// TODO: Move this outta here!
	for algo := e.algorithms.Front(); algo != nil; algo = algo.Next() {
		a := algo.Value.(*algorithms.Algorithm)

		a.StopReceiverLoop()
		a.PrintSummary()
		a.Account.PrintWeeklyStats(e.firstTick.Time, e.lastTick.Time)
	}

	seconds := time.Since(e.runStartedAt).Seconds()

	fmt.Printf(
		"***** SIMULATION STATS *****\n\n" +
		"Algorithms simulated: %s\n" +
		"Time period: %s - %s\n" +
		"Ticks in dataset: %s\n" +
		"Total orders executed: %s\n" +
		"Total ticks processed: %s\n" +
		"Ticks processed per second: %s\n" +
		"Execution time: %.0f seconds\n\n",
		utils.AddCommas(numAlgos),
		yearMonthDayFromTime(e.firstTick.Time),
		yearMonthDayFromTime(e.lastTick.Time),
		utils.AddCommas(tickCount),
		utils.AddCommas(e.totalOrdersProcessed),
		utils.AddCommas(e.totalTicksProcessed),
		utils.AddCommas(int64(float64(e.totalTicksProcessed) / (float64(seconds) + 0.00001))),
		seconds,
	)
}

func yearMonthDayFromTime(t time.Time) string {
	y, m, d := t.Date()

	return fmt.Sprintf("%d/%d/%d", y, m, d)
}

// ===== TICKS =====================================================================================

const BID_ASK_DEVIATION_TOLERANCE = float64(0.005)

func validateTick(current *ticks.MarketTick, last *ticks.MarketTick) {
	sameDay := current.Time.Day() == last.Time.Day()

	bidDiff := last.OpenBid * BID_ASK_DEVIATION_TOLERANCE
	if !sameDay && (current.OpenBid >= (last.OpenBid + bidDiff) || current.OpenBid <= (last.OpenBid - bidDiff)) {
		fmt.Printf(
			"BAD DATA: bid has deviated over %.1f%% (%.1f%%) from last Bid (curr: %.5f, last: %.5f)\n",
			BID_ASK_DEVIATION_TOLERANCE * 100,
			math.Abs(((current.OpenBid - last.OpenBid) / current.OpenBid) * 100),
			current.OpenBid,
			last.OpenBid,
		)
		fmt.Printf("Last:    %s - %#v\n", last.Time, last)
		fmt.Printf("Current: %s - %#v\n", current.Time, current)
		log.Fatalln("")
	}

	askDiff := last.OpenAsk * BID_ASK_DEVIATION_TOLERANCE
	if !sameDay && (current.OpenAsk >= (last.OpenAsk + askDiff) || current.OpenAsk <= (last.OpenAsk - askDiff)) {
		fmt.Printf(
			"BAD DATA: ask has deviated over %.1f%% (%.1f%%) from last ask (curr: %.5f, last: %.5f)\n",
			BID_ASK_DEVIATION_TOLERANCE * 100,
			math.Abs(((current.OpenAsk - last.OpenAsk) / current.OpenAsk) * 100),
			current.OpenAsk,
			last.OpenAsk,
		)
		fmt.Printf("Last:    %s - %#v\n", last.Time, last)
		fmt.Printf("Current: %s - %#v\n", current.Time, current)
		log.Fatalln("")
	}

	if last.Time.Unix() > current.Time.Unix() {
		log.Fatalf(
			"BAD DATA: current tick is newer than last tick (curr: %#v, last: %#v)\n",
			current.Time,
			last.Time,
		)
	}
}


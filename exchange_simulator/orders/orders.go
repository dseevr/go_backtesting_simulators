package orders

import (
	"container/list"
	"fmt"
	"time"

	"../pips"
	"../quotes"
	"../stops"
	"../ticks"
	"../utils"

	"github.com/robinsonstrategy/metastore"
)

// ===== TRADE DIRECTION ===========================================================================

type TradeDirection int64

const (
	BUY = TradeDirection(1)
	SELL = TradeDirection(-1)
)

// ===== ORDERS ====================================================================================

type Order struct {
	Symbol string

	OpenedAt          time.Time
	ClosedAt          time.Time

	OpenPrice         float64   // price filled at
	ClosePrice        float64   // price filled at
	DesiredOpenPrice  float64
	DesiredClosePrice float64

	Direction         TradeDirection
	LotSize           float64

	AllowedSlippage   float64 // TODO: implement this

	stopLoss   []stops.StopLoss
	takeProfit []stops.TakeProfit

	StopLossHit   bool
	TakeProfitHit bool

	// for calculating spreads and actual slippage
	OpenBid  float64
	OpenAsk  float64
	CloseBid float64
	CloseAsk float64

	// extra fields for debugging / analytics
	EquityAtOpen   float64
	EquityAtClose  float64
	BalanceAtOpen  float64
	BalanceAtClose float64

	LowestBid      float64
	LowestAsk      float64
	HighestBid     float64
	HighestAsk     float64

	OrdersOpenAtOpen  int64
	OrdersOpenAtClose int64

	DrawdownAtOpen  float64
	DrawdownAtClose float64

	ClosestPercentageToTakeProfit float64
	ClosestPercentageToStopLoss   float64

	Ticks list.List

	Metadata metastore.Metastore
}

func (o *Order) MaxPossibleLoss() float64 {
	if !o.GetStopLoss().Set {
		panic("no stop loss is set, so max possible loss is incalculable")
	}

	if o.IsBuy() {
		q  := quotes.NewQuote(o.Symbol, o.OpenAsk)
		q2 := quotes.NewQuote(o.Symbol, o.OpenAsk)
		q2.SubtractPips(o.GetStopLoss().Pips)
		return float64(q.ProfitInPipsAt(q2)) * o.LotSize * 10
	} else {
		q  := quotes.NewQuote(o.Symbol, o.OpenBid)
		q2 := quotes.NewQuote(o.Symbol, o.OpenBid)
		q2.SubtractPips(o.GetStopLoss().Pips)
		return float64(q.ProfitInPipsAt(q2)) * o.LotSize * 10
	}
}

func (o *Order) PercentToTakeProfit() float64 {
	if !o.GetTakeProfit().Set {
		return 0.0
	}

	percent := 0.0
	currentTick := o.Ticks.Back().Value.(*ticks.MarketTick)

	if o.IsBuy() {
		movement := currentTick.OpenBid - o.OpenBid
		mrange   := o.TakeProfitPrice() - o.OpenBid
		percent = (movement / mrange) * 100.0
	} else {
		movement := currentTick.OpenAsk - o.OpenAsk
		mrange   := o.TakeProfitPrice() - o.OpenAsk
		percent = (movement / mrange) * 100.0
		percent = -(percent)
	}

	if percent < 0.0 {
		percent = 0.0
	}

	return percent
}

func (o *Order) PercentToStopLoss() float64 {
	if !o.GetStopLoss().Set {
		return 0.0
	}

	percent := 0.0
	currentTick := o.Ticks.Back().Value.(*ticks.MarketTick)

	if o.IsBuy() {
		movement := o.OpenAsk - currentTick.OpenAsk
		mrange   := o.OpenAsk - o.StopLossPrice()
		percent = (movement / mrange) * 100.0
	} else {
		movement := o.OpenBid - currentTick.OpenBid
		mrange   := o.StopLossPrice() - o.OpenBid
		percent = (movement / mrange) * 100.0
		percent = -(percent)
	}

	if percent < 0.0 {
		percent = 0.0
	}

	return percent
}

func (o *Order) PrintDetails() {
	fmt.Printf(
		"Symbol: %s, Profit: %s, Pips: %.1f\n",
		o.Symbol,
		utils.FormatMoney(o.Profit()),
		o.ProfitInPips(),
	)
	fmt.Printf(
		"Duration: %s -> %s (O/C day: %d/%d)\n",
		utils.NiceTimeFormat(o.OpenedAt),
		utils.NiceTimeFormat(o.ClosedAt),
		o.OpenedAt.YearDay(),
		o.ClosedAt.YearDay(),
	)
	fmt.Printf("Open: %.5f, Close: %.5f\n", o.OpenPrice, o.ClosePrice)
	fmt.Printf(
		"SL/TP pips: %.1f/%.1f - Prices: %.5f/%.5f - Times set: %d/%d\n",
		o.GetStopLoss().Pips,
		o.GetTakeProfit().Pips,
		o.StopLossPrice(),
		o.TakeProfitPrice(),
		len(o.stopLoss),
		len(o.takeProfit),
	)
	fmt.Printf(
		"%% to TP/SL: %.1f%%/%.1f%%, High B/A: %.5f/%.5f, Low B/A: %.5f/%.5f\n",
		o.ClosestPercentageToTakeProfit,
		o.ClosestPercentageToStopLoss,
		o.HighestBid,
		o.HighestAsk,
		o.LowestBid,
		o.LowestAsk,
	)
	fmt.Printf(
		"DD @ open/close: %.2f%%/%.2f%%\n",
		o.DrawdownAtOpen,
		o.DrawdownAtClose,
	)

	fmt.Printf("Ticks: %d\n", o.Ticks.Len())
	count := int64(1)

	// leading_ticks := o.Metadata.Get("leading_ticks").(list.List)

	// fmt.Println("Pre-entry:")
	// for e := leading_ticks.Front(); e != nil; e = e.Next() {
	// 	t := e.Value.(*ticks.MarketTick)

	// 	fmt.Printf(
	// 		"%3d - %s - B: %.5f A: %.5f Spread: %.1f Vol: %4d Symbol: %s\n",
	// 		count,
	// 		utils.NiceTimeFormat(t.Time),
	// 		t.OpenBid,
	// 		t.OpenAsk,
	// 		t.Spread(),
	// 		t.Volume,
	// 		t.Symbol,
	// 	)

	// 	count += 1
	// }

	fmt.Println("===== TRADE OPENED =====")

	count = 1
	for e := o.Ticks.Front(); e != nil; e = e.Next() {
		t := e.Value.(*ticks.MarketTick)

		var pttp, ptsl float64

		res, ok := t.Metadata.Get("percent_to_tp")
		if ok {
			pttp = res.(float64)
		}

		res, ok = t.Metadata.Get("percent_to_sl")
		if ok {
			ptsl = res.(float64)
		}

		fmt.Printf(
			"%3d - %s - %% TP/SL: %3.1f%%/%3.1f%% B: %.5f A: %.5f Spread: %.1f Vol: %4d Symbol: %s\n",
			count,
			utils.NiceTimeFormat(t.Time),
			pttp,
			ptsl,
			t.OpenBid,
			t.OpenAsk,
			t.Spread(),
			t.Volume,
			t.Symbol,
		)

		count += 1
	}

	fmt.Println("")
}

func (o *Order) RecordTick(tick *ticks.MarketTick) {
	o.Ticks.PushBack(tick)
}

func (o *Order) IsWinner() bool {
	return o.Profit() > 0.0
}

func (o *Order) IsLoser() bool {
	return !o.IsWinner()
}

func (o *Order) Commission() float64 {
	return o.LotSize * 0 // TODO: get this from the exchange somehow
}

func (o *Order) IsBuy() bool {
	return BUY == o.Direction
}

func (o *Order) IsSell() bool {
	return !o.IsBuy()
}

func (o *Order) IsOpen() bool {
	return o.OpenedAt.Unix() > o.ClosedAt.Unix();
}

func (o *Order) IsClosed() bool {
	return !o.IsOpen();
}

func (o *Order) LongOrShort() string {
	if o.IsBuy() {
		return "LONG"
	} else {
		return "SHORT"
	}
}

func (o *Order) Profit() float64 {
	// TODO: support non-USD base currency pairs

	// 0.01 lots -> 1.23 pips -> $0.12
	// 1.00 lots -> 1.23 pips -> $12.30

	lotMultiplier := o.LotSize * 10
	return float64(o.ProfitInPips()) * lotMultiplier
}

func (o *Order) ProfitInPips() pips.Pip {
	closePrice := 0.0
	lastTick := o.Ticks.Back().Value.(*ticks.MarketTick)

	if o.IsBuy() {
		closePrice = lastTick.OpenBid
	} else {
		closePrice = lastTick.OpenAsk
	}

	q1 := quotes.NewQuote(o.Symbol, o.OpenPrice)
	q2 := quotes.NewQuote(o.Symbol, closePrice)

	pips := q1.ProfitInPipsAt(q2)

	if o.IsSell() {
		pips = -(pips)
	}

	return pips
}

func (o *Order) OnTick(tick *ticks.MarketTick) {
	if tick.OpenBid > o.HighestBid {
		o.HighestBid = tick.OpenBid
	} else if tick.OpenBid < o.LowestBid {
		o.LowestBid = tick.OpenBid
	}

	if tick.OpenAsk > o.HighestAsk {
		o.HighestAsk = tick.OpenAsk
	} else if tick.OpenAsk < o.LowestAsk {
		o.LowestAsk = tick.OpenAsk
	}

	pttp := o.PercentToTakeProfit()

	if pttp > o.ClosestPercentageToTakeProfit {
		o.ClosestPercentageToTakeProfit = pttp
	}

	ptsl := o.PercentToStopLoss()

	if ptsl > o.ClosestPercentageToStopLoss {
		o.ClosestPercentageToStopLoss = ptsl
	}

	tick.Metadata.Set("percent_to_tp", pttp)
	tick.Metadata.Set("percent_to_sl", ptsl)
}

// ===== STOP LOSS & TAKE PROFIT ===================================================================

func (o *Order) GetStopLoss() stops.StopLoss {
	if 0 == len(o.stopLoss) {
		return stops.NoStopLoss()
	}

	return o.stopLoss[len(o.stopLoss) - 1]
}

func (o *Order) SetStopLoss(p pips.Pip) {

	o.stopLoss = append(o.stopLoss, stops.NewStopLoss(p))
}

func (o *Order) GetTakeProfit() stops.TakeProfit {
	if 0 == len(o.takeProfit) {
		return stops.NoTakeProfit()
	}

	return o.takeProfit[len(o.takeProfit) - 1]
}

func (o *Order) SetTakeProfit(p pips.Pip) {
	o.takeProfit = append(o.takeProfit, stops.NewTakeProfit(p))
}

func (o *Order) StopLossPrice() float64 {
	if !o.GetStopLoss().Set {
		return 0.0
	}

	if o.IsBuy() {
		val := quotes.NewQuote(o.Symbol, o.OpenAsk)
		val.SubtractPips(o.GetStopLoss().Pips)
		return val.ToFloat64()
	} else {
		val := quotes.NewQuote(o.Symbol, o.OpenBid)
		val.AddPips(o.GetStopLoss().Pips)
		return val.ToFloat64()
	}
}

func (o *Order) TakeProfitPrice() float64 {
	if !o.GetTakeProfit().Set {
		return 0.0
	}

	if o.IsBuy() {
		val := quotes.NewQuote(o.Symbol, o.OpenAsk)
		val.AddPips(o.GetTakeProfit().Pips)
		return val.ToFloat64()
	} else {
		val := quotes.NewQuote(o.Symbol, o.OpenBid)
		val.SubtractPips(o.GetTakeProfit().Pips)
		return val.ToFloat64()
	}
}

func (o *Order) checkStopLoss(tick *ticks.MarketTick) bool {
	if !o.GetStopLoss().Set {
		return false
	}

	if o.IsBuy() && o.StopLossPrice() >= tick.OpenBid {
		o.StopLossHit = true
		return true
	} else if o.IsSell() && o.StopLossPrice() <= o.OpenAsk {
		o.StopLossHit = true
		return true
	}

	return false
}

func (o *Order) checkTakeProfit(tick *ticks.MarketTick) bool {
	if !o.GetTakeProfit().Set {
		return false
	}

	if o.IsBuy() && o.TakeProfitPrice() <= tick.OpenBid {
		o.TakeProfitHit = true
		return true
	} else if o.IsSell() && o.TakeProfitPrice() >= tick.OpenAsk {
		o.TakeProfitHit = true
		return true
	}

	return false
}

func (o *Order) ProcessStops(tick *ticks.MarketTick) bool {
	return o.checkStopLoss(tick) || o.checkTakeProfit(tick)
}

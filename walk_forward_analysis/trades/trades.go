package trades

import (
	"container/list"
	"time"

	"../ticks"
)

type Trade interface {
	AppendTick(*ticks.Tick)

	Close(*ticks.Tick)

	IsClosed() bool
	IsOpen() bool
	IsLong() bool
	IsShort() bool

	Profit() float64
}

type BaseTrade struct {
	Symbol string

	OpenedAt          time.Time
	ClosedAt          time.Time

	OpenPrice         float64   // price filled at
	ClosePrice        float64   // price filled at
	// DesiredOpenPrice  float64
	// DesiredClosePrice float64

	// Direction         TradeDirection
	LotSize           float64

	// AllowedSlippage   float64 // TODO: implement this

	// stopLoss   []stops.StopLoss
	// takeProfit []stops.TakeProfit

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

	ticks list.List

	// Metadata metastore.Metastore
}

func (bt *BaseTrade) IsOpen() bool {
	return bt.ClosedAt.Unix() > bt.OpenedAt.Unix()
}

func (bt *BaseTrade) IsClosed() bool {
	return !bt.IsOpen()
}

func (bt *BaseTrade) AppendTick(tick *ticks.Tick) {
	bt.ticks.PushBack(tick)
}

func (bt *BaseTrade) LastTick() *ticks.Tick {
	return bt.ticks.Back().Value.(*ticks.Tick)
}

// ===== LONG TRADE ================================================================================

func NewLongTrade() Trade {
	return &LongTrade{}
}

type LongTrade struct {
	BaseTrade
}

func (lt LongTrade) Close(tick *ticks.Tick) {
	lt.ClosedAt = tick.Time
	lt.ClosePrice = tick.OpenBid
}

func (lt LongTrade) IsLong() bool {
	return true
}

func (lt LongTrade) IsShort() bool {
	return false
}

func (lt *LongTrade) Profit() float64 {
	return lt.ClosePrice - lt.OpenPrice
}

// ===== SHORT TRADE ===============================================================================

func NewShortTrade() Trade {
	return &ShortTrade{}
}

type ShortTrade struct {
	BaseTrade
}

func (st ShortTrade) Close(tick *ticks.Tick) {
	st.ClosedAt = tick.Time
	st.ClosePrice = tick.OpenAsk
}

func (st ShortTrade) IsLong() bool {
	return false
}

func (st ShortTrade) IsShort() bool {
	return true
}

func (st *ShortTrade) Profit() float64 {
	return st.OpenPrice - st.ClosePrice
}

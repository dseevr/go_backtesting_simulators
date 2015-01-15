package candles

import (
	"container/list"
	"fmt"

	"../ticks"
)

const (
	M1  = 60
	M5  = M1 * 5
	M15 = M1 * 15
	M30 = M1 * 30
	H1  = M1 * 60
	H4  = H1 * 4
	D1  = H1 * 24
	W1  = D1 * 7
)

var validCharts = map[string]int{
	"M1":  M1,
	"M5":  M5,
	"M15": M15,
	"M30": M30,
	"H1":  H1,
	"H4":  H4,
	"D1":  D1,
	"W1":  W1,
}

// ===== CANDLE ====================================================================================

type Candle struct {
	OpenBid  float64
	OpenAsk  float64
	CloseBid float64
	CloseAsk float64

	HighBid float64
	HighAsk float64
	LowBid  float64
	LowAsk  float64

	Volume int

	id int
}

// // Note: can legitimately return negative values
// func (c *Candle) Spread() float64 {
// 	// TODO: get rid of hardcoded currency
// 	return float64(quotes.DifferenceInPips("EURUSD", c.OpenBid, c.OpenAsk))
// }

// ===== CANDLE CHART ==============================================================================

func NewCandleChart(period string, maxCandles int) *CandleChart {
	cc := CandleChart{maxCandles: maxCandles}

	val, ok := validCharts[period]
	if !ok {
		panic("unknown candle chart period: " + period)
	}

	cc.period = val

	return &cc
}

type CandleChart struct {
	list.List

	maxCandles int
	period     int
}

func (cc *CandleChart) newCandleFromTick(id int, tick *ticks.Tick) {
		cs := Candle{
			OpenBid:  tick.OpenBid,
			OpenAsk:  tick.OpenAsk,
			CloseBid: tick.CloseBid,
			CloseAsk: tick.CloseAsk,
			HighBid:  tick.HighBid,
			HighAsk:  tick.HighAsk,
			LowBid:   tick.LowBid,
			LowAsk:   tick.LowAsk,
			Volume:   tick.Volume,
			id: id,
		}

		cc.PushFront(&cs)
}

func (cc *CandleChart) GetCandles(n int) []*Candle {
	if n <= 0 {
		panic("n must be >= 0")
	}

	c := []*Candle{}

	count := 0

	// we don't want the current candle, start from the previous one
	e := cc.Front().Next()

	for ; count < n; e = e.Next() {
		if nil == e {
			panic(fmt.Sprintf(
				"not enough candles to fulfill request! wanted: %d, have: %d",
				n,
				cc.Len() - 1, // don't count latest candle, see above
			))
		}
		c = append(c, e.Value.(*Candle))
		count++
	}

	return c
}

func (cc *CandleChart) Print() {
	fmt.Printf("===== %d chart =====\n")

	for e := cc.Front(); e != nil; e = e.Next() {
		fmt.Printf("id: %d, val: %#v\n", e.Value.(*Candle).id, e.Value.(*Candle))
	}
}

func (cc *CandleChart) Update(current *ticks.Tick) {
	numCandles := cc.Len()
	id := current.Time.Unix() / cc.period

	if 0 == numCandles {
		cc.newCandleFromTick(id, current)
		return
	}

	cs := cc.Front().Value.(*Candle)

	if id > cs.id {
		cc.newCandleFromTick(id, current)

		if numCandles >= cc.maxCandles {
			cc.Remove(cc.Back())
		}

		return
	}
}


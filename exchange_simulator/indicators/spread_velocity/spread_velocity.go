package spread_velocity

import (
	// "fmt"
	"../../algorithms"
	"../../ticks"
)

func NewSV(n int64) *SpreadVelocity {
	return &SpreadVelocity{numCandles: n}
}

type SpreadVelocity struct {
	expanding   bool
	contracting bool

	numCandles int64
}

func (sv *SpreadVelocity) Init() {
	if sv.numCandles < 3 || sv.numCandles > 60 {
		panic("number of candles must be >= 3 and <= 60")
	}
}

func (sv *SpreadVelocity) OnTick(a interface{}, tick *ticks.MarketTick) {
	sv.expanding   = false
	sv.contracting = false

	chart := a.(*algorithms.Algorithm).Charts[tick.Symbol]["M1"]
	if int64(chart.Len()) < sv.numCandles + 1 {
		return
	}

	var first, last float64

	count := int64(0)
	for _, val := range chart.GetCandles(sv.numCandles) {
		if count == 0 {
			first = val.Spread()
		} else if count == sv.numCandles - 1 {
			last = val.Spread()
		}

		count++
	}

	// fmt.Printf("F: %.1f, L: %.1f\n", first, last)

	if first > last {
		sv.contracting = true
	} else if first < last {
		sv.expanding = true
	}

}

func (sv *SpreadVelocity) IsExpanding() bool {
	return sv.expanding
}

func (sv *SpreadVelocity) IsContracting() bool {
	return sv.contracting
}

func (sv *SpreadVelocity) IsConstant() bool {
	return !sv.expanding && !sv.contracting
}

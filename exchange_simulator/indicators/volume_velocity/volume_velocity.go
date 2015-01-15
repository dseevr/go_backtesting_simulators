package volume_velocity

import (
	// "fmt"

	"../../algorithms"
	"../../ticks"
	"../../utils"
)

const NUM_CANDLES = 5

func NewVV(n int64, change float64) *VolumeVelocity {
	return &VolumeVelocity{numCandles: n, change: change}
}

type VolumeVelocity struct {
	increasing bool
	decreasing bool

	numCandles int64
	change float64
}

func (vv *VolumeVelocity) Init(){
	if vv.numCandles < 3 || vv.numCandles > 60 {
		panic("number of candles must be >= 3 and <= 60")
	}

	if vv.change <= 0.00 || vv.change >= 100.00 {
		panic("change must be > 0.0 and < 100.0")
	}
}

func (vv *VolumeVelocity) OnTick(a interface{}, tick *ticks.MarketTick) {
	vv.increasing = false
	vv.decreasing = false

	chart := a.(*algorithms.Algorithm).Charts[tick.Symbol]["M1"]
	if chart.Len() < NUM_CANDLES + 1 {
		return
	}

	var first, last float64

	count := 0
	for _, val := range chart.GetCandles(NUM_CANDLES) {
		if count == 0 {
			first = float64(val.Volume)
		} else if count == NUM_CANDLES - 1 {
			last = float64(val.Volume)
		}

		count++
	}

	move := utils.PercentChange(first, last)

	// fmt.Printf("F: %.1f, L: %.1f\n", first, last)

	if move >= vv.change {
		vv.increasing = true
	} else if move <= -(vv.change) {
		vv.decreasing = true
	}

}

func (vv *VolumeVelocity) IsIncreasing() bool {
	return vv.increasing
}

func (vv *VolumeVelocity) IsDecreasing() bool {
	return vv.decreasing
}

func (vv *VolumeVelocity) IsConstant() bool {
	return !vv.increasing && !vv.decreasing
}

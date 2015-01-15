package steve_turn_detector

import (
	"fmt"

	"../../algorithms"
	"../../ticks"
)

func NewSTD(lower, upper float64) *SteveTurnDetector {
	return &SteveTurnDetector{lower: lower, upper: upper}
}

type SteveTurnDetector struct {
	lower float64
	upper float64

	turningUp   bool
	turningDown bool
}

func (std *SteveTurnDetector) Init(){
	if std.lower >= std.upper {
		panic(fmt.Sprintf(
			"lower (%.5f) can't be higher than upper (%.5f)\n",
			std.lower,
			std.upper,
		))
	}

	if std.upper > 1.00 {
		panic(fmt.Sprintf(
			"upper (%.5f) can't be > 1.0",
			std.upper,
		))
	}

	if std.lower < 0.0 {
		panic(fmt.Sprintf(
			"lower (%.5f) can't be < 0.0",
			std.lower,
		))
	}
}

func (std *SteveTurnDetector) OnTick(a interface{}, tick *ticks.MarketTick) {
	chart := a.(*algorithms.Algorithm).Charts[tick.Symbol]["M1"]
	if chart.Len() < 61 {
		return
	}

	ma5 := 0.0
	for _, val := range chart.GetCandles(5) {
		ma5 += val.OpenBid
	}
	ma5 /= 5.0

	ma60 := 0.0
	for _, val := range chart.GetCandles(60) {
		ma60 += val.OpenBid
	}
	ma60 /= 60.0

	div := ma60 / ma5

	x := std.lower
	y := std.upper

	std.turningUp   = div >= x && div <= y
	std.turningDown = div >= (1.00 + (1.00 - y)) && div <= (1.00 + (1.00 - x))
}

func (std *SteveTurnDetector) IsTurningUp() bool {
	return std.turningUp
}

func (std *SteveTurnDetector) IsTurningDown() bool {
	return std.turningDown
}

package quotes

import (
	"fmt"
	"strings"

	"../pips"
	"../utils"
)

// ===== QUOTE =====================================================================================

func DifferenceInPips(symbol string, f1, f2 float64) pips.Pip {
	return NewQuote(symbol, f1).ProfitInPipsAt(NewQuote(symbol, f2))
}

func NewQuote(symbol string, value float64) *Quote {
	q := Quote{}
	q.setSymbol(symbol)
	q.setValue(value)

	return &q
}

type Quote struct {
	jpyBase bool
	symbol  string
	value   float64
}

func (q *Quote) AddPips(p pips.Pip) {
	q.changeByPips(p)
}

func (q *Quote) SubtractPips(p pips.Pip) {
	q.changeByPips(pips.Pip(-(float64(p))))
}

func (q *Quote) changeByPips(p pips.Pip) {
	if q.jpyBase {
		q.value += float64(p) / utils.JPY_PIP_DIVISOR
	} else {
		q.value += float64(p) / utils.BASE_PIP_DIVISOR
	}
}

func (q *Quote) ProfitInPipsAt(current *Quote) pips.Pip {
	if q.symbol != current.symbol {
		panic(fmt.Sprintf(
			"can't compare quotes for different symbols (%s vs %s)\n",
			q.symbol,
			current.symbol,
		))
	}

	diff := float64(current.value) - float64(q.value)

	if q.jpyBase {
		return pips.Pip(diff * 100.0)
	} else {
		return pips.Pip(diff * 10000.0)
	}
}

func (q *Quote) setSymbol(symbol string) {
	// TODO: validate symbol actually exists lol
	//       e.g., StringArrayContainsString(symbol, listOfSymbols)
	q.symbol  = symbol
	q.jpyBase = strings.HasSuffix(symbol, "JPY")
}

func (q *Quote) setValue(value float64) {
	utils.EnsureZeroOrGreater(value)
	q.value = value
}

func (q *Quote) ToFloat64() float64 {
	return q.value
}

// ===== TEST CODE =================================================================================

func main() {
	e1 := NewQuote("EURUSD", 1.23450)
	e2 := NewQuote("EURUSD", 1.23470)

	j1 := NewQuote("USDJPY", 123.450)
	j2 := NewQuote("USDJPY", 123.470)

	fmt.Printf("EURUSD profit: %.2f pips\n", e1.ProfitInPipsAt(e2)) // # => 2.0
	fmt.Printf("USDJPY profit: %.2f pips\n", j1.ProfitInPipsAt(j2)) // # => 2.0

	e1Plus := NewQuote("EURUSD", 1.23450)
	e1Plus.AddPips(pips.Pip(1.0))
	fmt.Printf("%.5f + 1.0 pips = %.5f\n", e1.ToFloat64(), e1Plus.ToFloat64())

	j1Plus := NewQuote("USDJPY", 123.450)
	j1Plus.AddPips(pips.Pip(1.0))
	fmt.Printf("%.3f + 1.0 pips = %.3f\n", j1.ToFloat64(), j1Plus.ToFloat64())

	e1Minus := NewQuote("EURUSD", 1.23450)
	e1Minus.SubtractPips(pips.Pip(1.0))
	fmt.Printf("%.5f - 1.0 pips = %.5f\n", e1.ToFloat64(), e1Minus.ToFloat64())

	j1Minus := NewQuote("USDJPY", 123.450)
	j1Minus.SubtractPips(pips.Pip(1.0))
	fmt.Printf("%.3f - 1.0 pips = %.3f\n", j1.ToFloat64(), j1Minus.ToFloat64())
}

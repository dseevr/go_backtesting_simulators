package stops

import (
	"../pips"
)

func NewStopLoss(pips pips.Pip) StopLoss {
	return StopLoss{Set: true, Pips: pips}
}

func NewTakeProfit(pips pips.Pip) TakeProfit {
	return TakeProfit{Set: true, Pips: pips}
}

func NoStopLoss() StopLoss {
	return StopLoss{Set: false}
}

func NoTakeProfit() TakeProfit {
	return TakeProfit{Set: false}
}

// ===== STOPS =====================================================================================

type StopLoss struct {
	Set  bool
	Pips pips.Pip
}

type TakeProfit struct {
	Set  bool
	Pips pips.Pip
}


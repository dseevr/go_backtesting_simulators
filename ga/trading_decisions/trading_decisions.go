package trading_decisions

import (
	// "fmt"
)

type Decision int64

const (
	BUY  = Decision(1)
	NOOP = Decision(0)
	SELL = Decision(-1)
)

// ===== TRADING DECISION ==========================================================================

type TradingDecision interface {
	Run(map[string]interface{}) Decision
}

// ===== STEVEORITHM DECISION ======================================================================

func NewSTD() *SteveTradingDecision {
	return &SteveTradingDecision{}
}

type SteveTradingDecision struct {
}

func (std *SteveTradingDecision) Run(d map[string]interface{}) Decision {
	if d["crossover"].(float64) >= 1.0026 && d["spreadIncreased"].(bool) && d["totalTicks"].(int64) > 100 {
		return BUY
	} else {
		return NOOP
	}
}

package test1

import (
	"math/rand"

	"../../accounts"
	"../../algorithms"
	"../../exchanges"
	"../../ticks"
	"../../variables"
)

type Test1 struct {
	algorithms.Algorithm
}

func (t1 *Test1) Init() {
	t1.CreateFloat("ma_crossover_threshold_lower", 0.991, 0.993)
	t1.CreateFloat("ma_crossover_threshold_upper", 0.997, 0.999)
}

const TICKS_UNTIL_CLOSE = 2000
var ticksSinceOpen = 0

func (t1 *Test1) OnTick(account *accounts.Account, exchange *exchanges.Exchange, tick *ticks.Tick, vars *variables.Variables) {
	for _, trade := range account.OpenTrades() {
		someCriteria := true

		if someCriteria {
			exchange.CloseTrade(account, trade, tick)
		}

		return
	}

	lower := vars.GetFloat("ma_crossover_threshold_lower")
	upper := vars.GetFloat("ma_crossover_threshold_upper")

	currentCrossover := 1.23

	lots := 1.0
	stopLoss := 25.0
	takeProfit := 50.0

	if 0 == rand.Intn(1000) || currentCrossover > lower && currentCrossover < upper {
		if 0 == rand.Intn(2) {
			exchange.OpenLong(account, tick, lots, stopLoss, takeProfit)
		} else {
			exchange.OpenShort(account, tick, lots, stopLoss, takeProfit)
		}
	}
}

package main

import (
	"flag"
	"fmt"
	"time"

	std "../indicators/steve_turn_detector"
	sv  "../indicators/spread_velocity"
	vv  "../indicators/volume_velocity"

	td "../../ga/trading_decisions"

	"../accounts"
	"../algorithms"
	"../exchanges"
	"../indicators"
	"../pips"
	"../quotes"
	"../stops"
	"../ticks"
)

// ===== STEVE'S ALGORITHM #2 ======================================================================

func steveorithm2(algo *algorithms.Algorithm, tick *ticks.MarketTick) {
	if algo.Account.HasOpenOrders() {
		openOrderCount := 0

		for _, order := range algo.Account.OpenOrders() {
			openOrderCount += 1

			if order.Symbol != tick.Symbol {
				continue
			}

			if order.Metadata.Fetch("expiration_time").(time.Time).Before(tick.Time) {
				algo.Broker.CloseOrder(algo.Account, order, tick)
				continue
			}

			if order.PercentToTakeProfit() > 75.0 {
				p := pips.Pip(-(float64(order.GetTakeProfit().Pips) / 1.5))
				order.SetStopLoss(p)
			}
		}

		// if openOrderCount >= 3 {
		// 	return
		// }
	}

	chart := algo.Charts[tick.Symbol]["M1"]
	candles := chart.GetCandles(60)

	// 1. previous minute’s open_bid/(60 minute m1 MA) >=1.0026
	ma60 := 0.0
	for _, candle := range candles {
		ma60 += candle.OpenBid
	}
	ma60 /= 60.0

	lastOpen := candles[0].OpenBid

	crossover := lastOpen / ma60

	// 2. previous minute’s spread has increased since opening (open_ask-open_bid)<(close_ask-close_bid)
	// spreadIncreased := (candles[0].OpenAsk - candles[0].OpenBid) < (candles[0].CloseAsk - candles[0].CloseBid)
	spreadsies := candles[0].OpenBid - candles[0].CloseBid
	spreadIncreased := spreadsies <= 0.0011 && spreadsies >= -0.0011

	// 3. last 5 ticks have more than  100 ticks  (you might be able to use turning up for this)
	totalTicks := int64(0)
	for _, candle := range chart.GetCandles(5) {
		totalTicks += candle.Volume
	}

	_ = spreadIncreased
	_ = totalTicks
	_ = crossover

	data := make(map[string]interface{})
	data["spreadIncreased"] = spreadIncreased
	data["totalTicks"] = totalTicks
	data["crossover"] = crossover

	switch algo.TradingDecision.Run(data) {
	default:
	case td.BUY:
		tpPips := quotes.DifferenceInPips(tick.Symbol, tick.OpenAsk, tick.OpenAsk * 1.004)
		slPips := pips.Pip(float64(tpPips) * 0.90)

		o := algo.Broker.OpenBuyOrder(
			algo.Account,
			tick.Symbol,
			tick,
			// algo.Metadata.Fetch("lots").(float64),
			algo.Account.LotSizeForTrade(slPips),
			stops.NoStopLoss(),
			stops.NoTakeProfit(),
		)
		o.SetStopLoss(slPips)
		o.SetTakeProfit(tpPips)
		o.Metadata.Set("expiration_time", o.OpenedAt.Add(60 * time.Minute))
	case td.SELL:
		fmt.Println("SELL")
	}
}


// ===== PROGRAM ENTRYPOINT ========================================================================

func main() {
	var csvPath string
	var showOrders bool
	var lots float64
	var margin int

	flag.StringVar(&csvPath, "path", "", "path to CSV files")
	flag.Float64Var(&lots, "lots", 0.01, "lots to use")
	flag.IntVar(&margin, "margin", 1, "margin level (default: 1)")
	flag.BoolVar(&showOrders, "show-orders", false, "show order summary after account summary")
	flag.Parse()

	// ===== SETUP =============================================================================

	fmt.Println("Running CSV file:", csvPath)

	e := exchanges.NewWithDeets(&ticks.FXCMM1CsvReader{Path: csvPath})

	// ----- STEVE ALGORITHM 2 v0.0.1 ----------------------------------------------------------

	acc := accounts.NewWithDeets("Steve's Algorithm 2 v0.0.1", 10000.0)
	// acc.SetDrawdownLimit(6.0)
	acc.SetMargin(int64(margin))
	acc.SetMaxRiskPerTrade(1.0)

	if showOrders {
		acc.ShowOrders()
	}

	algo := algorithms.NewWithDeets(acc, steveorithm2)
	algo.TradingDecision = td.NewSTD()
	algo.AddCurrency("EURUSD")//, "AUDUSD")//, "GBPUSD")
	algo.AttachCharts("M1")
	algo.SetStartupDelay(120 * time.Minute) // TODO: remove this
	algo.AddIndicator("STD", func() indicators.Indicator { return std.NewSTD(0.9969, 0.9975) })
	algo.AddIndicator("SV",  func() indicators.Indicator { return sv.NewSV(5)                })
	algo.AddIndicator("VV",  func() indicators.Indicator { return vv.NewVV(5, 0.33)          })
	algo.Metadata.Set("lots", lots)

	e.AddAlgorithm(algo)

	e.Run()
}

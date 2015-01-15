package main

import (
	"flag"
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"

	std "../../exchange_simulator/indicators/steve_turn_detector"
	sv  "../../exchange_simulator/indicators/spread_velocity"
	vv  "../../exchange_simulator/indicators/volume_velocity"

	td "../trading_decisions"

	"../../exchange_simulator/algorithms"
	"../../exchange_simulator/accounts"
	"../../exchange_simulator/exchanges"
	"../../exchange_simulator/indicators"
	"../../exchange_simulator/pips"
	"../../exchange_simulator/quotes"
	"../../exchange_simulator/stops"
	"../../exchange_simulator/ticks"
	"../../exchange_simulator/utils"

	// "../tournament/competitors"
	// "../tournament/best_of"
)

type BestOf struct {
}

func (bo *BestOf) TopN(n int64, comps []*ExchangeCompetitor) []*ExchangeCompetitor {
	numComps := int64(len(comps))

	if numComps <= 0 {
		panic(fmt.Sprintf(
			"must have at least 1 competitor, got %d",
			n,
		))
	}

	if n <= 0 || n > numComps {
		panic(fmt.Sprintf(
			"n must be > 0 and < number of competitors (currently: %d)",
			numComps,
		))
	}

	// TODO: debug why this explodes violently LOL
	var wg sync.WaitGroup
	wg.Add(1)
	// for _, comp := range comps {
	// 	_ = comp
	// 	wg.Add(1)
	// 	go func() {
	// 		comp.Run()
	// 		wg.Done()
	// 	}()
	// }
	wg.Done()
	wg.Wait()

	for _, comp := range comps {
		comp.Run()
	}

	// for _, comp := range comps {
	// 	fmt.Printf("SCORE: %f\n", comp.Score)
	// }

	sort.Sort(ByScore(comps))

	return comps[0:n]
}

// ===== EXCHANGE COMPETITOR =======================================================================

type ByScore []*ExchangeCompetitor

func (a ByScore) Len() int           { return len(a) }
func (a ByScore) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByScore) Less(i, j int) bool { return a[i].Score < a[j].Score }

func NewExchangeCompetitor(e *exchanges.Exchange, a *algorithms.Algorithm) *ExchangeCompetitor {
	return &ExchangeCompetitor{e: e, a: a}
}

type ExchangeCompetitor struct {
	e *exchanges.Exchange
	a *algorithms.Algorithm

	Score float64
}

func (ec *ExchangeCompetitor) Run() {
	fmt.Printf("%#v\n", ec)
	ec.e.Run()
	ec.Score = ec.a.Account.GetBalance()
}

// ===== UTILITY FUNCTIONS =========================================================================

func newAlgo() *algorithms.Algorithm {
	acc := accounts.NewWithDeets("GA", float64(rand.Intn(5000)+10000))//10000.0)
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
	algo.SetStartupDelay(180 * time.Minute) // TODO: remove this
	algo.AddIndicator("STD", func() indicators.Indicator {
		return std.NewSTD(0.9969, 0.9975)
	})
	algo.AddIndicator("SV",  func() indicators.Indicator {
		return sv.NewSV(int64(rand.Intn(15) + 3))
	})
	algo.AddIndicator("VV",  func() indicators.Indicator {
		return vv.NewVV(int64(rand.Intn(7) + 3), 0.33)
	})
	algo.Metadata.Set("lots", lots)

	return algo
}

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

var csvPath string
var showOrders bool
var lots float64
var margin int
var numberOfCompetitors int

func main() {
	flag.StringVar(&csvPath, "path", "", "path to CSV files")
	flag.Float64Var(&lots, "lots", 0.01, "lots to use")
	flag.IntVar(&margin, "margin", 1, "margin level (default: 1)")
	flag.BoolVar(&showOrders, "show-orders", false, "show order summary after account summary")
	flag.IntVar(&numberOfCompetitors, "competitors", 32, "number of competitors (must be power of 2)")
	flag.Parse()

	if numberOfCompetitors < 2 {
		panic("must have at least two competitors")
	}

	// ===== SETUP =============================================================================

	gladiators := []*ExchangeCompetitor{}

	for i := 0; i < numberOfCompetitors; i++ {
		e := exchanges.NewWithDeets(&ticks.FXCMM1CsvReader{Path: csvPath})
		a := newAlgo()
		e.AddAlgorithm(a)

		gladiators = append(gladiators, NewExchangeCompetitor(e, a))
	}

	tourney := BestOf{}
	winners := tourney.TopN(5, gladiators)

	for i, winner := range winners {
		fmt.Printf("%d. %s\n", i + 1, utils.FormatMoney(winner.Score))
	}
}

package algorithms

import (
	"fmt"
	"sync"

	"../accounts"
	"../exchanges"
	"../sample_sets"
	"../ticks"
	"../variables"
)

type BaseAlgorithm interface {
	Init()
	ExecuteOn(*sample_sets.SampleSet) float64
	GetAccount() *accounts.Account
	GetExchange() *exchanges.Exchange
	GetScore() float64
	GetVariables() *variables.Variables
	OnTick(*accounts.Account, *exchanges.Exchange, *ticks.Tick, *variables.Variables)
	ParentInit()
	RandomizeVars()
	SendTick(*ticks.Tick)
	SetBaseAlgorithm(BaseAlgorithm)
	StartReceiver()
	StopReceiver(*sync.WaitGroup)
	TickReceiver()
	VarInit()
}

// ===== ALGORITHM =================================================================================

type Algorithm struct {
	account     *accounts.Account
	exchange    *exchanges.Exchange
	score       float64
	self        BaseAlgorithm
	tickChannel chan *ticks.Tick
	vars        *variables.Variables
}

func (algo *Algorithm) GetAccount() *accounts.Account {
	return algo.account
}

func (algo *Algorithm) GetExchange() *exchanges.Exchange {
	return algo.exchange
}

func (algo *Algorithm) ParentInit() {
	algo.account = accounts.NewAccount(10000.0)
	algo.exchange = exchanges.NewExchange()
}

func (algo *Algorithm) SetBaseAlgorithm(ba BaseAlgorithm) {
	algo.self = ba
}

func (algo *Algorithm) VarInit() {
	algo.vars = &variables.Variables{}
	algo.vars.Init()
}

func (algo *Algorithm) ExecuteOn(ss *sample_sets.SampleSet) float64 {
	fmt.Printf("Executing algorithm on out-of-sample set (%d ticks)\n", ss.Count())
	var wg sync.WaitGroup
	wg.Add(1)

	algo.StartReceiver()
	go algo.TickReceiver()

	first := true
	var lastTick *ticks.Tick

	for tick := range ss.Ticks() {
		if tick.AfterCutoff() {
			continue
		}

		if first {
			first = false
		} else {
			ticks.Validate(tick, lastTick)
		}

		algo.SendTick(tick)

		lastTick = tick
	}

	algo.StopReceiver(&wg)
	wg.Wait()

	algo.GetAccount().PrintSummary()

	return algo.GetScore()
}

func (algo *Algorithm) GetScore() float64 {
	return algo.account.GetBalance()
}

func (algo *Algorithm) RandomizeVars() {
	algo.vars.Randomize()
}

func (algo *Algorithm) SendTick(tick *ticks.Tick) {

	algo.tickChannel <- tick
}

func (algo *Algorithm) StartReceiver() {
	algo.tickChannel = make(chan *ticks.Tick, 1000)
}

func (algo *Algorithm) StopReceiver(wg *sync.WaitGroup) {
	close(algo.tickChannel)
	wg.Done()
}

func (algo *Algorithm) TickReceiver() {
	count := 0

	var lastTick *ticks.Tick

	for {
		tick, ok := <- algo.tickChannel
		if !ok {
			algo.GetExchange().CloseAllTrades(algo.GetAccount(), lastTick)
			break
		}

		algo.GetAccount().RecordTickOnOpenTrades(tick)
		algo.self.OnTick(algo.GetAccount(), algo.GetExchange(), tick, algo.GetVariables())

		count++
		lastTick = tick
	}

	fmt.Printf("Processed %d ticks\n", count)
}

// ===== VAR FUNCTIONS =============================================================================

func (algo *Algorithm) CreateBool(name string) {
	algo.vars.CreateBool(name)
}

func (algo *Algorithm) CreateFloat(name string, lower, upper float64) {
	algo.vars.CreateFloat(name, lower, upper)
}

func (algo *Algorithm) CreateInt(name string, lower, upper int) {
	algo.vars.CreateInt(name, lower, upper)
}

func (algo *Algorithm) GetVariables() *variables.Variables {
	return algo.vars
}

func (algo *Algorithm) GetBool(name string) bool {
	return algo.vars.GetBool(name)
}

func (algo *Algorithm) GetFloat(name string) float64 {
	return algo.vars.GetFloat(name)
}

func (algo *Algorithm) GetInt(name string) int {
	return algo.vars.GetInt(name)
}

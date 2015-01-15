package optimizers

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"../algorithms"
	"../sample_sets"
	"../ticks"
)

// ===== SCORE =====================================================================================

type Score struct {
	algorithm algorithms.BaseAlgorithm
	score     float64
}

type SortableScores []Score

func (s SortableScores) Len() int           { return len(s) }
func (s SortableScores) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s SortableScores) Less(i, j int) bool { return s[i].score < s[j].score }

// ===== OPTIMIZER =================================================================================

type Optimizer struct {
	BaseAlgorithm algorithms.BaseAlgorithm
}

func (opti *Optimizer) AlgorithmFor(ss *sample_sets.SampleSet) algorithms.BaseAlgorithm {
	fmt.Printf("Optimizing on in-sample set (%d ticks)\n", ss.Count())
	algos := []algorithms.BaseAlgorithm{}
	scores := []Score{}

	var wg sync.WaitGroup

	startTime := time.Now()

	for i := 0; i < 1; i++ {
		var algo algorithms.BaseAlgorithm = opti.BaseAlgorithm

		algo.ParentInit()
		algo.SetBaseAlgorithm(algo)
		algo.VarInit()
		algo.Init()
		algo.RandomizeVars()

		algo.StartReceiver()

		wg.Add(1)
		go algo.TickReceiver()

		algos = append(algos, algo)
	}

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

		for _, algo := range algos {
			algo.SendTick(tick)
		}

		lastTick = tick
	}

	for _, algo := range algos {
		algo.StopReceiver(&wg)
	}

	wg.Wait()

	fmt.Printf(
		"Optimized %d algorithms in %.2fs\n",
		len(algos),
		time.Since(startTime).Seconds(),
	)

	for _, algo := range algos {
		scores = append(scores, Score{algorithm: algo, score: algo.GetScore()})
	}

	sort.Sort(SortableScores(scores))

	return scores[len(scores)-1].algorithm
}

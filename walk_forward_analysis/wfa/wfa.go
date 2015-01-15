package main

import (
	"flag"
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"../optimizers"
	"../sample_sets"
	"../utils"

	"../strategies/test1"
)

var ISTime = 3 * utils.Week
var OSTime = 1 * utils.Week

func parseFlags() {

}

func main() {
	runtime.GOMAXPROCS(8)

	seed := 0

	flag.IntVar(&seed, "seed", 0, "custom seed to use")
	flag.Parse()

	if 0 == seed {
		seed = int(time.Now().UTC().UnixNano())
	}

	fmt.Printf("Seed: %d\n", seed)
	rand.Seed(int64(seed))

	// ===== VARIABLE SETUP ====================================================================

	startTime := time.Now()
	// csvfile := "/Volumes/RAMDisk/shorty.csv"
	csvfile := "/Users/bill/src/forex/exchange_simulator/simulator/shorty.csv"

	startOffset := 0 * utils.Week

	var wg sync.WaitGroup

	optimizer := optimizers.Optimizer{
		BaseAlgorithm: &test1.Test1{},
	}

	count := 0
	results := make(map[int]float64)

	// ===== MAIN LOOP =========================================================================

	for {
		in, out, ok := sample_sets.PairFromCSV(csvfile, startOffset, ISTime, OSTime)
		if !ok {
			break
		}

		wg.Add(1)

		func(wg *sync.WaitGroup, count int) {
			results[count] = optimizer.AlgorithmFor(in).ExecuteOn(out)
			wg.Done()
		}(&wg, count)

		startOffset += OSTime
		count++

		// TODO: debugging :D
		break
	}

	wg.Wait()

	// ===== PROCESS RESULTS ===================================================================

	for i := 0; i < count; i++ {
		fmt.Printf("%d. %.1f\n", i + 1, results[i])
	}

	fmt.Printf("Tested %d sample sets in %.2fs\n", count, time.Since(startTime).Seconds())
}

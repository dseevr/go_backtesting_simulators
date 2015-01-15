package fann_file

import (
	"fmt"
	"math/rand"
	"testing"

	"../../go/utils"
)

var filename = "test.data"
var inputNeurons = int64(5)

func TestGeneration(*testing.T) {
	ff := New(filename, inputNeurons, 1)

	for i := int64(1); i <= 5; i++ {
		x := []float64{}

		for j := int64(0); j < inputNeurons; j++ {
			x = append(x, float64(rand.Intn(100)))
		}

		ff.AddPair(x, []float64{utils.Average(x)})
	}

	ff.Finalize()
}

func TestParsing(*testing.T) {
	td := Read(filename)
	fmt.Printf(
		"Inputs: %d, Outputs: %d, Pairs: %d\n",
		td.InputCount(),
		td.OutputCount(),
		td.PairCount(),
	)

	fmt.Println("Training pairs:")

	count := 1

	for pair := range td.Pairs() {
		fmt.Printf("%d. %#v -> %#v\n", count, pair.Input(), pair.Output())
		count++
	}

	td.Close()
}

package main

import (
	"fmt"
	"math"
	"testing"

	"../fann_file"
	"../../go/utils"

	"github.com/white-pony/go-fann"
)

func TestSimple(t *testing.T) {
	ff := fann_file.Read("/Users/bill/src/forex/nn/data/week4.csv.data")

	ann := fann.CreateFromFile("nn1.net")

	numRight := int64(0)
	numWrong := int64(0)
	strongRight := int64(0)
	strongWrong := int64(0)

	avgError := 0.0

	for pair := range ff.Pairs() {
		output := float64(ann.Run(pair.Input())[0])

		expected := float64(pair.Output()[0])

		pc := math.Abs(utils.PercentChange(expected, output)) * 100
		avgError += pc

		fmt.Printf(
			"Expected: %.5f, Got: %.5f (%.5f%%)\n",
			expected,
			output,
			pc,
		)

		if pc <= 0.001 {
			strongRight++
		} else if pc <= 0.005 {
			numRight++
		} else if pc <= 0.01 {
			numWrong++
		} else {
			strongWrong++
		}

		// if output[0] <= 0.5 && 0 == answer || output[0] > 0.5 && 1 == answer {
		// 	numRight++
		// } else {
		// 	numWrong++
		// }

		// if output[0] <= 0.25 && 0 == answer || output[0] >= 0.75 && 1 == answer {
		// 	strongRight++
		// }

		// if output[0] >= 0.75 && 0 == answer || output[0] <= 0.25 && 1 == answer {
		// 	strongWrong++
		// }
	}

	ann.Destroy()

	total := numRight + numWrong + strongWrong + strongRight

	fmt.Println("Total tests:", utils.AddCommas(total))
	fmt.Printf("Average error: %.2f%%\n", avgError / float64(ff.PairCount()))
	fmt.Printf(
		"Very right (<= 0.001%%): %s (%.2f%%)\n",
		utils.AddCommas(strongRight),
		(float64(strongRight) / float64(total)) * 100.0,
	)
	fmt.Printf(
		"Right      (<= 0.005%%): %s (%.2f%%)\n",
		utils.AddCommas(numRight),
		(float64(numRight) / float64(total)) * 100.0,
	)
	fmt.Printf(
		"Wrong      (<= 0.01%%) : %s (%.2f%%)\n",
		utils.AddCommas(numWrong),
		(float64(numWrong) / float64(total)) * 100.0,
	)
	fmt.Printf(
		"Very wrong (> 1.0%%)  : %s (%.2f%%)\n",
		utils.AddCommas(strongWrong),
		(float64(strongWrong) / float64(total)) * 100.0,
	)
}

package main

import (
	"github.com/white-pony/go-fann"
)

func main() {
	const numLayers = 4
	const desiredError = 0.000001
	const maxEpochs = 500000
	const epochsBetweenReports = 5

	ann := fann.CreateStandard(numLayers, []uint32{60, 120, 30, 1})
	// ann.SetTrainingAlgorithm(fann.TRAIN_RPROP)
	// ann.SetTrainingAlgorithm(fann.TRAIN_INCREMENTAL)
	ann.SetActivationFunctionHidden(fann.SIGMOID)
	ann.SetActivationFunctionOutput(fann.SIGMOID)
	// ann.SetActivationFunctionHidden(fann.LINEAR)
	// ann.SetActivationFunctionOutput(fann.LINEAR)
	ann.TrainOnFile("../data/final.data", maxEpochs, epochsBetweenReports, desiredError)
	ann.Save("nn1.net")
	ann.Destroy()
}

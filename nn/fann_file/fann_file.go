package fann_file

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"../../go/utils"

	"github.com/white-pony/go-fann"
)

// ===== TRAINING PAIR =============================================================================

type TrainingPair struct {
	input  []fann.FannType
	output []fann.FannType
}

func (tp *TrainingPair) Input() []fann.FannType {
	return tp.input
}

func (tp *TrainingPair) Output() []fann.FannType {
	return tp.output
}

// ===== FANN FILE =================================================================================

func New(filename string, inputs, outputs int64) *FannFile {
	ff := FannFile{
		inputs: inputs,
		outputs: outputs,
		filename: filename,
		tempfileName: filename + ".tmp",
	}
	ff.PrepareForWrite()

	return &ff
}

func Read(filename string) *FannFile {
	ff := FannFile{
		filename: filename,
	}
	ff.parse()

	return &ff
}

type FannFile struct {
	inputs  int64
	outputs int64

	filename string
	fileHandle *os.File

	tempfileName string

	pairs     []*TrainingPair
	pairCount int64
}

func (ff *FannFile) AddPair(input, output []float64) {
	ff.fileHandle.WriteString(floatArrayToFannString(input))
	ff.fileHandle.WriteString(floatArrayToFannString(output))

	ff.pairCount++
}

func (ff *FannFile) Close() {
	err := ff.fileHandle.Close()
	if err != nil {
		panic("fileHandle.Close(): " + err.Error())
	}
}

func (ff *FannFile) Finalize() {
	_, err := ff.fileHandle.Seek(0, 0) // rewind to beginning
	if err != nil {
		panic("fileHandle.Seek(): " + err.Error())
	}

	f, err := os.Create(ff.filename)
	if err != nil {
		panic("os.Create: " + err.Error())
	}

	f.WriteString(fmt.Sprintf("%d %d %d\n", ff.pairCount, ff.inputs, ff.outputs))
	_, err = io.Copy(f, ff.fileHandle)
	if err != nil {
		panic("io.Copy: " + err.Error())
	}

	f.Close()
	ff.fileHandle.Close()

	os.Remove(ff.tempfileName)
}

func (ff *FannFile) InputCount() int64 {
	return ff.inputs
}

func (ff *FannFile) OutputCount() int64 {
	return ff.outputs
}

func (ff *FannFile) PairCount() int64 {
	return ff.pairCount
}

func (ff *FannFile) Pairs() chan *TrainingPair {
	c := make(chan *TrainingPair, 100)

	go func() {
		for _, pair := range ff.pairs {
			c <- pair
		}

		close(c)
	}()

	return c
}

func (ff *FannFile) parse() {
	f, err := os.Open(ff.filename)
	if err != nil {
		panic("os.Open: " + err.Error())
	}

	ff.fileHandle = f

	scanner := bufio.NewScanner(ff.fileHandle)

	first := true

	for scanner.Scan() {
		if first {
			first = false

			parts := strings.Split(scanner.Text(), " ")

			ff.pairCount = utils.StringToInt(parts[0])
			ff.inputs    = utils.StringToInt(parts[1])
			ff.outputs   = utils.StringToInt(parts[2])
		} else {
			input := fannStringToFannArray(scanner.Text())

			ok := scanner.Scan()
			if !ok {
				panic("expected another row after scanning input")
			}

			output := fannStringToFannArray(scanner.Text())

			if int64(len(input)) != ff.inputs {
				panic(fmt.Sprintf(
					"expected %d inputs, got %d",
					ff.inputs,
					len(input),
				))
			}

			if int64(len(output)) != ff.outputs {
				panic(fmt.Sprintf(
					"expected %d outputs, got %d",
					ff.outputs,
					len(output),
				))
			}

			tp := TrainingPair{
				input: input,
				output: output,
			}

			ff.pairs = append(ff.pairs, &tp)
		}
	}

	if err := scanner.Err(); err != nil {
		panic("scanner.Err(): " + err.Error())
	}

	if ff.pairCount != int64(len(ff.pairs)) {
		panic(fmt.Sprintf(
			"expected %d pairs of input/output, got %d",
			ff.pairCount,
			int64(len(ff.pairs)),
		))
	}
}

func (ff *FannFile) PrepareForWrite() {
	f, err := os.Create(ff.tempfileName)
	if err != nil {
		panic("os.Create: " + err.Error())
	}

	ff.fileHandle = f
}

// ===== UTILITY FUNCTIONS =========================================================================

func floatArrayToFannString(farr []float64) string {
	first := true
	str := ""

	for _, f := range farr {
		output := utils.FloatToString(float64(f))

		if first {
			first = false
			str += output
		} else {
			str += " " + output
		}
	}

	return str + "\n"
}

func fannStringToFannArray(fstr string) []fann.FannType {
	farr := []fann.FannType{}

	for _, val := range strings.Split(fstr, " ") {
		farr = append(farr, fann.FannType(utils.StringToFloat(val)))
	}

	return farr
}

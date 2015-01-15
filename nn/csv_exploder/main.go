package main

import (
	"container/list"
	"encoding/csv"
	"flag"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"../../go/utils"
	"../fann_file"
)


type Tick struct {
	Symbol string    `json:"symbol"`
	Bid    float64   `json:"bid"`
	Ask    float64   `json:"ask"`
	Time   time.Time `json:"time"`
}


func loadCsv(csvPath, currency string) *list.List {
	log.Println("Opening", csvPath)
	t1 := time.Now()

	ticks := list.New()

	csvfile, err := os.Open(csvPath)
	if err != nil {
		log.Fatalln(err)
	}
	defer csvfile.Close()

	reader := csv.NewReader(csvfile)

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatalln(err)
		}

		if currency != record[0] {
			continue
		}

		bid := utils.StringToFloat(record[1])
		ask := utils.StringToFloat(record[2])
		t := utils.ParsePGTime(record[3])

		ticks.PushBack(Tick{Symbol: record[0], Bid: bid, Ask: ask, Time: t})
	}

	log.Printf(
		"loaded %s %s records in %.2fs\n",
		utils.AddCommas(int64(ticks.Len())),
		currency,
		time.Since(t1).Seconds(),
	)

	return ticks
}


func firstNodeAfterTime(node *list.Element, cutoff time.Time) (*list.Element, bool) {
	for ; node != nil; node = node.Next() {
		tick := node.Value.(Tick)

		// log.Printf("Is %v > %v ?\n", tick.Time, cutoff)
		if tick.Time.After(cutoff) {
			return node, true
		}
	}

	return nil, false
}

var csvPaths []string
var path string
var currency string
var output string

func parseFlags() {
	flag.StringVar(&path,  "path", "", "path to CSV file(s) to process")
	flag.StringVar(&currency, "currency", "", "currency filter (must be uppercase)")
	flag.StringVar(&output, "output", "data.data", "destination data file, default \"data.data\"")
	flag.Parse()

	if "" == path || "" == currency {
		panic("must supply --path and --currency")
	}

	csvPaths = strings.Split(path, ",")
}

func normalize(f, top float64) float64 {
	return (top - f) / f
}

func main() {
	parseFlags()

	// ----- VARS ------------------------------------------------------------------------------

	targetValues := int64(60)

	bound     := time.Minute * time.Duration(targetValues - 1)
	lookAhead := time.Minute * 1
	increment := time.Minute * 1

	ff := fann_file.New(output, targetValues, 1)
	runBeginTime := time.Now()

	// ----- EXECUTION -------------------------------------------------------------------------

	for _, csvPath := range csvPaths {
		startTime     := time.Now()
		stopTime      := time.Now()
		lookAheadTime := time.Now()
		nextNodeTime  := time.Now()

		var ok bool

		first := true

		var startNode, lookAheadNode, nextNode *list.Element

		ticks := loadCsv(csvPath, currency)

		for e := ticks.Front(); e != nil; {
			tick := e.Value.(Tick)

			values := list.New()

			if first {
				startTime = tick.Time
				startNode = e
				first = false
			} else {
				startTime = startTime.Add(increment)
				startNode, ok = firstNodeAfterTime(e, startTime)
				if !ok {
					log.Println("Hit end of list looking for start node, bailing")
					break
				}
			}

			values.PushBack(startNode.Value)

			stopTime = startTime.Add(bound)
			nextNodeTime = startTime.Add(increment)
			nextNode = startNode
			e = startNode

			breakout := false

			for {
				if nextNodeTime.After(stopTime) {
					break
				}

				nextNode, ok = firstNodeAfterTime(nextNode, nextNodeTime)
				if !ok {
					log.Println("Hit end of list looking for next node, bailing")
					breakout = true
					break
				}

				values.PushBack(nextNode.Value)

				nextNodeTime = nextNodeTime.Add(increment)
			}

			if breakout {
				break
			} else if targetValues != int64(values.Len()) {
				log.Printf(
					"Expected %d values, got %d... skipping\n",
					targetValues,
					values.Len(),
				)
				continue
			}

			lookAheadTime = nextNodeTime.Add(lookAhead)
			lookAheadNode, ok = firstNodeAfterTime(nextNode, lookAheadTime)
			if !ok {
				log.Printf(
					"Hit end of list looking for lookahead node @ %v, bailing",
					lookAheadTime,
				)
				break
			}

			// ----- RECORD TRAINING PAIR ------------------------------------------------------

			inputs := []float64{}

			for v := values.Front(); v != nil; v = v.Next() {
				inputs = append(inputs, normalize(v.Value.(Tick).Bid, 2.0))
			}

			la := normalize(lookAheadNode.Value.(Tick).Bid, 2.0)
			ff.AddPair(inputs, []float64{la})
		}
	}

	log.Printf(
		"Generated %s pairs of output in %.2fs\n",
		utils.AddCommas(ff.PairCount()),
		time.Since(runBeginTime).Seconds(),
	)

	ff.Finalize()
}

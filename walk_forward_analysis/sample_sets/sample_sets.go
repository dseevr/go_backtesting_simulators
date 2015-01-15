package sample_sets

import (
	"encoding/csv"
	// "fmt"
	"io"
	"os"
	"time"

	"../ticks"
	"../utils"
)

func PairFromCSV(path string, offset, inPeriod, outPeriod time.Duration) (*SampleSet, *SampleSet, bool) {
	inSamples, ok := NewFromCSV(path, offset, inPeriod)
	if !ok {
		return nil, nil, false
	}

	outSamples, ok := NewFromCSV(path, offset + inPeriod, outPeriod)
	if !ok {
		return nil, nil, false
	}

	return inSamples, outSamples, true
}



func NewFromCSV(path string, offset, period time.Duration) (*SampleSet, bool) {
	// TODO: ensure in > out (?)

	csvfile, err := os.Open(path)
	utils.Check(err)
	defer csvfile.Close()
	realcsvfile, err := os.Open(path)
	utils.Check(err)
	// Note: don't close realcsvfile or the CSV reader will explode later

	rows     := csv.NewReader(csvfile)
	realrows := csv.NewReader(realcsvfile)

	foundStart := false
	foundEnd   := false
	var desiredStartTime, desiredStopTime time.Time

	firstRecord := true
	recordsBetweenStartAndEnd := 0
	var firstRecordTime time.Time

	for {
		row, err := rows.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}

		t := utils.MergeDateTime(row[1], row[2])

		if firstRecord {
			firstRecord = false
			firstRecordTime  = t
			desiredStartTime = firstRecordTime.Add(offset)
			desiredStopTime  = firstRecordTime.Add(offset + period)

			continue
		}

		if !foundStart {
			if t.After(desiredStartTime) {
				foundStart = true
				continue
			}

			_, err := realrows.Read()
			utils.Check(err)
		}

		if foundStart {
			recordsBetweenStartAndEnd++
		}

		if !foundEnd && t.After(desiredStopTime) {
			foundEnd = true
			break
		}
	}

	if !foundStart || !foundEnd {
		return nil, false
	} else {
		// fmt.Printf(
		// 	"Returning sample:\n\tStart: %s\n\tStop: %s\n\tRecords: %d\n\n",
		// 	desiredStartTime,
		// 	desiredStopTime,
		// 	recordsBetweenStartAndEnd,
		// )

		ss := SampleSet{
			rows: realrows,
			rowCount: recordsBetweenStartAndEnd,
		}

		return &ss, true
	}
}


type SampleSet struct {
	rows *csv.Reader
	rowCount int
}

func (ss *SampleSet) Count() int {
	return ss.rowCount
}

func (ss *SampleSet) Ticks() chan *ticks.Tick {
	c := make(chan *ticks.Tick, 1000)

	go func() {
		first := true
		count := 0

		for {
			record, err := ss.rows.Read()
			if err == io.EOF {
				break
			} else if err != nil {
				utils.Check(err)
			}

			if first {
				// skip header
				first = false
				continue
			}

			// "Symbol"
			// "Date"
			// "Time"
			// "OpenBid"
			// "HighBid"
			// "LowBid"
			// "CloseBid"
			// "OpenAsk"
			// "HighAsk"
			// "LowAsk"
			// "CloseAsk"
			// "Total Ticks" // aka "volume"

			tick := ticks.Tick{
				Symbol:   record[0],
				Time:     utils.MergeDateTime(record[1], record[2]),
				OpenBid:  utils.StringToFloat(record[3]),
				HighBid:  utils.StringToFloat(record[4]),
				LowBid:   utils.StringToFloat(record[5]),
				CloseBid: utils.StringToFloat(record[6]),
				OpenAsk:  utils.StringToFloat(record[7]),
				HighAsk:  utils.StringToFloat(record[8]),
				LowAsk:   utils.StringToFloat(record[9]),
				CloseAsk: utils.StringToFloat(record[10]),
				Volume:   utils.StringToInt(record[11]),
			}

			c <- &tick

			count++

			if count == ss.rowCount {
				// fmt.Printf("last tick: %s %#v\n", tick.Time.Format(time.RFC3339Nano), tick)
				break
			}
		}

		close(c)
	}()

	return c
}

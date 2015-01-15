package ticks

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"../quotes"
	"../utils"

	"github.com/robinsonstrategy/metastore"
)

type MarketTicker interface {
	Ticks() chan *MarketTick
}

// ===== MARKET TICK ===============================================================================

// a tick representing M1 data from FXCM
// "Date","Time","OpenBid","HighBid","LowBid","CloseBid","OpenAsk","HighAsk","LowAsk","CloseAsk","Total Ticks"
type MarketTick struct {
	Symbol   string
	Time     time.Time
	OpenBid  float64
	HighBid  float64
	LowBid   float64
	CloseBid float64
	OpenAsk  float64
	HighAsk  float64
	LowAsk   float64
	CloseAsk float64
	Volume   int64

	Metadata metastore.Metastore
}

// Note: can legitimately return negative values
func (mt *MarketTick) Spread() float64 {
	return float64(quotes.DifferenceInPips(mt.Symbol, mt.OpenBid, mt.OpenAsk))
}

// ===== FXCM M1 CSV READER ========================================================================

type FXCMM1CsvReader struct {
	Path string
}

func (m1cr *FXCMM1CsvReader) Ticker() MarketTicker {
	return MarketTicker(m1cr)
}

// "Symbol","Date","Time","OpenBid","HighBid","LowBid","CloseBid","OpenAsk","HighAsk","LowAsk","CloseAsk","Total Ticks"
// EURUSD,01/02/2013,04:08:00,1.32617,1.32722,1.32562,1.32708,1.32651,1.32749,1.32597,1.32729,384


func (m1cr *FXCMM1CsvReader) Ticks() chan *MarketTick {
	c := make(chan *MarketTick, 100)

	go func() {
		csvfile, err := os.Open(m1cr.Path)
		if err != nil {
			log.Fatalln(err)
		}

		defer csvfile.Close()

		reader := csv.NewReader(csvfile)

		first := true // used to skip header row in CSV

		for {
			record, err := reader.Read()
			if err == io.EOF {
				break
			} else if err != nil {
				log.Fatalln(err)
			}

			if first {
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

			tick := MarketTick{
				Symbol:   record[0],
				Time:     m1cr.mergeDateTime(record[1], record[2]),
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
		}

		close(c)
	}()

	return c
}

func (m1cr *FXCMM1CsvReader) mergeDateTime(part1, part2 string) time.Time {
	str := part1 + " " + part2

	t, err := time.Parse("2006-01-02 15:04:05", str)
	if err != nil {
		panic(fmt.Sprintf("time.Parse failed on \"%#v\": %s\n", str, err.Error()))
	}

	return t
}

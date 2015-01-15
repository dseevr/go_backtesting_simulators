package ticks

import (
	"fmt"
	"log"
	"math"
	"time"
)

// ===== TICK ======================================================================================

// FXCM switched to an ECN model with lower spreads the week of 2014/10/05
var CUTOFF = time.Date(2014, time.October, 5, 0, 0, 0, 0, time.UTC)

type Tick struct {
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
	Volume   int
}

func (t *Tick) AfterCutoff() bool {
	return t.Time.After(CUTOFF)
}

// ===== MISC FUNCTIONS ============================================================================

const BID_ASK_DEVIATION_TOLERANCE = float64(0.005) // 0.5% (may need adjustment for EOW/SOW)

func Validate(current *Tick, last *Tick) {
	sameDay := current.Time.Day() == last.Time.Day()

	bidDiff := last.OpenBid * BID_ASK_DEVIATION_TOLERANCE
	if !sameDay && (current.OpenBid >= (last.OpenBid + bidDiff) || current.OpenBid <= (last.OpenBid - bidDiff)) {
		fmt.Printf(
			"BAD DATA: bid has deviated over %.1f%% (%.1f%%) from last Bid (curr: %.5f, last: %.5f)\n",
			BID_ASK_DEVIATION_TOLERANCE * 100,
			math.Abs(((current.OpenBid - last.OpenBid) / current.OpenBid) * 100),
			current.OpenBid,
			last.OpenBid,
		)
		fmt.Printf("Last:    %s - %#v\n", last.Time, last)
		fmt.Printf("Current: %s - %#v\n", current.Time, current)
		log.Fatalln("")
	}

	askDiff := last.OpenAsk * BID_ASK_DEVIATION_TOLERANCE
	if !sameDay && (current.OpenAsk >= (last.OpenAsk + askDiff) || current.OpenAsk <= (last.OpenAsk - askDiff)) {
		fmt.Printf(
			"BAD DATA: ask has deviated over %.1f%% (%.1f%%) from last ask (curr: %.5f, last: %.5f)\n",
			BID_ASK_DEVIATION_TOLERANCE * 100,
			math.Abs(((current.OpenAsk - last.OpenAsk) / current.OpenAsk) * 100),
			current.OpenAsk,
			last.OpenAsk,
		)
		fmt.Printf("Last:    %s - %#v\n", last.Time, last)
		fmt.Printf("Current: %s - %#v\n", current.Time, current)
		log.Fatalln("")
	}

	if last.Time.Unix() > current.Time.Unix() {
		log.Fatalf(
			"BAD DATA: current tick is newer than last tick (curr: %#v, last: %#v)\n",
			current.Time,
			last.Time,
		)
	}
}

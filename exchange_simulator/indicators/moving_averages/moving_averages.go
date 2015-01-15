package moving_averages

import (
	"container/list"
	"time"

	"../../ticks"
)

type MAValue struct {
	TimeCreated  time.Time
	Value        float64
}

func (ma *MAValue) OlderThan(cutoff time.Time) bool {
	return ma.TimeCreated.Before(cutoff)
}

// ===== SIMPLE MOVING AVERAGE =====================================================================

func NewTBMA(ttl time.Duration, f func(*ticks.MarketTick) float64) *TimeBoundedMovingAverage {
	return &TimeBoundedMovingAverage{TTL: ttl, Function: f}
}

type TimeBoundedMovingAverage struct {
	Function func(*ticks.MarketTick) float64
	Periods  int64
	TTL      time.Duration
	Values   list.List
}

func (tma *TimeBoundedMovingAverage) Init() {
}

func (tma *TimeBoundedMovingAverage) OnTick(tick *ticks.MarketTick) {
	cutoff := tick.Time.Add(-tma.TTL)

	for e := tma.Values.Front(); e != nil; e = e.Next() {
		maval := e.Value.(MAValue)
		if maval.OlderThan(cutoff) {
			tma.Values.Remove(e)
			continue
		} else {
			break
		}
	}

	tma.Values.PushBack(MAValue{TimeCreated: tick.Time, Value: tma.Function(tick)})
}

func (tma *TimeBoundedMovingAverage) Value() interface{} {
	total := 0.0

	for e := tma.Values.Front(); e != nil; e = e.Next() {
		total += e.Value.(MAValue).Value
	}

	return total / float64(tma.Values.Len())
}


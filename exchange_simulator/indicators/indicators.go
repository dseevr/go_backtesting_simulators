package indicators

import (
	"../ticks"
)

type Indicator interface {
	Init()
	OnTick(interface{}, *ticks.MarketTick)
}

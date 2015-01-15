package close_location_value

import (
	"../../algorithms"
	"../../ticks"
)

// ===== CLOSE LOCATION VALUE ======================================================================

func NewCLV() *CloseLocationValue {
	return &CloseLocationValue
}

type CloseLocationValue struct {
	Values float64
}

func (clv *CloseLocationValue) Init() {
}

func (clv *CloseLocationValue) OnTick(algo interface{}, tick *ticks.MarketTick) {
	// clv.Value = ()
}

func (clv *CloseLocationValue) Value() interface{} {
	return clv.Value
}


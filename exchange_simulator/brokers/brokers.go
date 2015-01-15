package brokers

import (
	"../accounts"
	"../orders"
	"../stops"
	"../ticks"
)

type Broker interface {
	CloseOrder(*accounts.Account, *orders.Order, *ticks.MarketTick)
	CloseAllOrders(*accounts.Account, map[string]*ticks.MarketTick)
	OpenBuyOrder(*accounts.Account, string, *ticks.MarketTick, float64, stops.StopLoss, stops.TakeProfit) *orders.Order
	OpenSellOrder(*accounts.Account, string, *ticks.MarketTick, float64, stops.StopLoss, stops.TakeProfit) *orders.Order
}

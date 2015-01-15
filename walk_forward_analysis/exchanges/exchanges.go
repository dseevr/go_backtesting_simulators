package exchanges

import (
	"../accounts"
	"../ticks"
	"../trades"
)

func NewExchange() *Exchange {
	return &Exchange{}
}

type Exchange struct {
}

func (exchange *Exchange) CloseAllTrades(account *accounts.Account, tick *ticks.Tick) {
	for _, trade := range account.GetTrades() {
		exchange.CloseTrade(account, trade, tick)
	}
}

func (exchange *Exchange) CloseTrade(account *accounts.Account, trade trades.Trade, tick *ticks.Tick) {
	trade.Close(tick)

	// TODO: post-close balance updates and crap like that
}

func (exchange *Exchange) OpenLong(account *accounts.Account, tick *ticks.Tick, lots, stopLoss, takeProfit float64) {
	newTrade := trades.NewLongTrade()
	account.AppendTrade(newTrade)
}

func (exchange *Exchange) OpenShort(account *accounts.Account, tick *ticks.Tick, lots, stopLoss, takeProfit float64) {
	newTrade := trades.NewShortTrade()
	account.AppendTrade(newTrade)
}

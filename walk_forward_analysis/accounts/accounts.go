package accounts

import (
	"fmt"

	"../ticks"
	"../trades"
)

func NewAccount(pips float64) *Account {
	return &Account{
		balance: pips,
		startingBalance: pips,
	}
}

type Account struct {
	balance         float64
	startingBalance float64

	trades []trades.Trade
}

func (account *Account) GetBalance() float64 {
	return account.balance
}

func (account *Account) Profit() float64 {
	profit := 0.0

	for _, trade := range account.ClosedTrades() {
		profit += trade.Profit()
	}

	return profit
}

func (account *Account) RecordTickOnOpenTrades(tick *ticks.Tick) {
	for _, trade := range account.OpenTrades() {
		trade.AppendTick(tick)
	}
}

// ===== TRADES ====================================================================================

func (account *Account) AppendTrade(t trades.Trade) {
	account.trades = append(account.trades, t)
}

func (account *Account) ClosedTrades() []trades.Trade {
	closedTrades := []trades.Trade{}

	for _, trade := range account.trades {
		if trade.IsClosed() {
			closedTrades = append(closedTrades, trade)
		}
	}

	return closedTrades
}

func (account *Account) GetTrades() []trades.Trade {
	return account.trades
}

func (account *Account) OpenTrades() []trades.Trade {
	openTrades := []trades.Trade{}

	for _, trade := range account.trades {
		if trade.IsOpen() {
			openTrades = append(openTrades, trade)
		}
	}

	return openTrades
}

func (account *Account) LongTradeCount() int {
	count := 0

	for _, trade := range account.trades {
		if trade.IsLong() {
			count++
		}
	}

	return count
}

func (account *Account) ShortTradeCount() int {
	return len(account.trades) - account.LongTradeCount()
}

// ===== SUMMARY ===================================================================================

func (account *Account) PrintSummary() {
	fmt.Println("**************** SUMMARY ****************\n")

	profit := account.Profit()

	fmt.Printf(
		"Starting/Ending balance: %.2f/%.2f  Profit: %.2f\n",
		account.startingBalance,
		account.startingBalance + profit,
		profit,
	)
	fmt.Printf(
		"Total trades: %d (long: %d short: %d)\n",
		len(account.trades),
		account.LongTradeCount(),
		account.ShortTradeCount(),
	)

	for _, trade := range account.trades {
		fmt.Printf("%#v\n\n", trade)
	}

	fmt.Println("\n*****************************************\n")
}

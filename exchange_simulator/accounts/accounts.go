package accounts

import (
	"container/list"
	"fmt"
	"math"
	"sort"
	"time"

	"../orders"
	"../pips"
	"../utils"
)

const (
	MINIMUM_DEPOSIT = float64(100.00)
	MINIMUM_EQUITY  = float64(100.00)
	MINIMUM_MARGIN  = float64(100.00)
)

func NewWithDeets(name string, deposit float64) *Account {
	acc := Account{}
	acc.SetName(name)
	acc.SetDeposit(deposit)
	acc.SetMargin(1)

	return &acc
}

// ===== ACCOUNT ===================================================================================

type Account struct {
	name    string

	deposit float64
	marginAvailable float64

	margin  int64

	maxRiskPerTrade float64

	currentBalance float64

	commissionsPaid float64

	highestEquity  float64
	highestBalance float64
	lowestEquity   float64
	lowestBalance  float64

	highestAvailableMargin float64
	lowestAvailableMargin  float64

	marginCalled bool

	lastEquityHigh  float64
	lastEquityLow   float64
	worstDrawdown float64

	drawdownLimit    float64
	drawdownLimitSet bool

	Orders list.List

	showOrders bool
}

func (a *Account) PrintWeeklyStats(startTime, endTime time.Time) {
	currentWeekStart := startTime
	currentWeekEnd   := startTime.Add(1 * utils.Day)

	for {
		if time.Saturday == currentWeekEnd.Weekday() {
			break
		}

		currentWeekEnd = currentWeekEnd.Add(1 * utils.Day)
	}

	first := true
	weekCount := int64(1)

	fmt.Println("******* WEEKLY STATS ********")
	fmt.Println("")

	for {
		if first {
			first = false
		} else {
			currentWeekStart = currentWeekEnd
			currentWeekEnd   = currentWeekStart.Add(7 * utils.Day)
		}

		if currentWeekStart.After(endTime) {
			break
		}

		cwsu := currentWeekStart.Unix()
		cweu := currentWeekEnd.Unix()

		numTrades := int64(0)
		wins := int64(0)
		losses := int64(0)
		totalPips := 0.0
		totalProfit := 0.0

		for e := a.Orders.Front(); e != nil; e = e.Next() {
			order := e.Value.(*orders.Order)
			oau := order.OpenedAt.Unix()

			if oau >= cwsu && oau <= cweu {
				p := order.Profit()

				if p >= 0.0 {
					wins += 1
				} else {
					losses += 1
				}

				numTrades += 1
				totalPips += float64(order.ProfitInPips())
				totalProfit += p
			}
		}

		if numTrades > 0 {
			fmt.Printf(
				"Week %2d (%s): Trades: %2d, W/L: %2d/%2d, Pips: %6.1f, Profit: %s\n",
				weekCount,
				utils.NiceYearMonthFormat(currentWeekEnd),
				numTrades,
				wins,
				losses,
				totalPips,
				utils.FormatMoney(totalProfit),
			)
		}

		weekCount += 1
	}

	fmt.Println("")
}

func (a *Account) WinPercentage() float64 {
	return (float64(a.WinningTradeCount()) / float64(a.Orders.Len())) * 100.0
}

func (a *Account) WinningTradeCount() int64 {
	count := int64(0)

	for e := a.Orders.Front(); e != nil; e = e.Next() {
		if e.Value.(*orders.Order).IsWinner() {
			count += 1
		}
	}

	return count
}

func (a *Account) LosingTradeCount() int64 {
	return int64(a.Orders.Len()) - a.WinningTradeCount()
}

func (a *Account) HasExceededDrawdown() bool {
	a.UpdateDrawdown()

	return a.drawdownLimitSet && a.worstDrawdown <= a.drawdownLimit
}

func (a *Account) GetDrawdownLimit() float64 {
	return a.drawdownLimit
}

func (a *Account) SetDrawdownLimit(dd float64) {
	if dd <= 0.0 {
		panic("drawdown must be a positive value representing how much of the account " +
			"can be at risk at any given time, e.g., 6.0 for 6%%")
	}

	a.drawdownLimitSet = true
	a.drawdownLimit = -(dd)
}

func (a *Account) ShowOrders() {
	a.showOrders = true
}

func (a *Account) GetDrawdown() float64 {
	return a.worstDrawdown
}

func (a *Account) CurrentDrawdown() float64 {
	equity := a.GetEquity()

	if equity > a.lastEquityHigh {
		a.lastEquityHigh = equity
		a.lastEquityLow  = equity
	} else if equity <= a.lastEquityHigh {
		a.lastEquityLow = equity
	}

	return -(100.0 - ((a.lastEquityLow / a.lastEquityHigh) * 100.0))
}

func (a *Account) UpdateDrawdown() {
	current := a.CurrentDrawdown()

	if current < a.worstDrawdown {
		a.worstDrawdown = current
	}
}

func (a *Account) HasOpenOrders() bool {
	return 0 != len(a.OpenOrders())
}

func (a *Account) OpenOrders() []*orders.Order {
	res := []*orders.Order{}

	for e := a.Orders.Front(); e != nil; e = e.Next() {
		o := e.Value.(*orders.Order)

		if o.IsOpen() {
			res = append(res, o)
		}
	}

	return res
}

func (a *Account) TimesHitStopLoss() int64 {
	count := int64(0)

	for e := a.Orders.Front(); e != nil; e = e.Next() {
		o := e.Value.(*orders.Order)

		if o.StopLossHit {
			count += 1
		}
	}

	return count
}

func (a *Account) TimesHitTakeProfit() int64 {
	count := int64(0)

	for e := a.Orders.Front(); e != nil; e = e.Next() {
		o := e.Value.(*orders.Order)

		if o.TakeProfitHit {
			count += 1
		}
	}

	return count
}

func (a *Account) AddOrder(o *orders.Order) {
	a.Orders.PushBack(o)
}

func (a *Account) CanTrade() bool {
	return !a.marginCalled && !a.HasExceededDrawdown()
}

func (a *Account) MarginCalled() {
	a.marginCalled = true
}

func (a *Account) GetCommissionsPaid() float64 {
	return a.commissionsPaid
}

func (a *Account) GetEquity() float64 {
	total := a.currentBalance

	for e := a.Orders.Front(); e != nil; e = e.Next() {
		o := e.Value.(*orders.Order)

		if o.IsClosed() {
			continue
		}

		total += o.Profit()
	}

	return total
}

func (a *Account) PrintSummary() {
	fmt.Println("")
	fmt.Println("Name:", a.name,)
	fmt.Printf(
		"Deposit: %s, Balance: %s, Profit: %s, Commission paid: %s\n",
		utils.FormatMoney(a.deposit),
		utils.FormatMoney(a.GetBalance()),
		utils.FormatMoney(a.GetProfit()),
		utils.FormatMoney(a.GetCommission()),
	)
	fmt.Printf(
		"Pips: %.1f, MC'd: %t, Max consecutive wins/losses: %d/%d\n",
		a.ProfitInPips(),
		!a.CanTrade(),
		a.WinningTradesInARow(),
		a.LosingTradesInARow(),
	)
	fmt.Printf(
		"Trades: %d, Won/Lost: %d/%d (%.2f%%), Hit SL/TP: %d/%d, Worst DD: %.2f%%\n",
		a.Orders.Len(),
		a.WinningTradeCount(),
		a.LosingTradeCount(),
		a.WinPercentage(),
		a.TimesHitStopLoss(),
		a.TimesHitTakeProfit(),
		a.GetDrawdown(),
	)
	fmt.Printf(
		"[Highs/lows] Balance: %s/%s - Equity: %s/%s - Margin: %s/%s (%d:1)\n",
		utils.FormatMoney(a.highestBalance),
		utils.FormatMoney(a.lowestBalance),
		utils.FormatMoney(a.highestEquity),
		utils.FormatMoney(a.lowestEquity),
		utils.FormatMoney(a.highestAvailableMargin),
		utils.FormatMoney(a.lowestAvailableMargin),
		a.margin,
	)
	fmt.Println("")

	fmt.Println("====================== Best 5 Trades ======================")
	for i, o := range a.BestTrades(5) {
		fmt.Printf("#%d:\n", i + 1)
		o.PrintDetails()
	}

	fmt.Println("====================== Worst 5 Trades ======================")
	for i, o := range a.WorstTrades(5) {
		fmt.Printf("#%d:\n", i + 1)
		o.PrintDetails()
	}
	fmt.Println("")

	if a.showOrders || !a.CanTrade() {
		a.PrintOrderSummary()
	}
}

func (a *Account) BestTrades(n int64) []*orders.Order {
	return a.topNTrades(n, true)
}

func (a *Account) WorstTrades(n int64) []*orders.Order {
	return a.topNTrades(n, false)
}

func (a *Account) WinningTradesInARow() int64 {
	return a.tradesInARow(true)
}

func (a *Account) LosingTradesInARow() int64 {
	return a.tradesInARow(false)
}

func (a *Account) tradesInARow(winners bool) int64 {
	max := int64(0)
	current := int64(0)

	for e := a.Orders.Front(); e != nil; e = e.Next() {
		var counts bool

		if winners {
			counts = e.Value.(*orders.Order).IsWinner()
		} else {
			counts = e.Value.(*orders.Order).IsLoser()
		}

		if counts {
			current += 1

			if current > max {
				max = current
			}
		} else {
			current = 0
		}
	}

	return max
}

func (a *Account) topNTrades(n int64, normalSortOrder bool) []*orders.Order {
	fslice := sort.Float64Slice{}

	for e := a.Orders.Front(); e != nil; e = e.Next() {
		order := e.Value.(*orders.Order)
		fslice = append(fslice, order.Profit())
	}

	if normalSortOrder {
		sort.Sort(sort.Reverse(fslice))
	} else {
		fslice.Sort()
	}

	trades := []*orders.Order{}
	count := int64(0)

	for _, profit := range fslice {
		if count >= n {
			break
		}

		for e := a.Orders.Front(); e != nil; e = e.Next() {
			order := e.Value.(*orders.Order)

			if order.Profit() == profit {
				trades = append(trades, order)
				break
			}
		}

		count++
	}

	return trades
}

const TIME_FORMAT = "Mon Jan 2 15:04:05"

func (a *Account) PrintOrderSummary() {
	fmt.Println("=============== TRADES SUMMARY ===============")

	count := 1

	for e := a.Orders.Front(); e != nil; e = e.Next() {
		o := e.Value.(*orders.Order)

		fmt.Printf("Trade %d:\n", count)
		fmt.Printf(
			"Symbol: %s, Duration: %s -> %s (O/C day: %d/%d)\n",
			o.Symbol,
			o.OpenedAt.Format(TIME_FORMAT),
			o.ClosedAt.Format(TIME_FORMAT),
			o.OpenedAt.YearDay(),
			o.ClosedAt.YearDay(),
		)
		fmt.Printf(
			"[%5s] Price O/C: %.5f/%.5f - P/L: %9s - " +
			"Commissions: %7s - Net: %9s\n",
			o.LongOrShort(),
			o.OpenPrice,
			o.ClosePrice,
			utils.FormatMoney(o.Profit()),
			utils.FormatMoney(o.Commission()),
			utils.FormatMoney(o.Profit() - o.Commission()),
		)
		fmt.Printf(
			"SL/TP: %.1f(%1t)/%.1f(%1t), Prices: %.5f/%.5f\n",
			o.GetStopLoss().Pips,
			o.StopLossHit,
			o.GetTakeProfit().Pips,
			o.TakeProfitHit,
			o.StopLossPrice(),
			o.TakeProfitPrice(),
		)
		fmt.Printf(
			"Lots: %.2f - Pip change: (O: %.5f -> C: %.5f) = %.1f pips\n",
			o.LotSize,
			o.OpenPrice,
			o.ClosePrice,
			o.ProfitInPips(),
		)
		fmt.Printf(
			"Bid/Ask Open: %.5f/%.5f - Close: %.5f/%.5f - Lows: %.5f/%.5f - " +
			"Highs: %.5f/%.5f\n" +
			"Balance O/C: %s/%s - Equity O/C: %s/%s\n",
			o.OpenBid, o.OpenAsk,
			o.CloseBid, o.CloseAsk,
			o.LowestBid, o.LowestAsk,
			o.HighestBid, o.HighestAsk,
			utils.FormatMoney(o.BalanceAtOpen), utils.FormatMoney(o.BalanceAtClose),
			utils.FormatMoney(o.EquityAtOpen), utils.FormatMoney(o.EquityAtClose),
		)
		fmt.Printf(
			"Other orders open at open: %s, Close: %s\n",
			utils.AddCommas(o.OrdersOpenAtOpen),
			utils.AddCommas(o.OrdersOpenAtClose),
		)
		fmt.Println("")

		count += 1
	}
	fmt.Println("==============================================\n")
	fmt.Println("")
}

func (a *Account) GetCommission() float64 {
	total := 0.0

	for e := a.Orders.Front(); e != nil; e = e.Next() {
		total += e.Value.(*orders.Order).Commission()
	}

	return total
}

func (a *Account) GetProfit() float64 {
	total := 0.0

	for e := a.Orders.Front(); e != nil; e = e.Next() {
		total += e.Value.(*orders.Order).Profit()
	}

	return total
}

func (a *Account) GetBalance() float64 {
	return a.currentBalance
}

func (a *Account) RealizeProfit(o *orders.Order) {
	a.currentBalance += o.Profit()
	a.currentBalance += o.Commission()
}

func (a *Account) ProfitInPips() pips.Pip {
	total := pips.Pip(0.0)

	for e := a.Orders.Front(); e != nil; e = e.Next() {
		total += e.Value.(*orders.Order).ProfitInPips()
	}

	return total
}

func (a *Account) SetDeposit(deposit float64) {
	if deposit < MINIMUM_DEPOSIT {
		panic(fmt.Sprintf("balance must be >= %.2f", MINIMUM_DEPOSIT))
	}

	a.deposit = deposit
	a.marginAvailable = deposit

	a.highestEquity  = deposit
	a.highestBalance = deposit
	a.lowestEquity   = deposit
	a.lowestBalance  = deposit
	a.currentBalance = deposit

	a.lastEquityHigh  = deposit
	a.lastEquityLow   = deposit
	a.worstDrawdown = 0.0

	a.lowestAvailableMargin  = deposit
	a.highestAvailableMargin = deposit
}

func (a *Account) SetName(name string) {
	a.name = name
}

func (a *Account) UpdateBalance() float64 {
	balance := a.GetBalance()

	if balance > a.highestBalance {
		a.highestBalance = balance
	} else if balance < a.lowestBalance {
		a.lowestBalance = balance
	}

	return balance
}

func (a *Account) UpdateEquity() float64 {
	equity := a.GetEquity()

	if equity > a.highestEquity {
		a.highestEquity = equity
	} else if equity < a.lowestEquity {
		a.lowestEquity = equity
	}

	return equity
}

func (a *Account) UpdateMarginAvailable() float64 {
	m := a.GetBalance()

	if a.margin > 1 {
		for e := a.Orders.Front(); e != nil; e = e.Next() {
			o := e.Value.(*orders.Order)

			if o.IsOpen() {
				m -= a.MarginRequirementPerLot() * o.LotSize

				// TODO: is profit added to available margin?
				m += o.Profit()

				// TODO: is commission subtracted from available margin?

				// TODO: factor in cost of the spread
			}
		}
	}

	a.marginAvailable = m

	if m > a.highestAvailableMargin {
		a.highestAvailableMargin = m
	} else if m < a.lowestAvailableMargin {
		a.lowestAvailableMargin = m
	}

	return m
}

// ===== RISK ======================================================================================

func (a *Account) LotSizeForTrade(p pips.Pip) float64 {
	// calculate available, worst case margin
	wcm := a.GetBalance()

	for e := a.Orders.Front(); e != nil; e = e.Next() {
		o := e.Value.(*orders.Order)

		if o.IsClosed() {
			continue
		}

		mpl := math.Abs(o.MaxPossibleLoss())
		// fmt.Printf("MPL: %s\n", utils.FormatMoney(mpl))

		wcm -= mpl
		wcm -= a.MarginRequirementPerLot() * o.LotSize
	}

	maxRiskableMargin := wcm * (a.maxRiskPerTrade / 100.0)

	// find a lot size for our SL which keeps us under (target percentage * worst case margin)

	lots        := 0.01
	lastLotSize := lots
	goodLots    := lots

	// fmt.Printf(
	// 	"Looking for a lot level below %.2f%% of %s = %s\n",
	// 	a.maxRiskPerTrade,
	// 	utils.FormatMoney(wcm),
	// 	utils.FormatMoney(maxRiskableMargin),
	// )

	for {
		testValue := a.MarginRequirementPerLot() * lots
		testValue *= 1.03

		// fmt.Printf(
		// 	"(%s * %.2f) * 1.03 = %s\n",
		// 	utils.FormatMoney(a.MarginRequirementPerLot()),
		// 	lots,
		// 	utils.FormatMoney(testValue),
		// )

		if testValue >= maxRiskableMargin {
			goodLots = lastLotSize
			// fmt.Printf("Found good lot size: %.2f\n", goodLots)
			break
		}

		lastLotSize = lots
		lots = lots + 0.01
	}

	return goodLots
}

func (a *Account) SetMaxRiskPerTrade(risk float64) {
	if risk <= 0.0 || risk >= 100.0 {
		panic(fmt.Sprintf(
			"max risk must be between 0.0 and 100.0 (got: %.1f)\n",
			risk,
		))
	}

	a.maxRiskPerTrade = risk
}

// ===== MARGIN ====================================================================================

func (a *Account) GetMarginAvailable() float64 {
	return a.marginAvailable
}

func (a *Account) MarginRequirementPerLot() float64 {
	return float64(a.margin) * 100.0
}

func (a *Account) SetMargin(m int64) {
	if m < 1 || m > 50 {
		panic("margin must be between 1 and 50")
	}

	a.margin = m
}


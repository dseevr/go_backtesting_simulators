package trading_decisions

// always returns NOOP... useful for testing things

type NOOPTrader struct {
}

func (nt *NOOPTrader) Run(d map[string]interface{}) Decision {
	return NOOP
}

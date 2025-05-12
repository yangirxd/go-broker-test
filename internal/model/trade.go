package model

import (
	"fmt"
	"regexp"
)

type Trade struct {
	Account string  `json:"account"`
	Symbol  string  `json:"symbol"`
	Volume  float64 `json:"volume"`
	Open    float64 `json:"open"`
	Close   float64 `json:"close"`
	Side    string  `json:"side"`
}

var symbolRegex = regexp.MustCompile(`^[A-Z]{6}$`)

func ValidateTrade(trade Trade) error {
	if trade.Account == "" {
		return fmt.Errorf("account must not be empty")
	}
	if !symbolRegex.MatchString(trade.Symbol) {
		return fmt.Errorf("symbol must match ^[A-Z]{6}$")
	}
	if trade.Volume <= 0 {
		return fmt.Errorf("volume must be greater than 0")
	}
	if trade.Open <= 0 {
		return fmt.Errorf("open must be greater than 0")
	}
	if trade.Close <= 0 {
		return fmt.Errorf("close must be greater than 0")
	}
	if trade.Side != "buy" && trade.Side != "sell" {
		return fmt.Errorf("side must be either 'buy' or 'sell'")
	}
	return nil
}

package model

import (
	"testing"
)

func TestValidateTrade(t *testing.T) {
	tests := []struct {
		name    string
		trade   Trade
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid trade",
			trade: Trade{
				Account: "ACC123",
				Symbol:  "EURUSD",
				Volume:  1.5,
				Open:    1.2345,
				Close:   1.2350,
				Side:    "buy",
			},
			wantErr: false,
			errMsg:  "",
		},
		{
			name: "empty account",
			trade: Trade{
				Account: "",
				Symbol:  "EURUSD",
				Volume:  1.5,
				Open:    1.2345,
				Close:   1.2350,
				Side:    "buy",
			},
			wantErr: true,
			errMsg:  "account must not be empty",
		},
		{
			name: "invalid symbol",
			trade: Trade{
				Account: "ACC123",
				Symbol:  "EUR123",
				Volume:  1.5,
				Open:    1.2345,
				Close:   1.2350,
				Side:    "buy",
			},
			wantErr: true,
			errMsg:  "symbol must match ^[A-Z]{6}$",
		},
		{
			name: "negative volume",
			trade: Trade{
				Account: "ACC123",
				Symbol:  "EURUSD",
				Volume:  -1.5,
				Open:    1.2345,
				Close:   1.2350,
				Side:    "buy",
			},
			wantErr: true,
			errMsg:  "volume must be greater than 0",
		},
		{
			name: "zero volume",
			trade: Trade{
				Account: "ACC123",
				Symbol:  "EURUSD",
				Volume:  0,
				Open:    1.2345,
				Close:   1.2350,
				Side:    "buy",
			},
			wantErr: true,
			errMsg:  "volume must be greater than 0",
		},
		{
			name: "negative open",
			trade: Trade{
				Account: "ACC123",
				Symbol:  "EURUSD",
				Volume:  1.5,
				Open:    -1.2345,
				Close:   -1.2350,
				Side:    "buy",
			},
			wantErr: true,
			errMsg:  "open must be greater than 0",
		},
		{
			name: "zero open",
			trade: Trade{
				Account: "ACC123",
				Symbol:  "EURUSD",
				Volume:  1.5,
				Open:    0,
				Close:   1.2350,
				Side:    "buy",
			},
			wantErr: true,
			errMsg:  "open must be greater than 0",
		},
		{
			name: "negative close",
			trade: Trade{
				Account: "ACC123",
				Symbol:  "EURUSD",
				Volume:  1.5,
				Open:    1.2345,
				Close:   -1.2350,
				Side:    "buy",
			},
			wantErr: true,
			errMsg:  "close must be greater than 0",
		},
		{
			name: "zero close",
			trade: Trade{
				Account: "ACC123",
				Symbol:  "EURUSD",
				Volume:  1.5,
				Open:    1.2345,
				Close:   0,
				Side:    "buy",
			},
			wantErr: true,
			errMsg:  "close must be greater than 0",
		},
		{
			name: "invalid side",
			trade: Trade{
				Account: "ACC123",
				Symbol:  "EURUSD",
				Volume:  1.5,
				Open:    1.2345,
				Close:   1.2350,
				Side:    "hold",
			},
			wantErr: true,
			errMsg:  "side must be either 'buy' or 'sell'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTrade(tt.trade)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTrade() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && err.Error() != tt.errMsg {
				t.Errorf("ValidateTrade() error message = %q, want %q", err.Error(), tt.errMsg)
			}
		})
	}
}

package utils

import (
	"encoding/json"

	"github.com/shopspring/decimal"
)

type MDecimal struct {
	decimal.Decimal
}

func NewFromInt(value int64) MDecimal {
	return MDecimal{
		Decimal: decimal.NewFromInt(value),
	}
}

func NewFromString(value string) MDecimal {
	return MDecimal{
		Decimal: decimal.NewFromString(value),
	}
}

func (d MDecimal) MarshalJSON() ([]byte, error) {
	f, _ := d.Float64()
	return json.Marshal(f)
}

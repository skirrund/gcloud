package utils

import (
	"encoding/json"

	"github.com/shopspring/decimal"
)

type MDecimal struct {
	decimal.Decimal
}

func NewFromInt32(value int32) MDecimal {
	return MDecimal{
		Decimal: decimal.NewFromInt32(value),
	}
}

func NewFromInt(value int64) MDecimal {
	return MDecimal{
		Decimal: decimal.NewFromInt(value),
	}
}

func NewFromFloat32(value float32) MDecimal {
	return MDecimal{
		Decimal: decimal.NewFromFloat32(value),
	}
}

func NewFromFloat(value float64) MDecimal {
	return MDecimal{
		Decimal: decimal.NewFromFloat(value),
	}
}

func NewFromString(value string) (MDecimal, error) {
	mdc, err := decimal.NewFromString(value)
	if err != nil {
		return MDecimal{}, err
	}
	return MDecimal{
		Decimal: mdc,
	}, nil
}

func (d MDecimal) MarshalJSON() ([]byte, error) {
	f, _ := d.Float64()
	return json.Marshal(f)
}

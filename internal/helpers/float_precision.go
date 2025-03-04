package helpers

import "github.com/shopspring/decimal"

func ConvertToCents(amount float64) int64 {
	d := decimal.NewFromFloat(amount)
	d = d.Mul(decimal.NewFromInt(100))
	d = d.Round(0)

	return d.IntPart()
}

func CentsToDollars(amount int64) float64 {
	d := decimal.NewFromInt(amount)
	d = d.Div(decimal.NewFromFloat(100))
	f, _ := d.Float64()
	return f
}

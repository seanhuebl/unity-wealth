package helpers

import (
	"time"

	"github.com/google/uuid"
	"github.com/seanhuebl/unity-wealth/internal/database"
	"github.com/seanhuebl/unity-wealth/internal/models"
	"github.com/shopspring/decimal"
)

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

func MapToTx(id, userID uuid.UUID, date time.Time, merchant string,
	amountCents int64, detailedCategoryID int32,
) models.Tx {
	return models.Tx{
		ID:               id,
		UserID:           userID,
		Date:             date,
		Merchant:         merchant,
		Amount:           CentsToDollars(amountCents),
		DetailedCategory: detailedCategoryID,
	}
}

func AppendTxs(dst []models.Tx, rows []database.TxRow) []models.Tx {
	for _, r := range rows {
		id, userID, txDate, merch, amtCents, catID := r.GetFields()
		dst = append(dst, MapToTx(id, userID, txDate, merch, amtCents, catID))
	}
	return dst
}

func SliceToTxRows[T database.TxRow](in []T) []database.TxRow {
	out := make([]database.TxRow, len(in))
	for i := range in {
		out[i] = in[i]
	}
	return out
}

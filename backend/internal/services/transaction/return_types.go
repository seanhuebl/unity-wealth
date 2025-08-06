package transaction

import "github.com/seanhuebl/unity-wealth/internal/models"

type ListTxResult struct {
	Transactions  []models.Tx
	NextCursor    string
	HasMoreData   bool
	Clamped       bool
	EffectiveSize int32
}

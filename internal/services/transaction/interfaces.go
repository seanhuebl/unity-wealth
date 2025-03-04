package transaction

import (
	"github.com/google/uuid"
	"github.com/seanhuebl/unity-wealth/internal/database"
)

type TxPageRow interface {
	GetUserID() uuid.UUID
	GetTxDate() string
	GetMerchant() string
	GetAmountCents() int64
	GetDetailedCatID() int64
}

func (r database.GetUserTransactionsFirstPageRow) GetUserID() uuid.UUID {

}

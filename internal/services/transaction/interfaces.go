package transaction

import (
	"github.com/google/uuid"
)

type TxPageRow interface {
	GetTxID() uuid.UUID
	GetUserID() uuid.UUID
	GetTxDate() string
	GetMerchant() string
	GetAmountCents() int64
	GetDetailedCatID() int64
}

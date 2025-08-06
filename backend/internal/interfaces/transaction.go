package interfaces

import (
	"time"

	"github.com/google/uuid"
)

type TxPageRow interface {
	GetTxID() uuid.UUID
	GetUserID() uuid.UUID
	GetTxDate() time.Time
	GetMerchant() string
	GetAmountCents() int64
	GetDetailedCatID() int32
}

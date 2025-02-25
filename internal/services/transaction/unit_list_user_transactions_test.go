package transaction

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	dbmocks "github.com/seanhuebl/unity-wealth/internal/mocks/database"
)

func TestListUserTransactions(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name                    string
		userID                  uuid.UUID
		cursorDate              *string
		cursorID                *string
		pageSize                int64
		txSliceLegnth           int64
		getFirstPageErr         error
		getFirstPageErrSubStr   string
		getTxPaginatedErr       error
		getTxPaginatedErrSubStr error
	}{}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockTxQ := dbmocks.NewTransactionQuerier(t)

			var txns []Transaction
			year := 2025
			month := 2
			day := 1
			for i := range tc.txSliceLegnth {
				date := fmt.Sprintf("%v-%v-%v", year, month, day)
				txns[i] = Transaction{
					ID:               uuid.NewString(),
					UserID:           tc.userID.String(),
					Date:             date,
					Merchant:         "costco",
					Amount:           157.44,
					DetailedCategory: 40,
				}
				day += 1
			}

		})
	}
}

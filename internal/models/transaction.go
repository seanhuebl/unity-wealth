package models

import (
	"time"

	"github.com/google/uuid"
)

type NewTxRequest struct {
	Date             string  `json:"date" binding:"required"`
	Merchant         string  `json:"merchant" binding:"required"`
	Amount           float64 `json:"amount" binding:"required"`
	DetailedCategory int32   `json:"detailed_category" binding:"required"`
}

type Tx struct {
	ID               uuid.UUID `json:"id"`
	UserID           uuid.UUID `json:"user_id"`
	Date             time.Time `json:"date" binding:"required"`
	Merchant         string    `json:"merchant" binding:"required"`
	Amount           float64   `json:"amount" binding:"required"`
	DetailedCategory int32     `json:"detailed_category" binding:"required"`
}

type TxResponse struct {
	Date             time.Time `json:"date"`
	Merchant         string    `json:"merchant"`
	Amount           float64   `json:"amount"`
	DetailedCategory int32     `json:"detailed_category"`
}

func NewTransaction(id, userID uuid.UUID, date time.Time, merchant string, amount float64, detailedCategory int32) *Tx {
	return &Tx{
		ID:               id,
		UserID:           userID,
		Date:             date,
		Merchant:         merchant,
		Amount:           amount,
		DetailedCategory: detailedCategory,
	}
}

func ConvertToResponse(txn *Tx) *TxResponse {
	return &TxResponse{
		Date:             txn.Date,
		Merchant:         txn.Merchant,
		Amount:           txn.Amount,
		DetailedCategory: txn.DetailedCategory,
	}
}


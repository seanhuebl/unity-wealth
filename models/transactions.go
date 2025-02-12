package models

type NewTransactionRequest struct {
	Date             string  `json:"date" binding:"required"`
	Merchant         string  `json:"merchant" binding:"required"`
	Amount           float64 `json:"amount" binding:"required"`
	DetailedCategory int64   `json:"detailed_category" binding:"required"`
}

type Transaction struct {
	ID               string  `json:"id"`
	UserID           string  `json:"user_id"`
	Date             string  `json:"date" binding:"required"`
	Merchant         string  `json:"merchant" binding:"required"`
	Amount           float64 `json:"amount" binding:"required"`
	DetailedCategory int64   `json:"detailed_category" binding:"required"`
}

type TransactionResponse struct {
	Date             string  `json:"date"`
	Merchant         string  `json:"merchant"`
	Amount           float64 `json:"amount"`
	DetailedCategory int64   `json:"detailed_category"`
}

func NewTransaction(id, userID, date, merchant string, amount float64, detailedCategory int64) *Transaction {
	return &Transaction{
		ID:               id,
		UserID:           userID,
		Merchant:         merchant,
		Amount:           amount,
		DetailedCategory: detailedCategory,
	}
}

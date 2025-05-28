package models

type NewTxRequest struct {
	Date             string  `json:"date" binding:"required"`
	Merchant         string  `json:"merchant" binding:"required"`
	Amount           float64 `json:"amount" binding:"required"`
	DetailedCategory int64   `json:"detailed_category" binding:"required"`
}

type Tx struct {
	ID               string  `json:"id"`
	UserID           string  `json:"user_id"`
	Date             string  `json:"date" binding:"required"`
	Merchant         string  `json:"merchant" binding:"required"`
	Amount           float64 `json:"amount" binding:"required"`
	DetailedCategory int64   `json:"detailed_category" binding:"required"`
}

type TxResponse struct {
	Date             string  `json:"date"`
	Merchant         string  `json:"merchant"`
	Amount           float64 `json:"amount"`
	DetailedCategory int64   `json:"detailed_category"`
}

func NewTransaction(id, userID, date, merchant string, amount float64, detailedCategory int64) *Tx {
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

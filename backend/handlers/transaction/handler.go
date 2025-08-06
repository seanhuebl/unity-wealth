package transaction

type Handler struct {
	txSvc TransactionService
}

func NewHandler(txSvc TransactionService) *Handler {
	return &Handler{
		txSvc: txSvc,
	}
}

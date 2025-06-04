package auth

type Handler struct {
	authSvc AuthService
}

func NewHandler(authSvc AuthService) *Handler {
	return &Handler{
		authSvc: authSvc,
	}
}

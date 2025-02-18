package auth

import "github.com/seanhuebl/unity-wealth/internal/services/auth"

type Handler struct {
	authSvc *auth.AuthService
}

func NewHandler(authSvc *auth.AuthService) *Handler {
	return &Handler{
		authSvc: authSvc,
	}
}

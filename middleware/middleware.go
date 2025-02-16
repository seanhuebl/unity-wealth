package middleware

import (
	"github.com/seanhuebl/unity-wealth/internal/config"
	"github.com/seanhuebl/unity-wealth/internal/services/auth"
)

type Middleware struct {
	cfg         *config.ApiConfig
	authService *auth.AuthService
}

func NewMiddleware(cfg *config.ApiConfig, authSvc *auth.AuthService) *Middleware {
	return &Middleware{cfg: cfg, authService: authSvc}
}

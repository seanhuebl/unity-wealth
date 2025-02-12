package middleware

import (
	"github.com/seanhuebl/unity-wealth/internal/config"
	"github.com/seanhuebl/unity-wealth/services"
)

type Middleware struct {
	cfg         *config.ApiConfig
	authService *services.AuthService
}

func NewMiddleware(cfg *config.ApiConfig, authSvc *services.AuthService) *Middleware {
	return &Middleware{cfg: cfg, authService: authSvc}
}

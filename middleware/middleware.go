package middleware

import "github.com/seanhuebl/unity-wealth/internal/config"

type Middleware struct {
	cfg *config.ApiConfig
}

func NewMiddleware(cfg *config.ApiConfig) *Middleware {
	return &Middleware{cfg: cfg}
}

package handlers

import "github.com/seanhuebl/unity-wealth/internal/config"

type Handler struct {
	cfg *config.ApiConfig
}

func NewHandler(cfg *config.ApiConfig) *Handler {
	return &Handler{cfg: cfg}
}

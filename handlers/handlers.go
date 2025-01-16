package handlers

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/seanhuebl/unity-wealth/internal/database"
)

type ApiConfig struct {
	Port    string
	Queries Quierier
}

type Quierier interface {
	CreateUser(ctx context.Context, params database.CreateUserParams) error
}

func RegisterRoutes(router *gin.Engine, cfg *ApiConfig) {

	home := router.Group("/")
	{
		home.GET("/health", func(ctx *gin.Context) {
			health(ctx)
		})
	}

	api := router.Group("/api")
	{
		api.POST("/signup", func(ctx *gin.Context) {
			AddUser(ctx, cfg)
		})
	}
}

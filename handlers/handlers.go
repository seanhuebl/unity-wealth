package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/seanhuebl/unity-wealth/internal/database"
)

type ApiConfig struct {
	Port    string
	Queries *database.Queries
}

func RegisterRoutes(router *gin.Engine, cfg *ApiConfig) {

	home := router.Group("/")
	{
		home.GET("/health", func(ctx *gin.Context) {
			health(ctx)
		})
	}

}

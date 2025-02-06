package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/seanhuebl/unity-wealth/internal/config"
)

func RegisterRoutes(router *gin.Engine, cfg *config.ApiConfig) {

	home := router.Group("/")
	{
		home.GET("/health", func(ctx *gin.Context) {
			health(ctx)
		})
		home.POST("/signup", func(ctx *gin.Context) {
			AddUser(ctx, cfg)
		})

		home.POST("/login", func(ctx *gin.Context) {
			Login(ctx, cfg)
		})
	}

	app := router.Group("/app")
	app.Use(UserAuthMiddleware(cfg))
	{

		app.POST("/transactions", func(ctx *gin.Context) {
			NewTransaction(ctx, cfg)
		})

		app.PUT("/transactions/:id", func(ctx *gin.Context) {
			UpdateTransaction(ctx, cfg)
		})
	}
}

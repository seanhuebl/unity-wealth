package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/seanhuebl/unity-wealth/internal/config"
	"github.com/seanhuebl/unity-wealth/middleware"
)

func RegisterRoutes(router *gin.Engine, cfg *config.ApiConfig) {
	h := NewHandler(cfg)
	m := middleware.NewMiddleware(cfg)

	home := router.Group("/")
	{
		home.GET("/health", health)
		home.POST("/signup", h.AddUser)

		home.POST("/login", h.Login)
	}

	app := router.Group("/app")
	app.Use(m.UserAuthMiddleware())
	{

		app.POST("/transactions", func(ctx *gin.Context) {
			NewTransaction(ctx, cfg)
		})

		app.PUT("/transactions/:id", func(ctx *gin.Context) {
			UpdateTransaction(ctx, cfg)
		})
	}

	api := router.Group("/api")
	{
		lookups := api.Group("/lookups")
		{
			categories := lookups.Group("/categories")
			{
				categories.GET("/", h.GetCategories)
				categories.GET("/primary/:id", h.GetPrimaryCategoryByID)
				categories.GET("/detailed/:id", h.GetDetailedCategoryByID)
			}

		}
	}
}

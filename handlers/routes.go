package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/seanhuebl/unity-wealth/internal/config"
	"github.com/seanhuebl/unity-wealth/internal/middleware"
)

func RegisterRoutes(router *gin.Engine, cfg *config.ApiConfig, h *Handlers, m *middleware.Middleware) {

	home := router.Group("/")
	{
		home.POST("/signup", h.userHandler.AddUser)

		home.POST("/login", h.authHandler.Login)

		home.GET("/health", h.commonHandler.Health)

	}

	app := router.Group("/app")
	app.Use(m.UserAuthMiddleware(), m.ClaimsAuthMiddleware())
	{

		app.POST("/transactions", h.txHandler.NewTransaction)
		app.GET("/transactions", h.txHandler.GetTransactionsByUserID)
		app.GET("/transactions/:id", h.txHandler.GetTransactionByID)
		app.PUT("/transactions/:id", h.txHandler.UpdateTransaction)
		app.DELETE("/transactions/:id", h.txHandler.DeleteTransaction)
	}

	api := router.Group("/api")
	{
		lookups := api.Group("/lookups")
		{
			categories := lookups.Group("/categories")
			{
				categories.GET("/", h.catHandler.GetCategories)
				categories.GET("/primary/:id", h.catHandler.GetPrimaryCategoryByID)
				categories.GET("/detailed/:id", h.catHandler.GetDetailedCategoryByID)
			}

		}
	}
}

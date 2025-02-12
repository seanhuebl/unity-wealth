package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/seanhuebl/unity-wealth/internal/config"
	"github.com/seanhuebl/unity-wealth/middleware"
)

func RegisterRoutes(router *gin.Engine, cfg *config.ApiConfig, h *Handler, m *middleware.Middleware) {

	home := router.Group("/")
	{
		home.POST("/signup", h.AddUser)

		home.POST("/login", h.Login)

		home.GET("/health", health)

	}

	app := router.Group("/app")
	app.Use(m.UserAuthMiddleware(), m.ClaimsAuthMiddleware())
	{

		app.POST("/transactions", h.NewTransaction)
		app.GET("/transactions", h.GetTransactionsByUserID)
		app.GET("/transactions/:id", h.GetTransactionByID)
		app.PUT("/transactions/:id", h.UpdateTransaction)
		app.DELETE("/transactions", h.DeleteTransaction)
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

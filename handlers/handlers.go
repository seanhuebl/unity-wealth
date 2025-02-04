package handlers

import (
	"context"
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/seanhuebl/unity-wealth/internal/database"
)

type Quierier interface {
	CreateUser(ctx context.Context, params database.CreateUserParams) error
	GetUserByEmail(ctx context.Context, email string) (database.GetUserByEmailRow, error)
	RevokeToken(ctx context.Context, arg database.RevokeTokenParams) error
	GetDeviceInfoByUser(ctx context.Context, arg database.GetDeviceInfoByUserParams) (string, error)
	CreateRefreshToken(ctx context.Context, arg database.CreateRefreshTokenParams) error
	CreateDeviceInfo(ctx context.Context, arg database.CreateDeviceInfoParams) (string, error)
	WithTx(tx *sql.Tx) *database.Queries
}

func (cfg *ApiConfig) RegisterRoutes(router *gin.Engine) {

	home := router.Group("/")
	{
		home.GET("/health", func(ctx *gin.Context) {
			health(ctx)
		})
		home.POST("/signup", func(ctx *gin.Context) {
			cfg.AddUser(ctx)
		})
	
		home.POST("/login", func(ctx *gin.Context) {
			cfg.Login(ctx)
		})
	}

	app := router.Group("/app")
	app.Use(cfg.UserAuthMiddleware())
	{
		app.POST("/transactions", func(ctx *gin.Context) {
			//cfg.NewTransaction(ctx)
		})
	}
}

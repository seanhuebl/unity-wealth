package handlers

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/seanhuebl/unity-wealth/internal/database"
)

type ApiConfig struct {
	Port        string
	Queries     Quierier
	TokenSecret string
}

type Quierier interface {
	CreateUser(ctx context.Context, params database.CreateUserParams) error
	GetUserByEmail(ctx context.Context, email string) (database.GetUserByEmailRow, error)
	RevokeToken(ctx context.Context, arg database.RevokeTokenParams) error
	GetDeviceInfoByUser(ctx context.Context, arg database.GetDeviceInfoByUserParams) (interface{}, error)
	CreateRefreshToken(ctx context.Context, arg database.CreateRefreshTokenParams) error
	CreateDeviceInfo(ctx context.Context, arg database.CreateDeviceInfoParams) (interface{}, error)
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

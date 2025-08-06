package server

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/seanhuebl/unity-wealth/internal/config"
	"github.com/seanhuebl/unity-wealth/internal/middleware"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewRouter(
	cfg *config.ApiConfig,
	h *HandlersGroup,
	m *middleware.Middleware,
	logger *zap.Logger,
) *gin.Engine {

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		start := time.Now()
		c.Next()

		logger.Info("http_request",
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Duration("latency_ms", time.Since(start)),
			zap.String("client_ip", c.ClientIP()),
		)
	})

	r.Use(gin.CustomRecoveryWithWriter(
		zapcore.AddSync(os.Stderr),
		func(c *gin.Context, rec interface{}) {
			logger.Error("panic_recovered",
				zap.Any("error", rec),
				zap.String("path", c.Request.URL.Path),
			)
			c.AbortWithStatus(http.StatusInternalServerError)
		},
	))

	r.Use(m.RequestID())

	registerPublicRoutes(r, h)
	registerAppRoutes(r, h, m)
	registerLookupRoutes(r, h)

	return r
}

// helpers

func registerPublicRoutes(r *gin.Engine, h *HandlersGroup) {
	public := r.Group("/")
	public.POST("signup", h.User.SignUp)
	public.POST("login", h.Auth.Login)
	public.GET("health", h.Cmn.Health)
}

func registerAppRoutes(r *gin.Engine, h *HandlersGroup, m *middleware.Middleware) {
	app := r.Group("/app")
	app.Use(m.UserAuthMiddleware(), m.ClaimsAuthMiddleware(), m.RequestID())

	app.POST("transactions", h.Tx.NewTransaction)
	app.GET("transactions", m.Paginate(), h.Tx.GetTransactionsByUserID)
	app.GET("transactions/:id", h.Tx.GetTransactionByID)
	app.POST("transactions/:id", h.Tx.UpdateTransaction) // I want full transaction update to be re-written not partial
	app.DELETE("transactions/:id", h.Tx.DeleteTransaction)

}

func registerLookupRoutes(r *gin.Engine, h *HandlersGroup) {
	cat := r.Group("/api/lookups/categories")
	{
		cat.GET("", h.Cat.GetCategories)
		cat.GET("primary/:id", h.Cat.GetPrimaryCategoryByID)
		cat.GET("detailed/:id", h.Cat.GetDetailedCategoryByID)
	}
}

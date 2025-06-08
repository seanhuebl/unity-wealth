package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/seanhuebl/unity-wealth/cache"
	authHandler "github.com/seanhuebl/unity-wealth/handlers/auth"
	"github.com/seanhuebl/unity-wealth/handlers/category"
	"github.com/seanhuebl/unity-wealth/handlers/common"
	txHandler "github.com/seanhuebl/unity-wealth/handlers/transaction"
	userHandler "github.com/seanhuebl/unity-wealth/handlers/user"
	"github.com/seanhuebl/unity-wealth/internal/config"
	"github.com/seanhuebl/unity-wealth/internal/database"
	"github.com/seanhuebl/unity-wealth/internal/middleware"
	"github.com/seanhuebl/unity-wealth/internal/models"
	"github.com/seanhuebl/unity-wealth/internal/services/auth"
	"github.com/seanhuebl/unity-wealth/internal/services/transaction"
	userService "github.com/seanhuebl/unity-wealth/internal/services/user"
	"github.com/seanhuebl/unity-wealth/logger"
	"github.com/seanhuebl/unity-wealth/server"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
	"go.uber.org/zap"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("unable to load environment:", err)
	}

	appLogger, err := logger.InitLogger()
	if err != nil {
		log.Fatalf("failed to initialize zap logger: %v", err)
	}
	defer appLogger.Sync()

	db, err := sql.Open("libsql", os.Getenv("DATABASE_URL"))
	if err != nil {
		appLogger.Fatal("unable to connect to database", zap.Error(err))
	}
	if err := db.Ping(); err != nil {
		appLogger.Fatal("database connection test failed", zap.Error(err))
	}

	cfg := &config.ApiConfig{
		Port:     fmt.Sprintf(":%v", os.Getenv("PORT")),
		Queries:  database.New(db),
		Database: db,
	}

	if err := cache.WarmCategoriesCache(cfg); err != nil {
		appLogger.Warn("unable to warm cache", zap.Error(err))
	}

	pwdHasher := auth.NewRealPwdHasher()
	tokenGen := auth.NewRealTokenGenerator(os.Getenv("TOKEN_SECRET"), models.TokenType(os.Getenv("TOKEN_TYPE")))
	tokenExtract := auth.NewRealTokenExtractor()

	transactionalQ := database.NewRealTransactionalQuerier(cfg.Queries)

	sqlTxQ := database.NewRealSqlTxQuerier(transactionalQ)
	txQ := database.NewRealTransactionQuerier(transactionalQ)
	userQ := database.NewRealUserQuerier(transactionalQ)

	authSvc := auth.NewAuthService(sqlTxQ, userQ, tokenGen, tokenExtract, pwdHasher, appLogger)
	txnSvc := transaction.NewTransactionService(txQ, appLogger)
	userSvc := userService.NewUserService(cfg.Queries, pwdHasher, appLogger)

	authHandler := authHandler.NewHandler(authSvc)
	catHandler := category.NewHandler()
	commonHandler := common.NewHandler()
	txHandler := txHandler.NewHandler(txnSvc)
	userHandler := userHandler.NewHandler(userSvc)

	h := server.NewHandlers(
		authHandler,
		catHandler,
		commonHandler,
		txHandler,
		userHandler,
	)
	m := middleware.NewMiddleware(tokenGen, tokenExtract)

	router := server.NewRouter(cfg, h, m, appLogger)

	appLogger.Info("starting server", zap.String("port", cfg.Port))
	err = router.Run(cfg.Port)
	if err != nil {
		appLogger.Fatal("error starting server", zap.Error(err))
	}

}

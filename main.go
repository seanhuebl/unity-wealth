package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/seanhuebl/unity-wealth/cache"
	"github.com/seanhuebl/unity-wealth/handlers"
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
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("unable to load environment:", err)
	}
	db, err := sql.Open("libsql", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("unable to connect to database: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("database connection test failed: %v", err)
	}

	cfg := config.ApiConfig{
		Port:        fmt.Sprintf(":%v", os.Getenv("PORT")),
		Queries:     database.New(db),
		TokenSecret: os.Getenv("TOKEN_SECRET"),
		Database:    db,
	}

	tokenGen := auth.NewRealTokenGenerator(cfg.TokenSecret, models.TokenType(os.Getenv("TOKEN_TYPE")))
	tokenExtract := auth.NewRealTokenExtractor()
	pwdHasher := auth.NewRealPwdHasher()
	transactionalQ := database.NewRealTransactionalQuerier(cfg.Queries)
	sqlTxQ := database.NewRealSqlTxQuerier(transactionalQ)
	//tokenQ := database.NewRealTokenQuerier(cfg.Queries)
	userQ := database.NewRealUserQuerier(transactionalQ)

	authSvc := auth.NewAuthService(sqlTxQ, userQ, tokenGen, tokenExtract, pwdHasher)

	userSvc := userService.NewUserService(cfg.Queries, pwdHasher)

	if err := cache.WarmCategoriesCache(&cfg); err != nil {
		log.Printf("unable to warm cache: %v", err)
	}
	router := gin.Default()
	txQ := database.NewRealTransactionQuerier(transactionalQ)
	txnSvc := transaction.NewTransactionService(txQ)

	// Initialize handlers
	userHandler := userHandler.NewHandler(userSvc)
	catHandler := category.NewHandler()
	authHandler := authHandler.NewHandler(authSvc)
	txHandler := txHandler.NewHandler(txnSvc)
	commonHandler := common.NewHandler()
	h := handlers.NewHandlers(
		authHandler,
		catHandler,
		commonHandler,
		txHandler,
		userHandler,
	)
	m := middleware.NewMiddleware(tokenGen, tokenExtract)
	handlers.RegisterRoutes(router, &cfg, h, m)

	err = router.Run(cfg.Port)
	if err != nil {
		log.Fatal("error starting server:", err)
	}

}

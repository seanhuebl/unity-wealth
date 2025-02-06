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
	"github.com/seanhuebl/unity-wealth/internal/auth"
	"github.com/seanhuebl/unity-wealth/internal/config"
	"github.com/seanhuebl/unity-wealth/internal/database"
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
		Auth:        auth.NewAuthService(),
	}
	if err := cache.WarmCategoriesCache(&cfg); err != nil {
		log.Printf("unable to warm cache: %v", err)
	}

	router := gin.Default()

	handlers.RegisterRoutes(router, &cfg)

	err = router.Run(cfg.Port)
	if err != nil {
		log.Fatal("error starting server:", err)
	}

}

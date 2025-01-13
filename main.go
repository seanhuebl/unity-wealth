package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/seanhuebl/unity-wealth/handlers"
)

type apiConfig struct {
	port         string
	secretString string
}

func main() {
	godotenv.Load()
	var cfg apiConfig
	cfg.port = fmt.Sprintf(":%v", os.Getenv("PORT"))
	cfg.secretString = os.Getenv("JWT_SECRET")

	router := gin.Default()

	router.GET("/health", handlers.Health)

	err := router.Run(cfg.port)
	if err != nil {
		log.Fatal("error starting server")
	}

}

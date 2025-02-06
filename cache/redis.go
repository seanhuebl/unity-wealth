package cache

import (
	"context"
	"log"

	"github.com/go-redis/redis/v8"
	"github.com/seanhuebl/unity-wealth/handlers"
)

var RedisClient = redis.NewClient(&redis.Options{
	Addr:     "localhost:6379",
	Password: "", // No password set
	DB:       0,  // Use default DB
})

func WarmCategoriesCache(cfg *handlers.ApiConfig) error {
	ctx := context.Background()

	primaryCategories, err := cfg.Queries.GetPrimaryCategories(ctx)
	if err != nil {
		log.Printf("error getting primary categories: %v", err)
		return err
	}

	detailedCategories, err := cfg.Queries.GetDetailedCategories(ctx)
	if err != nil {
		log.Printf("error getting detailed categories: %v", err)
		return err
	}

	if err := RedisClient.Set(ctx, "primary_categories", primaryCategories, 0).Err(); err != nil {
		log.Printf("error caching primary categories: %v", err)
		return err
	}
	if err := RedisClient.Set(ctx, "detailed_categories", detailedCategories, 0).Err(); err != nil {
		log.Printf("error caching detailed categories: %v", err)
		return err
	}

	log.Println("categories cached sucessfully")
	return nil
}

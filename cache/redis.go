package cache

import (
	"context"
	"encoding/json"
	"log"
	"strconv"

	"github.com/go-redis/redis/v8"
	"github.com/seanhuebl/unity-wealth/internal/config"
	"github.com/seanhuebl/unity-wealth/internal/database"
)

var RedisClient = redis.NewClient(&redis.Options{
	Addr:     "localhost:6379",
	Password: "", // No password set
	DB:       0,  // Use default DB
})

func WarmCategoriesCache(cfg *config.ApiConfig) error {
	ctx := context.Background()

	primaryCategories, err := cfg.Queries.GetPrimaryCategories(ctx)
	if err != nil {
		log.Printf("error getting primary_categories: %v", err)
		return err
	}

	detailedCategories, err := cfg.Queries.GetDetailedCategories(ctx)
	if err != nil {
		log.Printf("error getting detailed_categories: %v", err)
		return err
	}

	if err = storeCategoriesAsHash(ctx, RedisClient, "primary_categories", primaryCategories, func(p database.PrimaryCategory) int64 {
		return p.ID
	}); err != nil {
		log.Printf("error hashing primary_categories into the cache: %v", err)
		return err
	}

	if err = storeCategoriesAsHash(ctx, RedisClient, "detailed_categories", detailedCategories, func(d database.DetailedCategory) int64 {
		return d.ID
	}); err != nil {
		log.Printf("error hashing detailed_categories into the cache: %v", err)
		return err
	}

	log.Println("categories cached successfully")
	return nil
}

func storeCategoriesAsHash[T any](ctx context.Context, client *redis.Client, keyName string, categories []T, idExtractor func(T) int64) error {
	redisHash := make(map[string]interface{}, len(categories))
	for _, cat := range categories {
		catJSON, err := json.Marshal(cat)
		if err != nil {
			return err
		}
		fieldName := strconv.FormatInt(idExtractor(cat), 10)
		redisHash[fieldName] = string(catJSON)
	}
	return client.HSet(ctx, keyName, redisHash).Err()
}

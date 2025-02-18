package category

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/seanhuebl/unity-wealth/cache"
)

func (h *Handler) GetCategories(ctx *gin.Context) {
	primaryHash, err := cache.RedisClient.HGetAll(ctx.Request.Context(), "primary_categories").Result()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "unable to load primary_categories",
		})
		return
	}
	detailedHash, err := cache.RedisClient.HGetAll(ctx.Request.Context(), "detailed_categories").Result()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "unable to load detailed_categories",
		})
		return
	}

	response := CategoriesResponse{
		PrimaryCategories:  primaryHash,
		DetailedCategories: detailedHash,
	}

	ctx.JSON(http.StatusOK, response)
}

func (h *Handler) GetPrimaryCategoryByID(ctx *gin.Context) {
	id := ctx.Param("id")
	primaryCategory, err := cache.RedisClient.HGet(ctx.Request.Context(), "primary_categories", id).Result()
	if err != nil {
		if err == redis.Nil {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": "primary category not found",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "unable to load primary category",
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"primary_category": primaryCategory,
	})
}

func (h *Handler) GetDetailedCategoryByID(ctx *gin.Context) {
	id := ctx.Param("id")
	detailedCategory, err := cache.RedisClient.HGet(ctx.Request.Context(), "detailed_categories", id).Result()
	if err != nil {
		if err == redis.Nil {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": "detailed category not found",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "unable to load detailed category",
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"detailed_category": detailedCategory,
	})
}

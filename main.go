package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

const (
	defaultRedisHost     = "redis:6379"
	defaultRedisPassword = ""
	defaultPort          = "10000"
)

func main() {

	ctx := context.Background()

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "redis:6379"
	}

	redisPassword := os.Getenv("REDIS_HOST")
	if redisAddr == "" {
		redisPassword = defaultRedisPassword
	}

	// Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       0, // use default DB
	})

	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": time.Now().Format(time.RFC3339),
		})
	})
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.GET("/redis", func(c *gin.Context) {
		const visitsKey = "visits"

		// Get current visit count
		visits, err := func() (int, error) {
			value, err := rdb.Get(ctx, visitsKey).Result()
			if err != nil {
				if err == redis.Nil {
					return 0, nil
				}
				return 0, err
			}
			// Parse visit count
			return strconv.Atoi(value)
		}()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			return
		}

		// Increase visit count
		visits++

		// Store new visits count
		err = rdb.Set(ctx, visitsKey, visits, 0).Err()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			return
		}

		// Response
		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("Hello ! View count: %d", visits),
		})
	})

	port := os.Getenv("API_PORT")
	if port == "" {
		port = defaultPort
	}
	err := r.Run(fmt.Sprintf("0.0.0.0:%s", port))
	if err != nil {
		panic(err)
	}
}

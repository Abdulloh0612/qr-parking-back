package db

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var redisClient *redis.Client

// GetRedisInstance creates a Redis singleton.
// If an instance was already created, it is returned immediately.
func GetRedisInstance(addr, password, dbStr string) *redis.Client {
	if redisClient != nil {
		return redisClient
	}

	dbNum, err := strconv.Atoi(dbStr)
	if err != nil || dbNum < 0 {
		dbNum = 0
	}

	start := time.Now()

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       dbNum,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Could not connect to Redis: %v", err)
	}

	log.Printf("Redis connection established -- %dms", time.Since(start).Milliseconds())
	redisClient = client
	return redisClient
}

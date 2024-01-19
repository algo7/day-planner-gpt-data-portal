package redisclient

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

var (
	// Rdb is the redis client
	Rdb = redis.NewClient(&redis.Options{
		Addr:     getRedisHostAddress(),   // use the given Addr or default Addr
		Password: os.Getenv("REDIS_PASS"), // no password set
		DB:       0,                       // use default DB
	})
)

// getRedisHostAddress checks if custom redis host address is supplied, if not, returns the default address
func getRedisHostAddress() string {

	// Get the redis host address from the environment variable
	redisHost := os.Getenv("REDIS_HOST")

	// If the environment variable is not set, use the default address
	if redisHost == "" {
		redisHost = "localhost:6379"
	}

	return redisHost
}

// RedisConnectionCheck checks if the redis server is up and running
func RedisConnectionCheck() error {

	ctx := context.Background()

	// Ping the redis server to check if it is up
	resp, err := Rdb.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("redis server is not running: %w", err)
	}

	log.Println("Redis connection established", resp)

	return nil
}

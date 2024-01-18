package redisclient

import (
	"os"

	"github.com/redis/go-redis/v9"
)

var (
	rdb = redis.NewClient(&redis.Options{
		Addr:     getRedisHostAddress(), // use the given Addr or default Addr
		Password: "",                    // no password set
		DB:       0,                     // use default DB
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

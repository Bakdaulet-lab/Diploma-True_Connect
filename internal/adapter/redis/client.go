package redisadapter

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// New creates a new Redis client and verifies connectivity.
func New(ctx context.Context, addr, password string, db int) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("pinging redis: %w", err)
	}

	return client, nil
}

package redisadapter

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/trueconnect/backend/internal/repository"
)

// MatchingCache implements repository.MatchingCache using Redis.
type MatchingCache struct {
	client *redis.Client
}

var _ repository.MatchingCache = (*MatchingCache)(nil)

// NewMatchingCache creates a new Redis-backed matching cache.
func NewMatchingCache(client *redis.Client) *MatchingCache {
	return &MatchingCache{client: client}
}

// AddSeen adds user IDs to the requester's "already seen" set.
func (m *MatchingCache) AddSeen(ctx context.Context, requesterID uuid.UUID, seenIDs []uuid.UUID, ttl time.Duration) error {
	if len(seenIDs) == 0 {
		return nil
	}

	key := fmt.Sprintf("seen:%s", requesterID.String())

	members := make([]any, len(seenIDs))
	for i, id := range seenIDs {
		members[i] = id.String()
	}

	pipe := m.client.Pipeline()
	pipe.SAdd(ctx, key, members...)
	pipe.Expire(ctx, key, ttl)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("adding seen IDs: %w", err)
	}

	return nil
}

// GetSeenIDs returns all user IDs already shown to the requester.
func (m *MatchingCache) GetSeenIDs(ctx context.Context, requesterID uuid.UUID) ([]uuid.UUID, error) {
	key := fmt.Sprintf("seen:%s", requesterID.String())

	members, err := m.client.SMembers(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("getting seen IDs: %w", err)
	}

	ids := make([]uuid.UUID, 0, len(members))
	for _, s := range members {
		id, err := uuid.Parse(s)
		if err != nil {
			continue // skip malformed entries
		}
		ids = append(ids, id)
	}

	return ids, nil
}

// CacheTrustScore stores a user's computed trust score in Redis.
func (m *MatchingCache) CacheTrustScore(ctx context.Context, userID uuid.UUID, score int, ttl time.Duration) error {
	key := fmt.Sprintf("trust:%s", userID.String())

	err := m.client.Set(ctx, key, score, ttl).Err()
	if err != nil {
		return fmt.Errorf("caching trust score: %w", err)
	}

	return nil
}

// GetCachedTrustScore retrieves a cached trust score. Returns -1 when not cached.
func (m *MatchingCache) GetCachedTrustScore(ctx context.Context, userID uuid.UUID) (int, error) {
	key := fmt.Sprintf("trust:%s", userID.String())

	val, err := m.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return -1, nil
	}
	if err != nil {
		return -1, fmt.Errorf("getting cached trust score: %w", err)
	}

	score, err := strconv.Atoi(val)
	if err != nil {
		return -1, fmt.Errorf("parsing cached trust score: %w", err)
	}

	return score, nil
}

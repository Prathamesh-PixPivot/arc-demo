package licensing

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type UsageTracker interface {
	IncrementAPIRequest(tenantID string) (int64, error)
	GetAPIRequestCount(tenantID string) (int64, error)
}

type RedisUsageTracker struct {
	client *redis.Client
}

func NewRedisUsageTracker(client *redis.Client) *RedisUsageTracker {
	return &RedisUsageTracker{client: client}
}

func (t *RedisUsageTracker) IncrementAPIRequest(tenantID string) (int64, error) {
	key := fmt.Sprintf("usage:api:%s:%s", tenantID, time.Now().Format("2006-01"))
	count, err := t.client.Incr(context.Background(), key).Result()
	if err != nil {
		return 0, err
	}

	// Set expiry to 40 days (to cover the month + buffer)
	if count == 1 {
		t.client.Expire(context.Background(), key, 40*24*time.Hour)
	}

	return count, nil
}

func (t *RedisUsageTracker) GetAPIRequestCount(tenantID string) (int64, error) {
	key := fmt.Sprintf("usage:api:%s:%s", tenantID, time.Now().Format("2006-01"))
	val, err := t.client.Get(context.Background(), key).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return val, err
}

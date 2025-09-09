package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/google/uuid"
)

type RankingsCache struct {
	client *redis.Client
	ttl    time.Duration
}

type RankingResult struct {
	VideoID uuid.UUID `json:"video_id"`
	Votes   int64     `json:"votes"`
}

func NewRankingsCache(client *redis.Client, ttl time.Duration) *RankingsCache {
	return &RankingsCache{
		client: client,
		ttl:    ttl,
	}
}

func (c *RankingsCache) GetRankings(ctx context.Context, limit int, city string) ([]RankingResult, bool) {
	key := c.buildKey(limit, city)
	
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		return nil, false
	}
	
	var rankings []RankingResult
	if err := json.Unmarshal([]byte(val), &rankings); err != nil {
		return nil, false
	}
	
	return rankings, true
}

func (c *RankingsCache) SetRankings(ctx context.Context, limit int, city string, rankings []RankingResult) error {
	key := c.buildKey(limit, city)
	
	data, err := json.Marshal(rankings)
	if err != nil {
		return err
	}
	
	return c.client.Set(ctx, key, data, c.ttl).Err()
}

func (c *RankingsCache) InvalidateAll(ctx context.Context) error {
	pattern := "rankings:*"
	keys, err := c.client.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}
	
	if len(keys) > 0 {
		return c.client.Del(ctx, keys...).Err()
	}
	
	return nil
}

func (c *RankingsCache) buildKey(limit int, city string) string {
	if city == "" {
		return fmt.Sprintf("rankings:limit:%d", limit)
	}
	return fmt.Sprintf("rankings:limit:%d:city:%s", limit, city)
}

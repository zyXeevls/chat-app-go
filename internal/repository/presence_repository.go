package repository

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type PresenceRepository struct {
	Redis *redis.Client
}

func (r *PresenceRepository) SetOnline(userID string) {
	r.Redis.Set(context.Background(), "presence:"+userID, "1", 30*time.Second)
}

func (r *PresenceRepository) SetOffline(userID string) {
	r.Redis.Del(context.Background(), "presence:"+userID)
	r.Redis.Set(context.Background(), "lastseen:"+userID, time.Now().Unix(), 0)
}

func (r *PresenceRepository) IsOnline(userID string) bool {
	val, _ := r.Redis.Get(context.Background(), "presence:"+userID).Result()
	return val == "1"
}

func (r *PresenceRepository) GetLastSeen(userID string) int64 {
	val, _ := r.Redis.Get(context.Background(), "lastseen:"+userID).Int64()
	return val
}

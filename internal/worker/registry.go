package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Registry struct {
	ID     string
	Addr   string
	Client *redis.Client
}

type WorkerInfo struct {
	ID       string `json:"id"`
	Addr     string `json:"addr"`
	LastSeen int64  `json:"last_seen"`
}

func NewRegistry(id, addr, redisAddr string) *Registry {
	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	return &Registry{ID: id, Addr: addr, Client: rdb}
}

func (r *Registry) StartHeartbeat(ctx context.Context) error {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			info := WorkerInfo{
				ID:       r.ID,
				Addr:     r.Addr,
				LastSeen: time.Now().Unix(),
			}
			b, _ := json.Marshal(info)
			key := fmt.Sprintf("glazel:workers:%s", r.ID)
			if err := r.Client.Set(ctx, key, b, 10*time.Second).Err(); err != nil {
				fmt.Println("heartbeat error:", err)
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

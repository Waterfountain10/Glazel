package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type WorkerInfo struct {
	ID       string `json:"id"`
	Addr     string `json:"addr"`
	LastSeen int64  `json:"last_seen"`
}

type Registry struct {
	Client *redis.Client
}

func NewRegistry(redisAddr string) *Registry {
	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	return &Registry{Client: rdb}
}

func (r *Registry) Monitor(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			keys, _ := r.Client.Keys(ctx, "glazel:workers:*").Result()
			fmt.Println("Active workers:")
			for _, k := range keys {
				val, err := r.Client.Get(ctx, k).Result()
				if err != nil {
					continue
				}
				var info WorkerInfo
				if err := json.Unmarshal([]byte(val), &info); err == nil {
					fmt.Printf("  - %s @ %s (last seen %ds ago)\n",
						info.ID, info.Addr, int(time.Now().Unix()-info.LastSeen))
				}
			}
			fmt.Println()
		case <-ctx.Done():
			return
		}
	}
}

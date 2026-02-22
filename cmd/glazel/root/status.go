package root

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/cobra"
)

const (
	reset = "\033[0m"
	bold  = "\033[1m"
	blue  = "\033[34m"
	green = "\033[32m"
	dim   = "\033[2m"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show active Glazel workers",
	Run: func(cmd *cobra.Command, args []string) {

		ctx := context.Background()
		rdb := redis.NewClient(&redis.Options{
			Addr: redisAddr,
		})

		keys, err := rdb.Keys(ctx, "glazel:workers:*").Result()
		if err != nil {
			fmt.Println("Redis connection failed:", err)
			return
		}

		fmt.Printf("\n%s%s>>> GLAZEL CLUSTER STATUS%s\n", bold, blue, reset)
		fmt.Println()

		if len(keys) == 0 {
			fmt.Println("No active workers found.")
			return
		}

		fmt.Printf("%-12s %-12s %-15s\n", "WORKER", "ADDRESS", "LAST SEEN")
		fmt.Println("------------------------------------------------")

		for _, k := range keys {
			val, err := rdb.Get(ctx, k).Result()
			if err != nil {
				continue
			}

			var info struct {
				ID       string `json:"id"`
				Addr     string `json:"addr"`
				LastSeen int64  `json:"last_seen"`
			}

			if err := json.Unmarshal([]byte(val), &info); err != nil {
				continue
			}

			age := int(time.Now().Unix() - info.LastSeen)

			fmt.Printf(
				"%s%-12s%s %-12s %s%ds ago%s\n",
				green, info.ID, reset,
				info.Addr,
				dim, age, reset,
			)
		}

		fmt.Println()
	},
}

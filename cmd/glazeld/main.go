package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/Waterfountain10/glazel/internal/orchestrator"
)

var redisAddr = flag.String("redis", "localhost:6379", "Redis address")

func main() {
	flag.Parse()
	logger := log.New(os.Stdout, "[glazeld] ", log.LstdFlags|log.Lmicroseconds)
	logger.Println("starting orchestrator with Redis", *redisAddr)

	ctx := context.Background()
	reg := orchestrator.NewRegistry(*redisAddr)
	reg.Monitor(ctx)
}

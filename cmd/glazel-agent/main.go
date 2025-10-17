package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/Waterfountain10/glazel/internal/worker"
)

var (
	httpAddr  = flag.String("http", ":9090", "HTTP listen address")
	id        = flag.String("id", "agent-1", "Worker ID")
	redisAddr = flag.String("redis", "localhost:6379", "Redis address")
)

func main() {
	flag.Parse()
	logger := log.New(os.Stdout, "[glazel-agent] ", log.LstdFlags|log.Lmicroseconds)
	logger.Printf("starting agent %s on %s", *id, *httpAddr)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	reg := worker.NewRegistry(*id, *httpAddr, *redisAddr)
	go reg.StartHeartbeat(ctx)

	// simple block so the program keeps running
	select {}
}

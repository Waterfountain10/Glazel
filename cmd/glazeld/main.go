package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/Waterfountain10/glazel/internal/orchestrator"
)

var (
	httpAddr  = flag.String("http", ":8080", "HTTP listen address")
	redisAddr = flag.String("redis", "localhost:6379", "Redis address")
)

func main() {
	flag.Parse()

	logger := log.New(os.Stdout, "[glazeld] ", log.LstdFlags|log.Lmicroseconds)
	logger.Printf("starting orchestrator on %s", *httpAddr)

	rdb := redis.NewClient(&redis.Options{Addr: *redisAddr})

	buildSrv := &orchestrator.BuildServer{
		Redis:   rdb,
		CASRoot: ".glazel/cas",
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	mux.HandleFunc("/v1/build", buildSrv.HandleBuild)

	srv := &http.Server{
		Addr:              *httpAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil {
		logger.Fatalf("server error: %v", err)
	}
}

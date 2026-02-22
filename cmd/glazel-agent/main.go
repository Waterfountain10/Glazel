package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"time"

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

	ctx := context.Background()

	// heartbeat to Redis
	reg := worker.NewRegistry(*id, *httpAddr, *redisAddr)
	go func() {
		_ = reg.StartHeartbeat(ctx)
	}()

	// exec server
	mux := http.NewServeMux()
	execSrv := &worker.ExecServer{WorkerID: *id}
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/v1/exec", execSrv.HandleExec)

	srv := &http.Server{
		Addr:              *httpAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatalf("agent http server error: %v", err)
	}
}

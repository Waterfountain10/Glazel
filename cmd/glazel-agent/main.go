package main

import (
    "flag"
    "fmt"
    "log"
    "net/http"
    "os"
    "time"
)

var (
    httpAddr = flag.String("http", ":9090", "HTTP listen address")
    id       = flag.String("id", "agent-1", "Worker ID")
)

func main() {
    flag.Parse()
    logger := log.New(os.Stdout, "[glazel-agent] ", log.LstdFlags|log.Lmicroseconds)
    logger.Printf("starting agent %s on %s", *id, *httpAddr)

    mux := http.NewServeMux()
    mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
        w.WriteHeader(http.StatusOK)
        _, _ = w.Write([]byte("ok"))
    })

    srv := &http.Server{
        Addr:              *httpAddr,
        Handler:           mux,
        ReadHeaderTimeout: 5 * time.Second,
    }

    if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        logger.Fatalf("agent server error: %v", err)
    }
    fmt.Println("agent stopped")
}

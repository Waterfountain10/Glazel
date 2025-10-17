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
    httpAddr = flag.String("http", ":8080", "HTTP listen address")
)

func main() {
    flag.Parse()
    logger := log.New(os.Stdout, "[glazeld] ", log.LstdFlags|log.Lmicroseconds)
    logger.Printf("starting orchestrator on %s", *httpAddr)

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
        logger.Fatalf("server error: %v", err)
    }
    fmt.Println("orchestrator stopped")
}

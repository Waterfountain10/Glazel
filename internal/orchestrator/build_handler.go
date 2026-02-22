package orchestrator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/Waterfountain10/glazel/internal/api"
	"github.com/Waterfountain10/glazel/internal/utils"
)

type BuildServer struct {
	Redis     *redis.Client
	NextIndex int
}

func (s *BuildServer) listWorkers(ctx context.Context) ([]WorkerInfo, error) {
	keys, err := s.Redis.Keys(ctx, "glazel:workers:*").Result()
	if err != nil {
		return nil, err
	}
	sort.Strings(keys)

	now := time.Now().Unix()
	out := make([]WorkerInfo, 0, len(keys))
	for _, k := range keys {
		val, err := s.Redis.Get(ctx, k).Result()
		if err != nil {
			continue
		}
		var w WorkerInfo
		if json.Unmarshal([]byte(val), &w) != nil {
			continue
		}
		// Consider worker healthy if last_seen within ~10s
		if now-w.LastSeen <= 10 {
			out = append(out, w)
		}
	}
	return out, nil
}

func (s *BuildServer) pickWorker(workers []WorkerInfo) WorkerInfo {
	if len(workers) == 1 {
		return workers[0]
	}
	w := workers[s.NextIndex%len(workers)]
	s.NextIndex++
	return w
}

// Cache format (simple, real):
// glazel:obj:<hash> -> raw .o bytes (Redis string)
func (s *BuildServer) cacheKeyObj(hash string) string {
	return "glazel:obj:" + hash
}

func (s *BuildServer) HandleBuild(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	var req api.BuildRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	if req.Compiler == "" {
		req.Compiler = "g++"
	}
	if req.Out == "" {
		req.Out = "a.out"
	}

	workers, err := s.listWorkers(ctx)
	if err != nil {
		http.Error(w, "redis error", http.StatusInternalServerError)
		return
	}
	if len(workers) == 0 {
		http.Error(w, "no healthy workers", http.StatusServiceUnavailable)
		return
	}

	resp := api.BuildResponse{Rows: make([]api.TaskRow, 0, len(req.Files))}

	for _, path := range req.Files {
		srcBytes, err := osReadFile(path)
		if err != nil {
			http.Error(w, fmt.Sprintf("read %s failed: %v", path, err), http.StatusBadRequest)
			return
		}

		// hash includes: file bytes + compiler + flags (keeps it deterministic)
		hashInput := append([]byte{}, srcBytes...)
		hashInput = append(hashInput, []byte(req.Compiler)...)
		for _, f := range req.CxxFlags {
			hashInput = append(hashInput, []byte(f)...)
		}
		hash := utils.Sha256Hex(hashInput)
		hash4 := utils.Last4(hash)

		// Cache check
		objKey := s.cacheKeyObj(hash)
		_, err = s.Redis.Get(ctx, objKey).Bytes()
		if err == nil {
			resp.CacheHits++
			resp.Rows = append(resp.Rows, api.TaskRow{
				File:     path,
				WorkerID: "-",
				Status:   "HIT",
				HashFull: hash,
				Hash4:    hash4,
			})
			continue
		}

		// MISS -> dispatch to worker
		worker := s.pickWorker(workers)

		execReq := api.ExecRequest{
			TaskID:   hash4,
			FileName: filepath.Base(path),
			Source:   srcBytes,
			Compiler: req.Compiler,
			Args:     append([]string{}, req.CxxFlags...),
			HashFull: hash,
		}

		body, _ := json.Marshal(execReq)
		httpResp, err := http.Post("http://localhost"+worker.Addr+"/v1/exec", "application/json", bytes.NewReader(body))
		if err != nil {
			http.Error(w, fmt.Sprintf("worker dispatch failed (%s): %v", worker.ID, err), http.StatusBadGateway)
			return
		}
		defer httpResp.Body.Close()

		var execResp api.ExecResponse
		if err := json.NewDecoder(httpResp.Body).Decode(&execResp); err != nil {
			http.Error(w, "bad worker response", http.StatusBadGateway)
			return
		}
		if !execResp.Ok {
			http.Error(w, fmt.Sprintf("compile failed on %s: %s", worker.ID, execResp.Stderr), http.StatusBadRequest)
			return
		}

		// Store object bytes into Redis
		if err := s.Redis.Set(ctx, objKey, execResp.Object, 0).Err(); err != nil {
			http.Error(w, "cache store failed", http.StatusInternalServerError)
			return
		}

		resp.CacheMisses++
		resp.Rows = append(resp.Rows, api.TaskRow{
			File:     path,
			WorkerID: worker.ID,
			Status:   "MISS",
			HashFull: hash,
			Hash4:    hash4,
		})
	}

	// For this milestone: we skip linking. Demo focus is distributed compile + cache.
	resp.OutPath = "(linking next milestone)"

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// small wrapper to keep imports clean
func osReadFile(path string) ([]byte, error) { return os.ReadFile(path) }

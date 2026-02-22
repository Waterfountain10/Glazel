package orchestrator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/Waterfountain10/glazel/internal/api"
	"github.com/Waterfountain10/glazel/internal/storage"
	"github.com/Waterfountain10/glazel/internal/utils"
)

type BuildServer struct {
	Redis     *redis.Client
	NextIndex int
	CASRoot   string
}

func (s *BuildServer) listWorkers(ctx context.Context) ([]WorkerInfo, error) {
	keys, err := s.Redis.Keys(ctx, "glazel:workers:*").Result()
	if err != nil {
		return nil, err
	}
	sort.Strings(keys)

	now := time.Now().Unix()
	out := []WorkerInfo{}

	for _, k := range keys {
		val, err := s.Redis.Get(ctx, k).Result()
		if err != nil {
			continue
		}
		var w WorkerInfo
		if json.Unmarshal([]byte(val), &w) != nil {
			continue
		}
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

	cas := storage.NewCAS(s.CASRoot)
	_ = cas.EnsureDirs()

	resp := api.BuildResponse{}

	for _, path := range req.Files {

		srcBytes, err := os.ReadFile(path)
		if err != nil {
			http.Error(w, fmt.Sprintf("read %s failed: %v", path, err), http.StatusBadRequest)
			return
		}

		hashInput := append([]byte{}, srcBytes...)
		hashInput = append(hashInput, []byte(req.Compiler)...)
		for _, f := range req.CxxFlags {
			hashInput = append(hashInput, []byte(f)...)
		}

		hash := utils.Sha256Hex(hashInput)
		hash4 := utils.Last4(hash)

		if cas.HasObj(hash) {
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

		worker := s.pickWorker(workers)

		execReq := api.ExecRequest{
			TaskID:   hash4,
			FileName: filepath.Base(path),
			Source:   srcBytes,
			Compiler: req.Compiler,
			Args:     req.CxxFlags,
			HashFull: hash,
		}

		body, _ := json.Marshal(execReq)

		httpResp, err := http.Post(
			"http://localhost"+worker.Addr+"/v1/exec",
			"application/json",
			bytes.NewReader(body),
		)
		if err != nil {
			http.Error(w, fmt.Sprintf("dispatch failed: %v", err), http.StatusBadGateway)
			return
		}
		defer httpResp.Body.Close()

		var execResp api.ExecResponse
		if err := json.NewDecoder(httpResp.Body).Decode(&execResp); err != nil {
			http.Error(w, "bad worker response", http.StatusBadGateway)
			return
		}
		if !execResp.Ok {
			http.Error(w, execResp.Stderr, http.StatusBadRequest)
			return
		}

		if _, err := cas.PutObj(hash, execResp.Object); err != nil {
			http.Error(w, "cas store failed", http.StatusInternalServerError)
			return
		}

		_ = s.Redis.Set(ctx, s.cacheKeyObj(hash), "present", 0)

		resp.CacheMisses++
		resp.Rows = append(resp.Rows, api.TaskRow{
			File:     path,
			WorkerID: worker.ID,
			Status:   "MISS",
			HashFull: hash,
			Hash4:    hash4,
		})
	}

	// ---- LINK STAGE ----

	outDir := ".glazel/out"
	_ = os.MkdirAll(outDir, 0755)

	exePath := filepath.Join(outDir, req.Out)

	objPaths := []string{}
	for _, row := range resp.Rows {
		objPaths = append(objPaths, cas.ObjPath(row.HashFull))
	}

	linkArgs := append(objPaths, "-o", exePath)

	linkCmd := exec.Command(req.Compiler, linkArgs...)
	output, err := linkCmd.CombinedOutput()
	if err != nil {
		http.Error(w, fmt.Sprintf("link failed: %s", string(output)), http.StatusBadRequest)
		return
	}

	resp.OutPath = exePath

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

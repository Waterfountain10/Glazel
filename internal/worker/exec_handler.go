package worker

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/Waterfountain10/glazel/internal/api"
)

type ExecServer struct {
	WorkerID string
}

func (s *ExecServer) HandleExec(w http.ResponseWriter, r *http.Request) {
	var req api.ExecRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	tmpDir, err := os.MkdirTemp("", "glazel-agent-*")
	if err != nil {
		http.Error(w, "tempdir failed", http.StatusInternalServerError)
		return
	}
	defer os.RemoveAll(tmpDir)

	srcPath := filepath.Join(tmpDir, req.FileName)
	objPath := filepath.Join(tmpDir, req.FileName+".o")

	if err := os.WriteFile(srcPath, req.Source, 0644); err != nil {
		http.Error(w, "write source failed", http.StatusInternalServerError)
		return
	}

	// compile: compiler [args...] -c srcPath -o objPath
	args := append([]string{}, req.Args...)
	args = append(args, "-c", srcPath, "-o", objPath)

	cmd := exec.Command(req.Compiler, args...)
	stderrPipe, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		http.Error(w, "exec start failed", http.StatusInternalServerError)
		return
	}

	stderrBytes, _ := io.ReadAll(stderrPipe)
	err = cmd.Wait()

	resp := api.ExecResponse{
		TaskID:   req.TaskID,
		WorkerID: s.WorkerID,
		Ok:       err == nil,
		Stderr:   string(stderrBytes),
		HashFull: req.HashFull,
	}

	if resp.Ok {
		objBytes, readErr := os.ReadFile(objPath)
		if readErr != nil {
			resp.Ok = false
			resp.Stderr = fmt.Sprintf("read object failed: %v", readErr)
		} else {
			resp.Object = objBytes
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

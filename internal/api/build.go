package api

type BuildRequest struct {
	Files    []string `json:"files"`     // e.g. ["examples/hello/main.cpp", ...]
	Out      string   `json:"out"`       // e.g. "hello"
	Compiler string   `json:"compiler"`  // e.g. "g++"
	CxxFlags []string `json:"cxx_flags"` // e.g. ["-O2", "-std=c++20"]
}

type TaskRow struct {
	File     string `json:"file"`
	WorkerID string `json:"worker_id"`
	Status   string `json:"status"`    // HIT or MISS
	HashFull string `json:"hash_full"` // sha256 hex
	Hash4    string `json:"hash4"`     // last 4 hex
}

type BuildResponse struct {
	Rows        []TaskRow `json:"rows"`
	CacheHits   int       `json:"cache_hits"`
	CacheMisses int       `json:"cache_misses"`
	OutPath     string    `json:"out_path"`
}

type ExecRequest struct {
	TaskID   string   `json:"task_id"`
	FileName string   `json:"file_name"` // base name like main.cpp
	Source   []byte   `json:"source"`    // file bytes
	Compiler string   `json:"compiler"`
	Args     []string `json:"args"` // compile args (excluding input/output)
	HashFull string   `json:"hash_full"`
}

type ExecResponse struct {
	TaskID   string `json:"task_id"`
	WorkerID string `json:"worker_id"`
	Ok       bool   `json:"ok"`
	Stderr   string `json:"stderr"`
	Object   []byte `json:"object"` // compiled .o bytes
	HashFull string `json:"hash_full"`
}

package api

type Task struct {
    ID      string   `json:"id"`
    Target  string   `json:"target"`
    Inputs  []string `json:"inputs"`
    Command []string `json:"command"`
    Hash    string   `json:"hash"`
}

type Artifact struct {
    TaskID   string `json:"task_id"`
    Path     string `json:"path"`
    Hash     string `json:"hash"`
    FromCache bool  `json:"from_cache"`
}

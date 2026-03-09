package project

const CurrentSchemaVersion = 2

type Project struct {
	Name             string `json:"name"`
	Path             string `json:"path"`
	UseFastWorktrees bool   `json:"use_fast_worktrees,omitempty"`
}

type storeData struct {
	Version  int                 `json:"version"`
	Projects map[string]*Project `json:"projects"`
}

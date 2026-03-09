package project

const CurrentSchemaVersion = 2

const DefaultCIWaitMinutes = 5

type Project struct {
	Name              string `json:"name"`
	Path              string `json:"path"`
	DefaultBaseBranch string `json:"default_base_branch,omitempty"`
	CIWaitMinutes     int    `json:"ci_wait_minutes,omitempty"`
}

func (p *Project) EffectiveCIWaitMinutes() int {
	if p.CIWaitMinutes <= 0 {
		return DefaultCIWaitMinutes
	}
	return p.CIWaitMinutes
}

func (p *Project) EffectiveBaseBranch() string {
	if p.DefaultBaseBranch == "" {
		return "origin/master"
	}
	return p.DefaultBaseBranch
}

type storeData struct {
	Version  int                 `json:"version"`
	Projects map[string]*Project `json:"projects"`
}

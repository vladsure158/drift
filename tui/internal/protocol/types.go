package protocol

import "time"

type ProjectStatus string

const (
	StatusActive    ProjectStatus = "active"
	StatusIdea      ProjectStatus = "idea"
	StatusPaused    ProjectStatus = "paused"
	StatusDone      ProjectStatus = "done"
	StatusAbandoned ProjectStatus = "abandoned"
)

var AllStatuses = []ProjectStatus{StatusActive, StatusIdea, StatusPaused, StatusDone, StatusAbandoned}

type Goal struct {
	Text string `json:"text"`
	Done bool   `json:"done"`
}

type Note struct {
	Ts   string `json:"ts"`
	Text string `json:"text"`
}

type Links struct {
	Repo   *string `json:"repo"`
	Deploy *string `json:"deploy"`
	Design *string `json:"design"`
}

type Project struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	Description  *string       `json:"description"`
	Status       ProjectStatus `json:"status"`
	Progress     int           `json:"progress"`
	Tags         []string      `json:"tags"`
	Created      string        `json:"created"`
	LastActivity string        `json:"lastActivity"`
	Goals        []Goal        `json:"goals"`
	Notes        []Note        `json:"notes"`
	Links        Links         `json:"links"`

	// Extra fields preserved on round-trip
	Extra map[string]any `json:"-"`
}

type RegistryEntry struct {
	ID           string `json:"id"`
	Path         string `json:"path"`
	Name         string `json:"name"`
	Status       string `json:"status"`
	LastActivity string `json:"lastActivity"`
}

type Registry struct {
	Version  int             `json:"version"`
	Projects []RegistryEntry `json:"projects"`
}

// FullProject is a loaded project with filesystem context
type FullProject struct {
	Project
	Path    string
	Missing bool
}

func Now() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05Z")
}

func CalcProgress(goals []Goal) int {
	if len(goals) == 0 {
		return 0
	}
	done := 0
	for _, g := range goals {
		if g.Done {
			done++
		}
	}
	return int(float64(done) / float64(len(goals)) * 100)
}

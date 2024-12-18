package status

import (
	"encoding/json"
	"os"
	"time"
)

type Status struct {
	WatcherRunning     bool      `json:"watcher_running"`
	FilesPending       int       `json:"files_pending"`
	LastCommitHash     string    `json:"last_commit_hash,omitempty"`
	LastCommitTime     time.Time `json:"last_commit_time,omitempty"`
	LastPushSuccessful bool      `json:"last_push_successful"`
	LastPushTime       time.Time `json:"last_push_time,omitempty"`
}

func LoadStatus(path string) (*Status, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var st Status
	if err := json.Unmarshal(data, &st); err != nil {
		return nil, err
	}
	return &st, nil
}

func SaveStatus(path string, st *Status) error {
	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/timr/internal/utils"
)

type TimerState struct {
	Pid       int       `json:"pid"`
	StartTime time.Time `json:"start_time"`
	Duration  string    `json:"duration"`
	EndTime   time.Time `json:"end_time"`
}

type AppState struct {
	Timers []TimerState `json:"timers"`
}

func stateFilePath() string {
	return filepath.Join(utils.XDGDataHome(), "timr", "state.json")
}

func loadState() (AppState, error) {
	path := stateFilePath()
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return AppState{Timers: []TimerState{}}, nil
		}
		return AppState{}, err
	}
	var state AppState
	if err := json.Unmarshal(data, &state); err != nil {
		return AppState{Timers: []TimerState{}}, nil
	}
	return state, nil
}

func saveState(state AppState) error {
	path := stateFilePath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// isProcessRunning checks if a PID is actually running.
func isProcessRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// On Unix, FindProcess always succeeds. We need to send signal 0 to check if it's alive.
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

// getActiveTimers returns only the timers that are still running processes.
// It automatically cleans up stale entries.
func getActiveTimers() ([]TimerState, error) {
	state, err := loadState()
	if err != nil {
		return nil, err
	}
	active := []TimerState{}
	changed := false
	for _, t := range state.Timers {
		if isProcessRunning(t.Pid) && time.Now().Before(t.EndTime) {
			active = append(active, t)
		} else {
			changed = true
		}
	}
	if changed {
		state.Timers = active
		_ = saveState(state)
	}
	return active, nil
}

func addActiveTimer(t TimerState) error {
	state, err := loadState()
	if err != nil {
		return err
	}
	state.Timers = append(state.Timers, t)
	return saveState(state)
}

func removeActiveTimer(pid int) error {
	state, err := loadState()
	if err != nil {
		return err
	}
	filtered := []TimerState{}
	for _, t := range state.Timers {
		if t.Pid != pid {
			filtered = append(filtered, t)
		}
	}
	state.Timers = filtered
	return saveState(state)
}

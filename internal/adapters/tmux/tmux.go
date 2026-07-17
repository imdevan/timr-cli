package tmux

import (
	"os"
	"os/exec"
	"strings"
)

// Adapter interacts with the Tmux window environment.
type Adapter struct {
	windowID     string
	originalName string
	isActive     bool
}

// New returns a new Tmux adapter initialized with the current window ID and original name.
func New() *Adapter {
	a := &Adapter{
		isActive: os.Getenv("TMUX") != "",
	}
	if a.isActive {
		if id, err := a.queryWindowID(); err == nil {
			a.windowID = id
			if name, err := a.queryWindowName(id); err == nil {
				a.originalName = name
			}
		}
	}
	return a
}

// IsActive reports whether the current process is running inside a Tmux session.
func (a *Adapter) IsActive() bool {
	return a.isActive
}

// WindowID returns the queried window ID.
func (a *Adapter) WindowID() string {
	return a.windowID
}

// OriginalName returns the original window name retrieved at initialization.
func (a *Adapter) OriginalName() string {
	return a.originalName
}

// RenameWindow sets the targeted Tmux window name.
func (a *Adapter) RenameWindow(name string) error {
	if !a.isActive || a.windowID == "" {
		return nil
	}
	return exec.Command("tmux", "rename-window", "-t", a.windowID, name).Run()
}

func (a *Adapter) queryWindowID() (string, error) {
	out, err := exec.Command("tmux", "display-message", "-p", "#{window_id}").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func (a *Adapter) queryWindowName(id string) (string, error) {
	out, err := exec.Command("tmux", "display-message", "-t", id, "-p", "#W").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

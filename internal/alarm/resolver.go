// Package alarm resolves the configured alarm_sound value to a concrete file path.
//
// alarm_sound supports three forms:
//   - A single file path        → use that file
//   - A directory path          → pick a random media file from the directory
//   - A comma-separated list    → pick a random entry (each entry may itself be a
//     file or a directory, resolved as above)
//
// Empty string → caller should fall back to terminal beep.
package alarm

import (
	"math/rand/v2"
	"os"
	"path/filepath"
	"strings"

	"github.com/timr/internal/domain"
)

// mediaExtensions is the set of file extensions considered playable.
var mediaExtensions = map[string]struct{}{
	".mp3":  {},
	".wav":  {},
	".ogg":  {},
	".flac": {},
	".aac":  {},
	".m4a":  {},
	".opus": {},
	".mp4":  {},
	".mkv":  {},
	".webm": {},
	".avi":  {},
	".mov":  {},
	".flv":  {},
}

// Resolve returns the concrete alarm sound file path to play based on
// cfg.AlarmSound, which may be:
//   - empty          → returns "" (caller should use terminal beep)
//   - a single path  → returned directly (or random file if it's a dir)
//   - "a,b,c" CSV    → picks a random entry, then resolves as above
func Resolve(cfg domain.Config) string {
	raw := strings.TrimSpace(cfg.AlarmSound)
	if raw == "" {
		return ""
	}

	// Split on commas to support CSV lists.
	parts := strings.Split(raw, ",")
	var candidates []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		expanded := expandPath(p)
		info, err := os.Stat(expanded)
		if err != nil {
			// Path doesn't exist — pass through so the player gives a clear error.
			candidates = append(candidates, expanded)
			continue
		}
		if info.IsDir() {
			candidates = append(candidates, mediaFilesInDir(expanded)...)
		} else {
			candidates = append(candidates, expanded)
		}
	}

	if len(candidates) == 0 {
		return ""
	}
	return candidates[rand.IntN(len(candidates))]
}

// mediaFilesInDir returns all media files (non-recursively) inside dir.
func mediaFilesInDir(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(e.Name()))
		if _, ok := mediaExtensions[ext]; ok {
			out = append(out, filepath.Join(dir, e.Name()))
		}
	}
	return out
}

// expandPath expands ~ and environment variables in a path.
func expandPath(value string) string {
	expanded := os.ExpandEnv(value)
	if expanded == "" {
		return expanded
	}
	if expanded == "~" {
		if home, err := os.UserHomeDir(); err == nil {
			return home
		}
		return expanded
	}
	if strings.HasPrefix(expanded, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, strings.TrimPrefix(expanded, "~/"))
		}
	}
	return expanded
}

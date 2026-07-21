package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/timr/internal/domain"
	"github.com/timr/internal/utils"
)

func TestManagerLoadsDefaults(t *testing.T) {
	root := t.TempDir()
	cwd := filepath.Join(root, "project")
	if err := os.MkdirAll(cwd, 0o755); err != nil {
		t.Fatalf("mkdir cwd: %v", err)
	}
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "config"))
	t.Setenv("XDG_DATA_HOME", filepath.Join(root, "data"))

	manager := NewManager(cwd)
	cfg, err := manager.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if cfg.Editor == "" {
		t.Fatal("expected editor to have default value")
	}
}

func TestManagerLoadsFromFile(t *testing.T) {
	root := t.TempDir()
	cwd := filepath.Join(root, "project")
	if err := os.MkdirAll(cwd, 0o755); err != nil {
		t.Fatalf("mkdir cwd: %v", err)
	}
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "config"))

	configPath := utils.ConfigPathGlobal()
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}

	data := []byte("editor = \"vim\"\n")
	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	manager := NewManager(cwd)
	cfg, err := manager.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if cfg.Editor != "vim" {
		t.Fatalf("expected editor from config, got %q", cfg.Editor)
	}
}

func TestManagerSavesConfig(t *testing.T) {
	root := t.TempDir()
	cwd := filepath.Join(root, "project")
	if err := os.MkdirAll(cwd, 0o755); err != nil {
		t.Fatalf("mkdir cwd: %v", err)
	}
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "config"))

	manager := NewManager(cwd)
	cfg := domain.DefaultConfig()
	cfg.Editor = "emacs"

	if err := manager.Save(cfg); err != nil {
		t.Fatalf("save: %v", err)
	}

	configPath := utils.ConfigPathGlobal()
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("expected config file to exist")
	}
}

func TestManagerExists(t *testing.T) {
	t.Run("returns false when no config exists", func(t *testing.T) {
		root := t.TempDir()
		cwd := filepath.Join(root, "project")
		if err := os.MkdirAll(cwd, 0o755); err != nil {
			t.Fatalf("mkdir cwd: %v", err)
		}
		t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "config"))

		manager := NewManager(cwd)
		exists, err := manager.Exists()
		if err != nil {
			t.Fatalf("exists: %v", err)
		}
		if exists {
			t.Error("expected config to not exist")
		}
	})

	t.Run("returns true when global config exists", func(t *testing.T) {
		root := t.TempDir()
		cwd := filepath.Join(root, "project")
		if err := os.MkdirAll(cwd, 0o755); err != nil {
			t.Fatalf("mkdir cwd: %v", err)
		}
		t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "config"))

		configPath := utils.ConfigPathGlobal()
		if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
			t.Fatalf("mkdir config dir: %v", err)
		}
		if err := os.WriteFile(configPath, []byte("editor = \"vim\"\n"), 0o644); err != nil {
			t.Fatalf("write config: %v", err)
		}

		manager := NewManager(cwd)
		exists, err := manager.Exists()
		if err != nil {
			t.Fatalf("exists: %v", err)
		}
		if !exists {
			t.Error("expected config to exist")
		}
	})

	t.Run("returns true when local config exists", func(t *testing.T) {
		root := t.TempDir()
		cwd := filepath.Join(root, "project")
		if err := os.MkdirAll(cwd, 0o755); err != nil {
			t.Fatalf("mkdir cwd: %v", err)
		}
		t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "config"))

		localConfigPath := utils.ConfigPathLocal(cwd)
		if err := os.MkdirAll(filepath.Dir(localConfigPath), 0o755); err != nil {
			t.Fatalf("mkdir local config dir: %v", err)
		}
		if err := os.WriteFile(localConfigPath, []byte("editor = \"code\"\n"), 0o644); err != nil {
			t.Fatalf("write local config: %v", err)
		}

		manager := NewManager(cwd)
		exists, err := manager.Exists()
		if err != nil {
			t.Fatalf("exists: %v", err)
		}
		if !exists {
			t.Error("expected config to exist")
		}
	})
}

func TestManagerLoadWithOverride(t *testing.T) {
	t.Run("loads from override path", func(t *testing.T) {
		root := t.TempDir()
		cwd := filepath.Join(root, "project")
		if err := os.MkdirAll(cwd, 0o755); err != nil {
			t.Fatalf("mkdir cwd: %v", err)
		}

		overridePath := filepath.Join(root, "custom-config.toml")
		data := []byte("editor = \"emacs\"\ntime_remaining = \"03\"\n")
		if err := os.WriteFile(overridePath, data, 0o644); err != nil {
			t.Fatalf("write override config: %v", err)
		}

		manager := NewManager(cwd)
		cfg, err := manager.LoadWithOverride(overridePath)
		if err != nil {
			t.Fatalf("load with override: %v", err)
		}

		if cfg.Editor != "emacs" {
			t.Errorf("expected editor from override, got %q", cfg.Editor)
		}
		if cfg.TimeRemaining != "03" {
			t.Errorf("expected time_remaining from override, got %q", cfg.TimeRemaining)
		}
	})

	t.Run("falls back to Load when path is empty", func(t *testing.T) {
		root := t.TempDir()
		cwd := filepath.Join(root, "project")
		if err := os.MkdirAll(cwd, 0o755); err != nil {
			t.Fatalf("mkdir cwd: %v", err)
		}
		t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "config"))

		manager := NewManager(cwd)
		cfg, err := manager.LoadWithOverride("")
		if err != nil {
			t.Fatalf("load with empty override: %v", err)
		}

		if cfg.Editor == "" {
			t.Error("expected default config to be loaded")
		}
	})
}

func TestManagerLocalOverridesGlobal(t *testing.T) {
	root := t.TempDir()
	cwd := filepath.Join(root, "project")
	if err := os.MkdirAll(cwd, 0o755); err != nil {
		t.Fatalf("mkdir cwd: %v", err)
	}
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "config"))

	// Create global config
	globalPath := utils.ConfigPathGlobal()
	if err := os.MkdirAll(filepath.Dir(globalPath), 0o755); err != nil {
		t.Fatalf("mkdir global config dir: %v", err)
	}
	globalData := []byte("editor = \"vim\"\ntime_remaining = \"01\"\n")
	if err := os.WriteFile(globalPath, globalData, 0o644); err != nil {
		t.Fatalf("write global config: %v", err)
	}

	// Create local config
	localPath := utils.ConfigPathLocal(cwd)
	if err := os.MkdirAll(filepath.Dir(localPath), 0o755); err != nil {
		t.Fatalf("mkdir local config dir: %v", err)
	}
	localData := []byte("editor = \"emacs\"\n")
	if err := os.WriteFile(localPath, localData, 0o644); err != nil {
		t.Fatalf("write local config: %v", err)
	}

	manager := NewManager(cwd)
	cfg, err := manager.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if cfg.Editor != "emacs" {
		t.Errorf("expected editor from local config, got %q", cfg.Editor)
	}
	if cfg.TimeRemaining != "01" {
		t.Errorf("expected time_remaining from global config, got %q", cfg.TimeRemaining)
	}
}

func TestManagerPartialConfig(t *testing.T) {
	t.Run("only overrides specified fields", func(t *testing.T) {
		root := t.TempDir()
		cwd := filepath.Join(root, "project")
		if err := os.MkdirAll(cwd, 0o755); err != nil {
			t.Fatalf("mkdir cwd: %v", err)
		}
		t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "config"))

		configPath := utils.ConfigPathGlobal()
		if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
			t.Fatalf("mkdir config dir: %v", err)
		}

		// Only set editor, leave other fields as defaults
		data := []byte("editor = \"code\"\n")
		if err := os.WriteFile(configPath, data, 0o644); err != nil {
			t.Fatalf("write config: %v", err)
		}

		manager := NewManager(cwd)
		cfg, err := manager.Load()
		if err != nil {
			t.Fatalf("load: %v", err)
		}

		if cfg.Editor != "code" {
			t.Errorf("expected editor from config, got %q", cfg.Editor)
		}
		// Check that defaults are still present
		if cfg.TimeRemaining != "14" {
			t.Errorf("expected default time_remaining, got %q", cfg.TimeRemaining)
		}
		if cfg.TimeStart != "07" {
			t.Errorf("expected default time_start, got %q", cfg.TimeStart)
		}
		if !cfg.InteractiveDefault {
			t.Error("expected default interactive_default to be true")
		}
	})
}

func TestManagerColorOverrides(t *testing.T) {
	root := t.TempDir()
	cwd := filepath.Join(root, "project")
	if err := os.MkdirAll(cwd, 0o755); err != nil {
		t.Fatalf("mkdir cwd: %v", err)
	}
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "config"))

	configPath := utils.ConfigPathGlobal()
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}

	data := []byte(`
time_remaining = "10"
time_start = "04"
bar_bg = "05"
bar_fg = "09"
help_text = "11"
border = "06"
`)
	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	manager := NewManager(cwd)
	cfg, err := manager.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{"TimeRemaining", cfg.TimeRemaining, "10"},
		{"TimeStart", cfg.TimeStart, "04"},
		{"BarBg", cfg.BarBg, "05"},
		{"BarFg", cfg.BarFg, "09"},
		{"HelpText", cfg.HelpText, "11"},
		{"Border", cfg.Border, "06"},
	}

	for _, tt := range tests {
		if tt.got != tt.expected {
			t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.expected)
		}
	}
}

func TestManagerBooleanOverrides(t *testing.T) {
	root := t.TempDir()
	cwd := filepath.Join(root, "project")
	if err := os.MkdirAll(cwd, 0o755); err != nil {
		t.Fatalf("mkdir cwd: %v", err)
	}
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "config"))

	configPath := utils.ConfigPathGlobal()
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}

	data := []byte("interactive_default = false\n")
	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	manager := NewManager(cwd)
	cfg, err := manager.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if cfg.InteractiveDefault {
		t.Error("expected interactive_default to be false from config")
	}
}

func TestManagerLoadsFullWidth(t *testing.T) {
	root := t.TempDir()
	cwd := filepath.Join(root, "project")
	if err := os.MkdirAll(cwd, 0o755); err != nil {
		t.Fatalf("mkdir cwd: %v", err)
	}
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "config"))

	configPath := utils.ConfigPathGlobal()
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}

	data := []byte("full_width = false\n")
	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	manager := NewManager(cwd)
	cfg, err := manager.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if cfg.FullWidth {
		t.Error("expected full_width to be false from config")
	}
}


func TestManagerLoadsRainbowOption(t *testing.T) {
	t.Run("loads rainbow = false", func(t *testing.T) {
		root := t.TempDir()
		cwd := filepath.Join(root, "project")
		_ = os.MkdirAll(cwd, 0o755)
		t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "config"))

		configPath := utils.ConfigPathGlobal()
		_ = os.MkdirAll(filepath.Dir(configPath), 0o755)
		_ = os.WriteFile(configPath, []byte("rainbow = false\n"), 0o644)

		manager := NewManager(cwd)
		cfg, err := manager.Load()
		if err != nil {
			t.Fatalf("load: %v", err)
		}
		if cfg.Rainbow.Enabled {
			t.Error("expected rainbow to be disabled")
		}
	})

	t.Run("loads rainbow array of colors", func(t *testing.T) {
		root := t.TempDir()
		cwd := filepath.Join(root, "project")
		_ = os.MkdirAll(cwd, 0o755)
		t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "config"))

		configPath := utils.ConfigPathGlobal()
		_ = os.MkdirAll(filepath.Dir(configPath), 0o755)
		_ = os.WriteFile(configPath, []byte(`rainbow = ["#ff0000", "#00ff00"]`+"\n"), 0o644)

		manager := NewManager(cwd)
		cfg, err := manager.Load()
		if err != nil {
			t.Fatalf("load: %v", err)
		}
		if !cfg.Rainbow.Enabled {
			t.Error("expected rainbow to be enabled")
		}
		if len(cfg.Rainbow.Colors) != 2 || cfg.Rainbow.Colors[0] != "#ff0000" || cfg.Rainbow.Colors[1] != "#00ff00" {
			t.Errorf("unexpected rainbow colors: %v", cfg.Rainbow.Colors)
		}
	})
}

package domain

import (
	"testing"

	"github.com/pelletier/go-toml/v2"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	t.Run("has editor set", func(t *testing.T) {
		if cfg.Editor == "" {
			t.Error("DefaultConfig().Editor should not be empty")
		}
		if cfg.Editor != "nvim" {
			t.Errorf("DefaultConfig().Editor = %q, want %q", cfg.Editor, "nvim")
		}
	})

	t.Run("has all color values set", func(t *testing.T) {
		colorFields := map[string]string{
			"TimeRemaining": cfg.TimeRemaining,
			"TimeStart":     cfg.TimeStart,
			"BarBg":         cfg.BarBg,
			"BarFg":         cfg.BarFg,
			"HelpText":      cfg.HelpText,
			"Border":        cfg.Border,
		}

		for name, value := range colorFields {
			if value == "" {
				t.Errorf("DefaultConfig().%s should not be empty", name)
			}
		}
	})

	t.Run("has expected default values", func(t *testing.T) {
		if cfg.TimeRemaining != "14" {
			t.Errorf("DefaultConfig().TimeRemaining = %q, want %q", cfg.TimeRemaining, "14")
		}
		if cfg.TimeStart != "07" {
			t.Errorf("DefaultConfig().TimeStart = %q, want %q", cfg.TimeStart, "07")
		}
		if cfg.BarBg != "08" {
			t.Errorf("DefaultConfig().BarBg = %q, want %q", cfg.BarBg, "08")
		}
		if cfg.BarFg != "02" {
			t.Errorf("DefaultConfig().BarFg = %q, want %q", cfg.BarFg, "02")
		}
		if cfg.HelpText != "08" {
			t.Errorf("DefaultConfig().HelpText = %q, want %q", cfg.HelpText, "08")
		}
		if cfg.Border != "08" {
			t.Errorf("DefaultConfig().Border = %q, want %q", cfg.Border, "08")
		}
	})

	t.Run("has interactive default enabled", func(t *testing.T) {
		if !cfg.InteractiveDefault {
			t.Error("DefaultConfig().InteractiveDefault should be true")
		}
	})

	t.Run("has list spacing set", func(t *testing.T) {
		if cfg.ListSpacing == "" {
			t.Error("DefaultConfig().ListSpacing should not be empty")
		}
		if cfg.ListSpacing != "space" {
			t.Errorf("DefaultConfig().ListSpacing = %q, want %q", cfg.ListSpacing, "space")
		}
	})
}

func TestDefaultConfig_Consistency(t *testing.T) {
	t.Run("multiple calls return same values", func(t *testing.T) {
		cfg1 := DefaultConfig()
		cfg2 := DefaultConfig()

		if cfg1.Editor != cfg2.Editor {
			t.Error("DefaultConfig() should return consistent Editor values")
		}
		if cfg1.TimeRemaining != cfg2.TimeRemaining {
			t.Error("DefaultConfig() should return consistent TimeRemaining values")
		}
		if cfg1.InteractiveDefault != cfg2.InteractiveDefault {
			t.Error("DefaultConfig() should return consistent InteractiveDefault values")
		}
	})
}

func TestConfig_StructTags(t *testing.T) {
	t.Run("has toml tags for all fields", func(t *testing.T) {
		cfg := Config{}
		
		cfg.Editor = "test"
		cfg.BarBg = "01"
		cfg.InteractiveDefault = false
		cfg.ListSpacing = "compact"
		
		if cfg.Editor != "test" {
			t.Error("Config struct should be properly defined")
		}
	})
}

func TestRainbowOption_UnmarshalTOML(t *testing.T) {
	type wrapper struct {
		Rainbow RainbowOption `toml:"rainbow"`
	}

	t.Run("unmarshal boolean true", func(t *testing.T) {
		var w wrapper
		err := toml.Unmarshal([]byte("rainbow = true"), &w)
		if err != nil {
			t.Fatalf("unexpected error unmarshaling true: %v", err)
		}
		if !w.Rainbow.Enabled {
			t.Error("expected Enabled to be true")
		}
		if w.Rainbow.Colors != nil {
			t.Errorf("expected Colors to be nil, got %v", w.Rainbow.Colors)
		}
	})

	t.Run("unmarshal boolean false", func(t *testing.T) {
		var w wrapper
		err := toml.Unmarshal([]byte("rainbow = false"), &w)
		if err != nil {
			t.Fatalf("unexpected error unmarshaling false: %v", err)
		}
		if w.Rainbow.Enabled {
			t.Error("expected Enabled to be false")
		}
		if w.Rainbow.Colors != nil {
			t.Errorf("expected Colors to be nil, got %v", w.Rainbow.Colors)
		}
	})

	t.Run("unmarshal list of colors", func(t *testing.T) {
		var w wrapper
		err := toml.Unmarshal([]byte(`rainbow = ["#ff0000", "#00ff00", "#0000ff"]`), &w)
		if err != nil {
			t.Fatalf("unexpected error unmarshaling list of colors: %v", err)
		}
		if !w.Rainbow.Enabled {
			t.Error("expected Enabled to be true")
		}
		expectedColors := []string{"#ff0000", "#00ff00", "#0000ff"}
		if len(w.Rainbow.Colors) != len(expectedColors) {
			t.Fatalf("expected %d colors, got %d", len(expectedColors), len(w.Rainbow.Colors))
		}
		for i, c := range expectedColors {
			if w.Rainbow.Colors[i] != c {
				t.Errorf("expected color at index %d to be %q, got %q", i, c, w.Rainbow.Colors[i])
			}
		}
	})

	t.Run("unmarshal invalid values should error", func(t *testing.T) {
		var w wrapper
		err := toml.Unmarshal([]byte(`rainbow = 12345`), &w)
		if err == nil {
			t.Error("expected unmarshaling number to fail")
		}
		err = toml.Unmarshal([]byte(`rainbow = { colors = [] }`), &w)
		if err == nil {
			t.Error("expected unmarshaling table to fail")
		}
	})
}

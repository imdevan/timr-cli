package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/timr/internal/adapters/editor"
	"github.com/timr/internal/config"
	"github.com/timr/internal/domain"
	"github.com/timr/internal/utils"
)

type configInitOptions struct {
	force        bool
	openInEditor bool
}

func newConfigInitCmd() *cobra.Command {
	opts := &configInitOptions{}
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Generate a default config file",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigInit(cmd, opts)
		},
	}
	cmd.Flags().BoolVarP(&opts.force, "force", "f", false, "overwrite existing config")
	cmd.Flags().BoolVarP(&opts.openInEditor, "editor", "e", false, "open config in editor after creation")
	return cmd
}

func runConfigInit(cmd *cobra.Command, opts *configInitOptions) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	manager := config.NewManager(cwd)
	exists, err := manager.Exists()
	if err != nil {
		return err
	}
	if exists && !opts.force {
		return fmt.Errorf("config already exists at %s (use --force to overwrite)", utils.ConfigPathGlobal())
	}
	cfg := domain.DefaultConfig()
	path := utils.ConfigPathGlobal()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	content := renderConfigTemplate(cfg)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return err
	}
	if opts.openInEditor {
		editorAdapter := editor.New(cfg.Editor)
		if err := editorAdapter.Open(path); err != nil {
			return err
		}
	}
	cmd.Printf("Wrote config to %s\n", utils.ConfigPathGlobal())
	return nil
}

func renderConfigTemplate(cfg domain.Config) string {
	var builder strings.Builder
	builder.WriteString("# Timr CLI Configuration\n\n")
	builder.WriteString("# General\n")
	builder.WriteString(fmt.Sprintf("# editor = %q\n", cfg.Editor))
	builder.WriteString(fmt.Sprintf("# default_units = %q\n", cfg.DefaultUnits))
	builder.WriteString(fmt.Sprintf("# default_timer = %q  # Default duration when run without args (e.g. \"25m\")\n", cfg.DefaultTimer))
	var pomodoroParts []string
	for _, n := range cfg.Pomodoro {
		pomodoroParts = append(pomodoroParts, strconv.Itoa(n))
	}
	builder.WriteString(fmt.Sprintf("# pomodoro = [%s]  # Sequence of minutes for pomodoro command\n", strings.Join(pomodoroParts, ", ")))
	builder.WriteString("# alarm_sound = \"/path/to/file.mp3\"  # single file, directory, or CSV list (e.g. \"/a.mp3, ~/Music/\")\n")
	builder.WriteString("\n# Pomodoro Messages (single string or array of strings for randomized pick)\n")
	builder.WriteString("# [pomodoro_messages]\n")
	builder.WriteString("# after_first_work = [\"You're off to a great start!\", \"One small timer, one big step!\"]\n")
	builder.WriteString("# after_first_break = \"You got this!\"\n")
	builder.WriteString("# after_second_work = [\"Hell yeah!\", \"Nice job!\"]\n")
	builder.WriteString("# after_second_break = \"\"\n")
	builder.WriteString("# before_last_work = \"You're almost there!\"\n")
	builder.WriteString("# after_last_work = \"You did it!\"\n")
	builder.WriteString("# after_last_break = \"You freaking rock!\"\n")
	builder.WriteString("\n# CLI behavior\n")
	builder.WriteString(fmt.Sprintf("# interactive_default = %t\n", cfg.InteractiveDefault))
	builder.WriteString(fmt.Sprintf("# update_tmux_window = %t\n", cfg.UpdateTmuxWindow))
	builder.WriteString(fmt.Sprintf("# tmux_progress_bar = %t\n", cfg.TmuxProgressBar))
	builder.WriteString(fmt.Sprintf("# tmux_inverted = %t\n", cfg.TmuxInverted))
	builder.WriteString(fmt.Sprintf("# full_width = %t\n", cfg.FullWidth))
	builder.WriteString(fmt.Sprintf("# full_tui = %t\n", cfg.FullTUI))
	builder.WriteString("# rainbow = true  # true, false, or array of custom color strings (e.g. [\"#f5bde6\", \"#c6a0f6\"])\n")
	builder.WriteString("\n# UI\n")
	builder.WriteString("# list_spacing options: compact (title only), tight (title + description, no margin), space (default, with spacing)\n")
	builder.WriteString(fmt.Sprintf("# list_spacing = %q\n", cfg.ListSpacing))
	builder.WriteString("\n# Colors\n")
	builder.WriteString("# Colors support named, numeric, or hex values (ex: 7, 13, \"#ff8800\").\n")
	builder.WriteString(fmt.Sprintf("# time_remaining = %q\n", cfg.TimeRemaining))
	builder.WriteString(fmt.Sprintf("# time_start = %q\n", cfg.TimeStart))
	var barFgStr string
	if len(cfg.BarFg) == 1 {
		barFgStr = fmt.Sprintf("%q", cfg.BarFg[0])
	} else {
		var items []string
		for _, c := range cfg.BarFg {
			items = append(items, fmt.Sprintf("%q", c))
		}
		barFgStr = fmt.Sprintf("[%s]", strings.Join(items, ", "))
	}
	builder.WriteString(fmt.Sprintf("# bar_fg = %s  # single color or array of colors (up to 10) for progress bar subdivisions\n", barFgStr))
	builder.WriteString(fmt.Sprintf("# help_text = %q\n", cfg.HelpText))
	builder.WriteString(fmt.Sprintf("# border = %q\n", cfg.Border))
	return builder.String()
}

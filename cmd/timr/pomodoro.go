package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/timr/internal/config"
	"github.com/timr/internal/domain"
	"github.com/timr/internal/ui"
)

func newPomodoroCmd(opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "pomodoro",
		Aliases: []string{"p"},
		Short:   "Start a Pomodoro timer sequence",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get working directory: %w", err)
			}
			manager := config.NewManager(cwd)

			var cfg domain.Config
			if opts.configPath != "" {
				cfg, err = manager.LoadWithOverride(opts.configPath)
			} else {
				cfg, err = manager.Load()
			}
			if err != nil {
				cfg = domain.DefaultConfig()
			}
			theme := ui.ThemeFromConfig(cfg)

			sequence := cfg.Pomodoro
			if len(sequence) == 0 {
				sequence = domain.DefaultConfig().Pomodoro
			}

			isInteractive := cfg.InteractiveDefault
			if cmd.Flags().Changed("interactive") {
				isInteractive = opts.interactive
			}



			for {
				for i, minutes := range sequence {
					durationStr := fmt.Sprintf("%dm", minutes)
					progressStr := fmt.Sprintf("[%d/%d]", i+1, len(sequence))
					cancelled, err := runSingleTimer(cmd, cfg, theme, durationStr, progressStr, opts)
					if err != nil {
						return err
					}
					if cancelled {
						cmd.Printf("Pomodoro sequence cancelled at step %d/%d.\n", i+1, len(sequence))
						return nil
					}

					if isInteractive {
						msg := domain.GetPomodoroMessage(cfg.PomodoroMessages, i, len(sequence))
						title := fmt.Sprintf("[%d/%d] %s", i+1, len(sequence), msg)
						if strings.TrimSpace(msg) == "" {
							title = fmt.Sprintf("[%d/%d] Pomodoro", i+1, len(sequence))
						}
						prompt := domain.GetPomodoroPrompt(i+1, len(sequence))

						confirmed, err := ui.PromptConfirmation(title, prompt, theme, cfg.FullTUI)
						if err != nil || !confirmed {
							cmd.Println("Pomodoro sequence stopped.")
							return nil
						}
					}
				}
				if !isInteractive {
					break
				}
			}
			return nil
		},
	}
	return cmd
}

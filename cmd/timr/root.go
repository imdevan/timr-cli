package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime/debug"
	"strings"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/timr/internal/config"
	"github.com/timr/internal/domain"
	pkg "github.com/timr/internal/package"
	"github.com/timr/internal/ui"
	"github.com/timr/internal/utils"
)

// Metadata loaded from package.toml at build time
var (
	version = pkg.Version()
	name    = pkg.Name()
	short   = pkg.Short()
)

type rootOptions struct {
	configPath  string
	showVersion bool
	interactive bool
	detached    bool
}

var rootCmd = newRootCmd()

// Execute is the CLI entrypoint.
func Execute() error {
	return rootCmd.Execute()
}

func newRootCmd() *cobra.Command {
	opts := &rootOptions{}
	cmd := &cobra.Command{
		Use:   name + " [duration]",
		Short: short,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.showVersion {
				ver := resolvedVersion()
				cmd.Printf("%s\n", ver)
				return nil
			}

			// Load configuration
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

			// If a duration argument is provided, start a new timer
			if len(args) > 0 {
				durationStr := args[0]
				d, err := utils.ParseDuration(durationStr, cfg.DefaultUnits)
				if err != nil {
					return fmt.Errorf("failed to parse duration %q: %w", durationStr, err)
				}

				endTime := time.Now().Add(d)

				// Detached/background mode
				if opts.detached {
					pid, err := startDaemon(durationStr, endTime)
					if err != nil {
						return fmt.Errorf("failed to start background timer: %w", err)
					}
					cmd.Printf("Timer of %s started in background (PID: %d, ending at %s)\n", 
						formatDuration(d), pid, endTime.Format("15:04:05"))
					return nil
				}

				// Foreground mode
				isInteractive := cfg.InteractiveDefault
				if cmd.Flags().Changed("interactive") {
					isInteractive = opts.interactive
				}
				if isInteractive {
					m := timerModel{
						duration:     d,
						remaining:    d,
						lastTickTime: time.Now(),
						endTime:      endTime,
						theme:        theme,
						alarmSound:   cfg.AlarmSound,
						tickInterval: 100 * time.Millisecond,
					}
					p := tea.NewProgram(m)
					if _, err := p.Run(); err != nil {
						return err
					}
				} else {
					// Non-interactive simple countdown
					ticker := time.NewTicker(1 * time.Second)
					defer ticker.Stop()

					remaining := d
					for remaining > 0 {
						fmt.Printf("\rTimer: %s remaining... [Ctrl+C to cancel]", formatDuration(remaining))
						select {
						case <-ticker.C:
							remaining = time.Until(endTime)
						}
					}
					fmt.Println("\n⏰ Time's up!")
					playAlarm(cfg.AlarmSound)
				}
				return nil
			}

			// If no duration argument is provided, show/monitor running timer(s)
			active, err := getActiveTimers()
			if err != nil {
				return err
			}

			if len(active) == 0 {
				cmd.Println("No active timers running. Start one with: timr <duration>")
				return nil
			}

			// Monitor the first active timer
			t := active[0]
			d, _ := utils.ParseDuration(t.Duration, cfg.DefaultUnits)

			isInteractive := cfg.InteractiveDefault
			if cmd.Flags().Changed("interactive") {
				isInteractive = opts.interactive
			}
			if isInteractive {
				m := timerModel{
					duration:     d,
					remaining:    time.Until(t.EndTime),
					endTime:      t.EndTime,
					isMonitor:    true,
					theme:        theme,
					tickInterval: 100 * time.Millisecond,
				}
				p := tea.NewProgram(m)
				if _, err := p.Run(); err != nil {
					return err
				}
			} else {
				// Non-interactive monitor
				ticker := time.NewTicker(1 * time.Second)
				defer ticker.Stop()

				for {
					remaining := time.Until(t.EndTime)
					if remaining <= 0 {
						break
					}
					if !isProcessRunning(t.Pid) {
						cmd.Println("\nTimer process terminated.")
						return nil
					}
					fmt.Printf("\rMonitoring Timer: %s remaining... (PID: %d)", formatDuration(remaining), t.Pid)
					select {
					case <-ticker.C:
					}
				}
				cmd.Println("\n⏰ Timer finished.")
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&opts.configPath, "config", "c", "", "config file path")
	cmd.Flags().BoolVarP(&opts.showVersion, "version", "v", false, "print version information")
	cmd.Flags().BoolVarP(&opts.interactive, "interactive", "i", false, "show live countdown timer TUI")
	cmd.Flags().BoolVarP(&opts.detached, "detached", "d", false, "run timer in background")

	cmd.AddCommand(newConfigCmd())
	cmd.AddCommand(newCompletionCmd())
	cmd.AddCommand(newStopCmd())
	cmd.AddCommand(newDaemonRunCmd())

	return cmd
}

func resolvedVersion() string {
	ver := version
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return ver
	}
	if ver == "dev" && strings.TrimSpace(info.Main.Version) != "" && info.Main.Version != "(devel)" {
		ver = info.Main.Version
	}
	return ver
}

func startDaemon(duration string, endTime time.Time) (int, error) {
	exe, err := os.Executable()
	if err != nil {
		return 0, err
	}

	endTimeStr := endTime.Format(time.RFC3339)
	cmd := exec.Command(exe, "daemon-run", duration, endTimeStr)
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}

	if err := cmd.Start(); err != nil {
		return 0, err
	}
	return cmd.Process.Pid, nil
}

func newDaemonRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "daemon-run [duration] [endTime]",
		Short:  "Run a timer daemon in the background",
		Hidden: true,
		Args:   cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			durationStr := args[0]
			endTimeStr := args[1]

			endTime, err := time.Parse(time.RFC3339, endTimeStr)
			if err != nil {
				return err
			}

			pid := os.Getpid()
			entry := TimerState{
				Pid:       pid,
				StartTime: time.Now(),
				Duration:  durationStr,
				EndTime:   endTime,
			}
			if err := addActiveTimer(entry); err != nil {
				return err
			}

			defer func() {
				_ = removeActiveTimer(pid)
			}()

			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
			go func() {
				<-sigChan
				_ = removeActiveTimer(pid)
				os.Exit(0)
			}()

			cwd, err := os.Getwd()
			var alarmSound string
			if err == nil {
				manager := config.NewManager(cwd)
				if cfg, err := manager.Load(); err == nil {
					alarmSound = cfg.AlarmSound
				}
			}

			time.Sleep(time.Until(endTime))
			playAlarm(alarmSound)

			return nil
		},
	}
	return cmd
}

func newStopCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop any active or background timers",
		RunE: func(cmd *cobra.Command, args []string) error {
			active, err := getActiveTimers()
			if err != nil {
				return err
			}
			if len(active) == 0 {
				cmd.Println("No active or background timers running.")
				return nil
			}

			stoppedCount := 0
			for _, t := range active {
				proc, err := os.FindProcess(t.Pid)
				if err != nil {
					continue
				}
				_ = proc.Signal(syscall.SIGTERM)
				stoppedCount++
				_ = removeActiveTimer(t.Pid)
			}

			cmd.Printf("Stopped %d active background timer(s).\n", stoppedCount)
			return nil
		},
	}
	return cmd
}

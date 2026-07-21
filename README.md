# timr

<img width="480" height="270" alt="screenshot-2026-02-23_16-30-13" src="https://github.com/user-attachments/assets/65386b56-f06f-47be-9063-5c947b30dc51" />

A simple cli timer

## Features

- **Inline TUI timer**: Built with Bubble Tea and Lip Gloss, wrapped in a clean, colored ANSI border.
- **Flexible inputs**: Start timers with `timr 10`, `timr 10m`, `timr 1:10:10`, etc.
- **Detached background timers**: Start background timers using `-d` / `--detached` and stop them with `timr stop`.
- **Tmux integration**: Renames the active Tmux window to show the countdown, and restores it when finished.
- **Alarm sound & control**: Plays a custom sound on completion — `alarm_sound` accepts a single file, a directory (picks random), or a comma-separated list of files/dirs. Falls back to terminal beeps.
- **TUI confirmations**: Prompts for confirmation before cancelling a running timer on exit.
- **Cobra-based CLI**: TOML configuration management, subcommands, and auto-generated shell completion.

## Author's Note

This project was created by deconstructing another cli tool I created: [prompter-cli](https://devan.gg/prompter-cli), a cli tool to organize prompts and skills. I wanted a reusable way to create go cli tools that were fast and looked pretty.

I then used this project to bootstrap [bookmark](https://devan.gg/bookmark), a go based approach to organize shell aliases. I then took what I learned building bookmark and incorporated those learnings back into this project. 

If you're here; those projects may also interest you! :)


## Requirements

- [Go]([https://go.dev/]) for doing the thing.
- [Just](https://just.systems/) for running scripts.
- [Bun](https://bun.sh/) for docs generation. Easily sub for `npm` if preferred.

## Quick start


```bash
# Clone the repo
gh repo clone imdevan/timr-cli
cd timr

# Just build and run
just build-run
```

## Using this as a template

This project is designed to be forked and adapted. The single source of truth for all project metadata is `internal/package/package.toml`. 
<br>
Changing it and running `just sync` propagates your values everywhere.

### Steps

1. Fork or clone the repository.
2. Edit `internal/package/package.toml` with your project details:

```toml
name        = "my-tool"
module      = "github.com/you/my-tool"
description = "What my tool does"
version     = "0.1.0"
repository  = "https://github.com/you/my-tool"
docs_site   = "https://you.github.io"
docs_base   = "/my-tool"
```

3. Run `just sync` to propagate changes.
4. Review the diff with `git diff`.
5. Build and verify: `just build && just test`

### What `just sync` updates

- Go module name in `go.mod` and all import paths throughout `internal/` and `cmd/`
- Binary name in the justfile and build scripts
- Config directory paths in `internal/utils/paths.go`
- Shell completion examples
- README description block
- Version constant in `cmd/*/root.go`

After syncing, add your own commands under `cmd/`, domain types under `internal/domain/`, and UI components under `internal/ui/`.

## Documentation

Documentation is built with [Astro Starlight](https://starlight.astro.build/) and lives in `docs/`. Content is generated automatically from the Go source and project markdown files — you generally don't edit `docs/src/content/docs/` by hand.

However you can easily customize the look of your docs by editing the styles located in `docs/src/styles/custom.css`.

### Go Docs

Go docs are generated via [gomarkdoc](https://github.com/princjef/gomarkdoc)

Into **API Reference**

These docs are intended to be be seen only for the development of the project.

They contain the documentation for the development of the project; they are not needed by users of the cli tool you are building.

### readme, install, config, and contributing

These pages are pulled whole sale from the markdown files in the project. 
Frontmatter is added to play nice with Startlight.

### Commands `cmd`

These pages are generated from go doc comments from the `/cmd` folder which
contains all commands a potential user of the cli tool would use. 

As well as some bash scripting to pull out information on any flag params for a given command.

### How docs are generated

#### just docs-dev

Build the docs and watch for changes.

#### just docs-build

Build docs for production.

#### just docs-generate

Running `just docs-generate` (or implicitly via `just docs-dev` / `just docs-build`) runs `scripts/docs_generate.sh`, which:

1. **Reads `internal/package/package.toml`** and writes `docs/config.mjs` and `docs/sidebar.mjs` with the current project name, description, repository URL, and base path.
2. **Imports markdown files** from the repository root:
   - `README.md` → `docs/src/content/docs/index.md`
   - `INSTALL.md` → `docs/src/content/docs/install.md`
   - `CONFIG.md` → `docs/src/content/docs/configuration.md`
3. **Generates command pages** by parsing each `cmd/<name>/*.go` file for `Use`, `Short`, flags, and godoc comments — one page per command under `docs/src/content/docs/commands/`.
4. **Generates API reference pages** using [gomarkdoc](https://github.com/princjef/gomarkdoc) for every package under `internal/` (including `internal/adapters/*`), outputting to `docs/src/content/docs/api/`.
    - The API reference is only rendered in the side bar during development. 
    - As it is internal and not pertinent to users of the cli tool, but still very helpful for maintainers

### API Reference visibility

The API Reference section is **internal** and only rendered in development by default. When you run `just docs-dev`, the sidebar includes all `internal/` package docs. In a production build (`NODE_ENV=production`), the API Reference is hidden — unless the project name is `go-cli-template`, in which case it is always shown as a live example.

This means when you use this template for your own project, your production docs site will be clean and user-facing, while you still get the full API reference locally during development.

```bash
just docs-dev      # Serves docs at http://localhost:4321 with API reference visible
just docs-build    # Builds production site — API reference hidden for non-template projects
```

## Architecture

Items marked `*` are updated by `just sync`.

```
.
├── go.mod                      # Go packages       *
├── justfile                    # Just run commands *
├── README.md                   # You are here      *
│
├── cmd/                        # CLI commands
│   └── go-cli-template/        # Renamed with just sync      *
│       ├── main.go             # Binary entry point
│       ├── root.go             # Root command, config wiring, app init
│       ├── config.go           # `config` subcommand
│       ├── config_init.go      # `config init` subcommand
│       └── completion.go       # Shell completion subcommand *
│
└── internal/
    ├── package/
    │   └── package.toml        # Source of truth — edit this, then run just sync
    │
    ├── app/                    # Application bootstrap
    ├── config/                 # Loads and parses config.toml
    ├── domain/                 # Core types and data models
    ├── errors/                 # Shared error types
    ├── workflow/               # Business logic layer
    ├── ui/                     # Bubble Tea TUI components
    │   ├── list.go             # Interactive filterable list
    │   ├── theme.go            # Color/style definitions
    │   ├── confirmation.go     # Yes/no prompt
    │   ├── textarea.go         # Multi-line text input
    │   ├── help.go             # Help bar
    │   ├── container.go        # Layout helpers
    │   ├── exit_message.go     # Post-exit message rendering
    │   └── responsive.go       # Terminal size utilities
    ├── adapters/               # Thin wrappers around external interactions
    │   ├── editor/             # Opens files in the user's configured editor
    │   ├── clipboard/          # Read/write system clipboard
    │   ├── shell/              # Shell command execution
    │   ├── tty/                # TTY detection
    │   └── icon/               # Nerd Font icon helpers
    ├── utils/                  # Stateless helpers
    │   ├── paths.go            # XDG config/data/cache path resolution
    │   └── time.go             # Time formatting helpers
    ├── testutil/               # Shared test fixtures and helpers
    └── package/                # Reads package.toml metadata at runtime
```

## Commands

```bash
timr [duration]                 # Start a timer or view running timer(s)
timr stop                       # Stop any active or background timers
timr config                     # View or edit configuration
timr config init                # Generate default config file
timr completion                 # Generate shell completion scripts
```

## Development

### Build & Run

```bash
just build           # Build the binary
just build-run       # Build and run the binary
just dev-build       # Build with debug symbols (disables optimizations)
just watch           # Watch for changes and rebuild automatically
just install         # Install binary to /usr/local/bin
just uninstall       # Remove binary from /usr/local/bin
just clean           # Remove build artifacts (bin/)
```

### Testing

```bash
just test            # Run all tests
just test-verbose    # Run tests with verbose output
```

### Project Sync

```bash
just sync            # Sync all project files from package.toml metadata
```

### Documentation

```bash
just docs-init       # Install documentation dependencies (bun install)
just docs-generate   # Generate API docs and content pages from source
just docs-dev        # Generate docs and start local dev server
just docs-build      # Generate docs and build production site
just docs-preview    # Preview the production build locally
just docs-clean      # Remove generated docs and build artifacts
```

### Package Distribution

```bash
just init-homebrew-tap            # Initialize a Homebrew tap repository
just init-aur-repo                # Initialize an AUR repository
just update-homebrew-formula 1.0.0  # Update Homebrew formula to a version
just update-aur-pkgbuild 1.0.0      # Update AUR PKGBUILD to a version
```

### Tags & Releases

```bash
just tag 1.0.0           # Create and push a git tag
just tag-delete 1.0.0    # Delete a git tag locally and remotely
just tag-list            # List recent tags
just release 1.0.0       # Full release (build, tag, publish)
just github-release 1.0.0   # Create a GitHub release with assets
just deploy-aur 1.0.0       # Deploy to AUR
just deploy-homebrew 1.0.0  # Deploy to Homebrew tap
just deploy-all 1.0.0       # Deploy to all targets
```

## Configuration

Configuration is stored at `$XDG_CONFIG_HOME/timr/config.toml` (typically `~/.config/timr/config.toml`).

### Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `editor` | string | `nvim` | Editor opened by `config` and other editor-aware commands |
| `default_units` | string | `minutes` | Default units to use when duration is a raw number (`seconds`, `minutes`, `hours`) |
| `default_timer` | string | `""` | Default timer duration when run with no arguments (when empty, displays active timer state) |
| `pomodoro` | []int | `[25, 5, 25, 5, 25, 20]` | Array of minutes for Pomodoro sequence (`timr pomodoro` / `timr p`) |
| `alarm_sound` | string | `""` | Path to a file, directory (picks random media file), or comma-separated list of files/dirs |
| `interactive_default` | bool | `true` | Start in interactive TUI mode by default |
| `update_tmux_window` | bool | `false` | When true, rename the active tmux window to the remaining time |
| `tmux_progress_bar` | bool | `true` | When update_tmux_window is true, prefix window title with Nerd Font weather moon icons showing progress |
| `tmux_inverted` | bool | `false` | When true, use inverted moon icon set for tmux progress bar |
| `full_width` | bool | `true` | When true, progress bar and TUI width flex to the full width of the terminal |
| `full_tui` | bool | `true` | When true, render TUI in full window alternate screen mode |
| `rainbow` | bool or []string | `true` | Show an oscillating rainbow progress bar on completion (`true`), disable it (`false`), or pass custom array of color hexes/names |

### `[pomodoro_messages]`

Configures contextual messages for pomodoro phase transitions. Options accept a single string or an array of strings (randomized when selected).

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `after_first_work` | string or []string | `["You're off to a great start!", "One small timer, one big step!"]` | Message after 1st work block |
| `after_first_break` | string or []string | `["You got this!"]` | Message after 1st break block |
| `after_second_work` | string or []string | `["Hell yeah!", "Nice job!"]` | Message after 2nd work block |
| `after_second_break` | string or []string | `[""]` | Message after 2nd break block |
| `before_last_work` | string or []string | `["You're almost there!"]` | Message before starting last work block |
| `after_last_work` | string or []string | `["You did it!"]` | Message after finishing last work block |
| `after_last_break` | string or []string | `["You freaking rock!"]` | Message after finishing last break / final block |
| `list_spacing` | string | `space` | List density: `compact` (title only), `tight` (title + description), `space` (with margins) |
| `time_remaining` | string | `14` | Remaining time color |
| `time_start` | string | `07` | Total/start duration color |
| `bar_bg` | string | `08` | Background part of progress bar |
| `bar_fg` | string or []string | `["02", "03", "01"]` | Foreground/filled part of progress bar (single color or array up to 10 colors for time subdivisions) |
| `help_text` | string | `08` | Key controls help text color |
| `border` | string | `08` | Border color |

Colors accept named values, terminal palette indices, or hex strings (e.g. `7`, `"#ff8800"`).

### How configuration flows through the project

- `internal/config` loads and parses `config.toml` at startup, providing a `Config` struct available to all commands.
- `internal/utils/paths.go` uses XDG paths derived from the project name to locate the config file.
- `internal/ui` reads color and spacing values from config to build Bubble Tea styles — all theme colors come from the loaded config rather than hardcoded constants.
- `internal/adapters/editor` uses the `editor` field to open files in the user's preferred editor.
- The `interactive_default` flag controls whether the root command drops into the TUI automatically or requires an explicit argument.

To generate a config file with defaults:

```bash
timr config init
timr config init --force    # Overwrite existing
timr config init --editor   # Create and open in editor
```

To edit the config directly:

```bash
timr config
```

See `CONFIG.md` for the full reference or `example-config.toml` for a ready-to-copy example.

## Installation

See `INSTALL.md` for installation options.

# Thank you!

This project was made by deconstructing another cli project of mine [Prompter](http://devan.gg/prompter-cli/). Check it out if you like fiddling with coding agents and want a more vim centric way of managing your prompting!

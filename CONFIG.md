# Configuration

Configuration file location: `$XDG_CONFIG_HOME/timr/config.toml`

## Configuration Options

The following options can be set in your configuration file:

### General Settings

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `editor` | string | `nvim` | Editor to use for editing config files |
| `default_units` | string | `minutes` | Default units when raw number is given (`seconds`, `minutes`, `hours`) |
| `alarm_sound` | string | `""` | Path to a file, directory (picks random media file), or comma-separated list of files/dirs |
| `interactive_default` | bool | `true` | Start in interactive mode by default when running a timer |
| `update_tmux_window` | bool | `false` | When true, rename the active tmux window to the remaining time |
| `tmux_progress_bar` | bool | `true` | When update_tmux_window is true, prefix window title with Nerd Font weather moon icons showing progress |
| `tmux_inverted` | bool | `false` | When true, use the inverted moon icon set for the tmux progress bar |
| `full_width` | bool | `true` | When true, progress bar and TUI width flex to the full width of the terminal |
| `rainbow` | bool or []string | `true` | Show an oscillating rainbow progress bar on completion (`true`), disable it (`false`), or pass a custom array of color hexes/names |

### Display Settings

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `list_spacing` | string | `space` | List item spacing. Options: `compact` (title only), `tight` (title + description, no margin), `space` (default, with spacing) |

### Colors

Colors support named, numeric, or hex values (e.g., `7`, `13`, `"#ff8800"`).

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `time_remaining` | string | `14` | Remaining time color |
| `time_start` | string | `07` | Total/start duration color |
| `bar_bg` | string | `08` | Background part of progress bar |
| `bar_fg` | string or []string | `["02", "03", "01"]` | Foreground/filled part of progress bar (single color or array up to 10 colors for time subdivisions) |
| `help_text` | string | `08` | Key controls help text color |
| `border` | string | `08` | Border color |

## Example Configuration

```toml
# General
editor = "nvim"
default_units = "minutes"

# Alarm (pick one style — all handled by alarm_sound):
# alarm_sound = "/path/to/alarm.mp3"       # single file
# alarm_sound = "~/Music/alarms/"          # random file from directory
# alarm_sound = "/a.mp3, /b.mp3, ~/Music/" # CSV list, picks random entry

update_tmux_window = false
tmux_progress_bar = true
full_width = true

# Completed timer animation options:
# rainbow = true                                         # enable rainbow bar (default)
# rainbow = false                                        # disable rainbow bar (blank line)
# rainbow = ["#f5bde6", "#c6a0f6", "#ed8796"]            # custom colors for rainbow bar

# CLI behavior
interactive_default = true

# UI
# list_spacing options: compact (title only), tight (title + description, no margin), space (default, with spacing)
list_spacing = "space"

# Colors
# Colors support named, numeric, or hex values (ex: 7, 13, "#ff8800").
time_remaining = "14"
time_start = "07"
bar_bg = "08"
bar_fg = ["02", "03", "01"]
help_text = "08"
border = "08"
```

## Initializing Configuration

To create a new configuration file with default values:

```bash
timr config init
```

To overwrite an existing configuration:

```bash
timr config init --force
```

To create and immediately open in your editor:

```bash
timr config init --editor
```

## Editing Configuration

To edit your configuration file:

```bash
timr config
```

This will open the config file in your configured editor.

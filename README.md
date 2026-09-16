# VisualRoom (Vroom)

A Go terminal app that scans disk usage, explains it in plain language, and
animates **Bender** as a live mood-indicator of how cramped your disk is.

Default mode is an interactive TUI. Cleanup is **dry-run unless you pass
`--force`**. Files are moved to quarantine, not silently deleted.

Requires **Go 1.22+**.

---

## Build

```bash
git clone https://github.com/sohaib1khan/VisualRoom.git
cd VisualRoom
go build -o vroom .
```

Run tests:

```bash
go test ./...
```

---

## Run

Interactive TUI (default). Scans `$HOME` unless you pass a path:

```bash
./vroom
./vroom ~
./vroom watch /path/to/folder
./vroom --root /path/to/folder
```

One-shot size listing (no TUI):

```bash
./vroom scan ~
./vroom scan . --quiet
```

Preview Bender’s moods without walking a disk:

```bash
./vroom preview
```

Keys in preview: `1`–`5` switch mood, `n` / `p` cycle, `q` quit.

Suggestions in Bender’s voice:

```bash
./vroom suggest ~
```

---

## TUI keys

| Key | Action |
|-----|--------|
| `↑` `↓` / `j` `k` | Move selection |
| `enter` / `l` | Open folder (or file details) |
| `backspace` / `h` | Go up |
| `space` | Toggle mark |
| `d` / `c` | Review marked items for quarantine |
| `i` | Details for the selection |
| `s` | Suggestions |
| `r` | Rescan this folder |
| `?` | Help |
| `q` | Quit |

Bender’s pose follows **free-space %**: relaxed → sarcastic → sweating → panicking → critical.

---

## Cleanup (safe by default)

`vroom clean` **prints what would happen**. Nothing moves until `--force`.

```bash
./vroom clean ./some-file          # dry-run
./vroom clean --force ./some-file  # move to ~/.vroom/quarantine
./vroom restore <batch-id>         # put files back
./vroom quarantine                 # list batches
./vroom purge                      # permanently drop expired batches
./vroom purge --all                # drop every batch
```

Rules baked in:

- Dry-run unless `--force`
- Quarantine under `~/.vroom/quarantine/<timestamp>/` with a 7-day TTL
- Restore with `vroom restore <id>`
- Protected paths (`~/.ssh`, `.git`, system dirs, anything outside `$HOME` by default)
- Never auto-sudo. Running cleanup as root needs `--allow-elevated`
- Audit log at `~/.vroom/audit.log`
- Auto-clean (cron `--mode clean`) only matches explicit whitelist patterns such as `*.tmp`; `node_modules` is flagged, never auto-deleted

---

## Cron / notifications

```bash
./vroom cron install --interval daily --mode notify --engine native
./vroom cron status
./vroom cron remove
./vroom cron run --mode notify
./vroom cron daemon --interval daily --mode notify   # in-process scheduler
```

- **notify** — scan, write `~/.vroom/status.txt`, optional desktop notification
- **clean** — whitelist-only auto-quarantine (still logged)
- **native** — writes a crontab line
- **internal** — run `vroom cron daemon` in the foreground

Intervals: `hourly`, `daily`, `weekly`, or a 5-field cron expression.

---

## Config

On first run Vroom uses built-in defaults and can seed:

```text
~/.vroom/config.yaml
```

A copy of the shipped defaults lives in [`config/default.yaml`](config/default.yaml).
Edit the file under `~/.vroom/` to change thresholds, protected paths, quarantine TTL, and tick rate.

Data Vroom writes:

| Path | Purpose |
|------|---------|
| `~/.vroom/config.yaml` | User config |
| `~/.vroom/cache.json` | Incremental scan cache |
| `~/.vroom/quarantine/` | Quarantined files |
| `~/.vroom/audit.log` | Every scan, quarantine, restore, purge |
| `~/.vroom/status.txt` | Last notify-mode summary |

---

## Docker

Scan a host path read-only (viewing/reporting; destructive clean is blocked on `:ro` mounts):

```bash
docker compose up --build
```

Or:

```bash
docker build -t vroom .
docker run --rm -it -v "$HOME:/scan:ro" -v vroom-data:/root/.vroom vroom watch --root /scan
```

Cleanup from a container needs a **writable** mount **and** `--force`.

---

## Commands cheat sheet

```text
vroom [path]              interactive watch (default)
vroom watch [path]
vroom scan [path]
vroom preview
vroom suggest [path]
vroom clean [paths...]    dry-run; add --force to quarantine
vroom restore <id>
vroom quarantine
vroom purge [--all]
vroom cron install|remove|status|run|daemon
```

Global flags: `--root <path>`, `-q` / `--quiet`.

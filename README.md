# VisualRoom (Vroom)

**Author:** [Sohaib Khan](https://github.com/sohaib1khan)  
**Repository:** [github.com/sohaib1khan/VisualRoom](https://github.com/sohaib1khan/VisualRoom)

A Go terminal app that scans disk usage, explains it in plain language, and
animates a live mood-indicator of how cramped your disk is.

Default mode is an interactive TUI. Cleanup is **dry-run unless you pass
`--force`**. Files are moved to quarantine, not silently deleted.

Source builds need **Go 1.22+**. Distro packages are often older than that
(the `x/exp/slog` / `atomic.Int64` errors). Use the installer below — it
downloads a release binary when one exists, otherwise bootstraps Go 1.22.8
and builds a static Linux binary. You do not need a working Go install.

---

## Install (any Linux)

One line, from anywhere:

```bash
curl -fsSL https://raw.githubusercontent.com/sohaib1khan/VisualRoom/main/install.sh | bash
```

Or from a clone (this is the one to use if `go build` just failed):

```bash
git clone https://github.com/sohaib1khan/VisualRoom.git
cd VisualRoom
chmod +x install.sh
./install.sh
```

That drops `vroom` in `~/.local/bin` (or `/usr/local/bin` if you can write
there). If the command is not found:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

Optional:

```bash
PREFIX=/usr/local/bin sudo -E ./install.sh   # system-wide, if you want
VROOM_GO_VERSION=1.22.8 ./install.sh         # pin the bootstrap toolchain
```

---

## Build (developers)

Need Go 1.22 or newer. Check with `go version`.

```bash
git clone https://github.com/sohaib1khan/VisualRoom.git
cd VisualRoom
make          # -> ./vroom
make test
make install  # -> ~/.local/bin/vroom
```

Equivalent:

```bash
CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o vroom .
```

Do **not** use a distro Go older than 1.22; `./install.sh` or the official
[Go tarball](https://go.dev/dl/) is the supported path.

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

Preview mascot moods without walking a disk:

```bash
./vroom preview
```

Keys in preview: `1`–`5` switch mood, `n` / `p` cycle, `q` quit.

Suggestions:

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

The mascot pose follows **free-space %**: relaxed → sarcastic → sweating → panicking → critical.

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

Global flags: `--root <path>`, `-q` / `--quiet`, `--version`.

---

## Credits

VisualRoom / Vroom is written by **Sohaib Khan**.

- GitHub: [sohaib1khan](https://github.com/sohaib1khan)
- Repository: [https://github.com/sohaib1khan/VisualRoom](https://github.com/sohaib1khan/VisualRoom)

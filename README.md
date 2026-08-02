# agentrule

CLI to install agent instruction rules into a project (or globally). Interactive Bubble Tea wizard, or fast non-interactive install / uninstall.

Writes rules to the agents you use — **Cursor**, **Claude Code**, **Codex CLI**, **opencode**, **Hermes**, **Gemini CLI**, **Antigravity**, **Kiro** — via `--target` / `--location`, same idea as [CodeGraph](https://github.com/colbymchenry/codegraph)'s installer.

Bundled rules (embedded from [`rules/rewrite/`](rules/rewrite/)): **codegraph**, **ponytail** ([DietrichGebert/ponytail](https://github.com/DietrichGebert/ponytail)), **i-have-adhd** ([ayghri/i-have-adhd](https://github.com/ayghri/i-have-adhd)), **powershell** (Windows only).

## Install

**One-liner (recommended):**

```bash
# macOS / Linux
curl -fsSL https://get.mewis.me/agentrule.sh | sh
```

```powershell
# Windows (PowerShell)
irm https://get.mewis.me/agentrule.ps1 | iex
```

Pin a version with `AGENTRULE_VERSION=v0.1.0`. Uninstall on Unix: `sh install.sh --uninstall`.

**Scoop (Windows):**

```powershell
scoop bucket add mew https://github.com/mewisme/scoop-mew
scoop install mew/agentrule
```

Or install the release manifest directly:

```powershell
scoop install https://github.com/mewisme/agentrule/releases/latest/download/agentrule.json
```

**Homebrew (macOS / Linux):**

```bash
brew tap mewisme/tools
brew install --cask agentrule
```

**From source:**

```bash
go install github.com/mewisme/agentrule@latest
```

## Usage

Run in the project you want to configure:

```bash
agentrule                          # interactive 3-step wizard (rules → location → agents)
agentrule install codegraph        # fast path
agentrule i ponytail               # short alias for install
agentrule install codegraph ponytail
agentrule install --all            # all registered rules
agentrule i -a
agentrule i codegraph -u           # force reinstall codegraph binary, then hooks
agentrule i -a --update
agentrule i ponytail --target=all
agentrule i codegraph --target=cursor,claude --location=local
agentrule i codegraph --target=auto --location=global

agentrule uninstall ponytail       # remove rule from agents
agentrule un codegraph             # short alias; also runs codegraph uninstall --keep-cli
agentrule un -a                    # uninstall all rules
agentrule un codegraph --location=global --target=all
```

### Interactive wizard (`agentrule` with no args)

Runs in the alternate screen with the agentrule banner on every view.

**Selection (3 steps)**

1. **Rules** — multi-select which rules to install  
2. **Location** — `local` (this project) or `global` (all projects)  
3. **Agents** — multi-select; detected installs are pre-checked (falls back to Cursor)

**Then install runs inside the TUI** (spinner + progress): success shows a completion screen; failure stays on a failure screen until you exit. `q` / `esc` / `ctrl+c` during install cancels the in-flight step (context cancellation) and exits with an error.

| Key | Action |
|-----|--------|
| `↑` / `↓` / `k` / `j` | Move |
| `space` | Toggle (rules / agents) |
| `a` | Toggle all (rules / agents) |
| `enter` | Next step / start install / dismiss done or failed |
| `←` / `backspace` | Previous selection step (not during install/done/failed) |
| `q` / `esc` / `ctrl+c` | Quit (or cancel install) — `esc` never means “back” |

Non-interactive installs use `--target` / `--location` instead (see below). Their printed timeline is unchanged.

### Flags

| Flag | Values | Default | Meaning |
|------|--------|---------|---------|
| `--target` | `auto`, `all`, or csv (`cursor,claude,…`) | install: `auto` · uninstall: `all` | Which agents. `auto` = detected installs (falls back to Cursor). Uninstall defaults to `all` so leftovers get swept. |
| `--location` | `local`, `global` | `local` | Project-local vs user-wide instruction files. |
| `-u` / `--update` | | | Install only: force reinstall the `codegraph` binary. |
| `-a` / `--all` | | | Every registered rule. |

Codex, Hermes, and Antigravity are **global-only** (skipped when `--location=local`). Hermes / Antigravity have no instruction-file surface for rules — they still participate in `codegraph install` / `uninstall` MCP wiring when selected.

### What each rule does

| Rule | Install | Uninstall |
|------|---------|-----------|
| `codegraph` | ensure binary → `codegraph install` (MCP) → `codegraph init` → write rule to agents | remove rule from agents → `codegraph uninstall --keep-cli` (leaves `.codegraph/` index and the CLI binary) |
| `ponytail` | write rule to agents | remove rule from agents |
| `i-have-adhd` | write rule to agents | remove rule from agents |
| `git-commit` | write rule to agents | remove rule from agents |
| `karpathy-guidelines` | write rule to agents | remove rule from agents |
| `powershell` | write rule to agents (Windows only; hidden on other OS) | remove rule from agents |

### Where rules land

| Agent | Local | Global |
|-------|-------|--------|
| Cursor | `.cursor/rules/{name}.mdc` | — |
| Claude Code | `.claude/CLAUDE.md` (marker section) | `~/.claude/CLAUDE.md` |
| Codex CLI | — | `~/.codex/AGENTS.md` |
| opencode | `./AGENTS.md` | `~/.config/opencode/AGENTS.md` |
| Gemini CLI | `./GEMINI.md` | `~/.gemini/GEMINI.md` |
| Kiro | `.kiro/steering/{name}.md` | `~/.kiro/steering/{name}.md` |

Markdown agents get a marker-fenced section (`<!-- PONYTAIL_START -->` …). Cursor keeps the full `.mdc` (YAML frontmatter + body).

### Authoring rules

Put rewritten rule bodies in [`rules/rewrite/`](rules/rewrite/) — that folder is what `//go:embed` packages into the binary. Register the rule name in [`rules/rule.go`](rules/rule.go).

## Develop

```bash
go test ./...
go build -o agentrule .
```

Release builds use [GoReleaser](https://goreleaser.com/). Tag and push:

```bash
git tag v0.1.0
git push origin v0.1.0
```

## License

[MIT](LICENSE)

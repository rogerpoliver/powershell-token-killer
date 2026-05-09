# AGENTS.md

Guidance for AI agents (Claude Code, Copilot, Cursor, etc.) working in this repository.

## Project Overview

**PTK (PowerShell Token Killer)** is a Go CLI that reduces LLM token consumption when working in Windows/PowerShell environments. It combines two compression layers:

1. **PreToolUse hook** (`ptk-rewrite.ps1`) — rewrites raw PS cmdlets to `ptk <subcmd>` before execution, so the agent sees compressed output instead of verbose PS tables
2. **Caveman hooks** (JS) — inject a system prompt that compresses agent prose by 60-70%

Inspired by [RTK (Rust Token Killer)](https://github.com/rtk-ai/rtk), rebuilt in Go with a PowerShell-first focus for legacy Windows projects that cannot run bash or Unix tooling.

## Working Directory

**ALWAYS confirm before starting work:**

```powershell
pwd          # verify you're in the ptk project root
git branch   # verify correct branch
```

Never assume which project to edit. This repo lives at `~/www/project-ptk`, not `~/www/rtk`.

## Development Commands

### Build & Test

```bash
go build ./...                  # compile check
go build -o ptk .               # build binary
go test ./...                   # all tests (83 tests, 9 packages)
go test ./cmd/ps/ -v            # PS filter tests only
go vet ./...                    # static analysis
```

### Pre-commit Gate (mandatory)

```bash
go build ./... && go vet ./... && go test ./...
```

Never commit code that does not pass all three.

### Install Locally

```bash
go install .                    # installs ptk to $GOPATH/bin
ptk --version
ptk init                        # installs hooks into .claude/
```

## Architecture

PTK uses a **two-stage rewrite pipeline**:

```
Agent issues: "gci src/"
  ↓
ptk-rewrite.ps1 (PreToolUse hook)
  ↓
ptk rewrite "gci src/"
  ↓ aliases.Normalize: "gci" → "Get-ChildItem"
  ↓ registry.Rewrite: "Get-ChildItem" → "ptk gci"
  ↓ permissions.Check: allow/deny/ask verdict
  ↓ exit 0 (auto-allow) + stdout "ptk gci src/"
  ↓
ptk gci src/
  ↓ InvokePwshJSON("Get-ChildItem src/ | Select-Object Name,Length,Attributes")
  ↓ formatGCI() — noise dirs filtered, human sizes, summary
  ↓ compressed output (~75% savings)
```

### Module Map

```
main.go                             entry point
cmd/
├── root.go                         cobra root, --version
├── rewrite.go                      ptk rewrite — hook protocol (exit 0/1/2/3)
├── proxy.go                        ptk proxy — passthrough + tracking
├── gain.go                         ptk gain — token savings report
├── init_cmd.go                     ptk init — installs hooks via embedded.FS
└── ps/
    ├── gci.go                      Get-ChildItem filter (JSON→compact table)
    ├── sls.go                      Select-String filter (prefix strip, truncate)
    ├── gps.go                      Get-Process filter (JSON→CPU-sorted table)
    ├── gsv.go                      Get-Service filter (tabular→compact)
    ├── measure.go                  Measure-Object filter (drop null properties)
    ├── history.go                  Get-History filter (strip PS table header)
    └── help.go                     Get-Help filter (synopsis only, ~95% savings)
internal/
├── powershell/
│   ├── aliases.go                  PS alias → canonical cmdlet (gci→Get-ChildItem)
│   └── invoke.go                   InvokePwsh/InvokePwshJSON/InvokePwshText
├── registry/
│   └── registry.go                 canonical cmdlet → ptk subcommand
├── permissions/
│   └── permissions.go              allow/deny/ask verdict (deny > ask > allow)
├── config/
│   └── config.go                   TOML config loader + defaults
├── tracking/
│   └── tracking.go                 SQLite token savings metrics
└── embedded/
    └── embedded.go                 go:embed for hooks/ and skills/
hooks/
├── claude/ptk-rewrite.ps1          PreToolUse hook (PowerShell native, no bash)
├── caveman-activate.js             SessionStart — injects caveman system prompt
├── caveman-mode-tracker.js         UserPromptSubmit — maintains caveman mode
└── ...
```

## Hook Protocol (exit codes)

`ptk rewrite <cmd>` follows RTK's protocol:

| Exit code | Meaning | ps1 action |
|-----------|---------|------------|
| `0` + stdout | Rewrite found, allow rule matched → auto-allow | emit JSON with `permissionDecision: "allow"` |
| `1` | No PTK equivalent → passthrough | `exit 0` (Claude handles normally) |
| `2` | Deny rule matched | `exit 0` (Claude native deny takes over) |
| `3` + stdout | Rewrite found, ask rule → prompt user | emit JSON without `permissionDecision` |

## Permission Rules

Priority order: **deny > ask > allow > default**

Defaults in `internal/config/config.go`:
- **Allow**: read-only cmdlets (`Get-ChildItem *`, `git *`, `go *`, etc.) → exit 0
- **Ask**: destructive but recoverable (`Stop-Process *`, `Remove-Item * -Recurse`) → exit 3
- **Deny**: irreversible (`Remove-Item * -Recurse -Force`, `Format-Volume`) → exit 2

Override in `~/.config/ptk/config.toml` (Unix) or `%APPDATA%\ptk\config.toml` (Windows).

## Adding a New PS Filter

1. Create `cmd/ps/<cmdlet>.go` — implement `filter<Cmdlet>(raw string) string`
2. Add cobra command, wire into `cmd/root.go`
3. Add canonical cmdlet → `ptk <subcmd>` mapping in `internal/registry/registry.go`
4. Add PS aliases in `internal/powershell/aliases.go` if needed
5. Write tests: format correctness + token savings assertion
6. Run `go build ./... && go vet ./... && go test ./...`

### Savings Targets

| Filter type | Target | Mechanism |
|-------------|--------|-----------|
| JSON cmdlets (gps, gci) | ≥55% | Compact table vs verbose JSON |
| Text-tabular (gsv) | ≥5% | Header/separator removal |
| Null-heavy output (measure) | ≥70% | Drop null properties |
| Full help pages (phelp) | ≥80% | Synopsis section only |
| Long-line output (sls) | ≥25% | Truncate at 120 chars |

## Coding Rules

- **No `unwrap()`** — use `err` returns with context messages
- **Fallback pattern** — if filter fails, print raw output unchanged (never block the agent)
- **No CGo** — `modernc.org/sqlite` for pure-Go SQLite; binary must cross-compile
- **No async** — single-threaded, startup target <50ms
- **PowerShell cmdlets are not binaries** — never `exec.LookPath("Get-ChildItem")`; always invoke via `pwsh -NoProfile -NonInteractive -Command`
- **Exit code propagation** — if underlying pwsh fails, propagate exit code

## Testing Rules

- Every filter must have: format correctness test + savings assertion
- Savings thresholds must reflect what the filter actually achieves (don't inflate)
- Use `countTokens(s string) int { return len(strings.Fields(s)) }` for savings math
- No mocking of PowerShell — unit tests operate on pre-captured string output

## Config File

```toml
# ~/.config/ptk/config.toml  (Unix)
# %APPDATA%\ptk\config.toml  (Windows)

[hooks]
allow = ["Get-ChildItem *", "git *", "go *"]   # exit 0 auto-allow
deny  = ["Remove-Item * -Recurse -Force"]       # exit 2 block
ask   = ["Stop-Process *", "Remove-Item * -Recurse"]  # exit 3 confirm

[powershell]
prefer_pwsh = true   # pwsh (PS7+) over powershell.exe (PS5)
max_items   = 100    # max entries in filter output

[tracking]
enabled = true       # SQLite savings metrics
```

## Commit Convention

Conventional Commits enforced via commitlint + GitHub Actions:

```
feat(ps): add Get-EventLog filter
fix(rewrite): preserve tool_input fields on PS hook rewrite
test(ps): add savings test for gsv filter
docs: update AGENTS.md with new filter guide
```

Types: `feat fix docs style refactor perf test build ci chore revert`

Subject must be **lower-case**. Validated by `commit-msg` hook and CI.

## Git Hooks

Install before committing:

```bash
mise run hooks   # or: lefthook install
```

`pre-commit` — runs `go build ./...`, `go vet ./...`, `go test ./...`. Commit blocked if any fails.

`commit-msg` — validates conventional commit format (pure shell, no npm needed). Blocks non-conforming messages.

## No AI Co-Authors

**Never add Co-Authored-By, co-authored-by, or any AI attribution to commits.**

This applies to all AI tools: Claude, Copilot, Kiro, Gemini, OpenCode, or any other.
Commits must be clean — no AI signatures in the git history.

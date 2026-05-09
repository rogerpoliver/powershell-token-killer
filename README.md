# PTK — PowerShell Token Killer

> [!WARNING]
> **This is a proof of concept.** PTK is being actively tested and validated — nothing here guarantees this becomes a production-ready tool. If the idea resonates with you, leave a ⭐ and check back in a few days.

> Reduce LLM token consumption by 60–90% when working in Windows/PowerShell environments with AI coding agents.

PTK is a Go CLI that sits between your AI agent and PowerShell, compressing verbose cmdlet output before it reaches the context window. Built for legacy Windows projects that depend exclusively on PowerShell and cannot run bash, Unix tools, or Node-based proxies.

Inspired by [RTK (Rust Token Killer)](https://github.com/rtk-ai/rtk) — rebuilt from scratch in Go with a PowerShell-first architecture.

---

## The Problem

AI agents working on Windows-only legacy codebases face a unique problem: PowerShell cmdlet output is extremely verbose. A single `Get-ChildItem` produces 8–12 tokens per file entry (Mode, LastWriteTime, Length, Name). `Get-Help` dumps hundreds of lines. `Measure-Object` lists 7 properties, 6 of which are null.

Tools like RTK solve this for Unix workflows, but they depend on bash and jq — which don't exist on native Windows environments. PowerShell has no equivalent hook ecosystem out of the box.

PTK fills that gap.

---

## How It Works

Two compression layers work together:

```
┌─────────────────────────────────────────────────────┐
│  Layer 1: Command Output Compression                │
│                                                     │
│  Agent: "gci src/"                                  │
│    ↓ ptk-rewrite.ps1 (PreToolUse hook)              │
│    ↓ rewrites to: "ptk gci src/"                    │
│    ↓ ptk invokes pwsh, filters JSON output          │
│  Agent sees: compact table, ~75% fewer tokens       │
└─────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────┐
│  Layer 2: Agent Prose Compression (Caveman)         │
│                                                     │
│  SessionStart hook injects system prompt            │
│  Agent replies in terse caveman style               │
│  Result: ~65% fewer tokens in responses             │
└─────────────────────────────────────────────────────┘
```

---

## Token Savings

| Command | Raw tokens | PTK tokens | Savings |
|---------|-----------|------------|---------|
| `Get-ChildItem` (20 items) | ~280 | ~65 | **77%** |
| `Get-Process` (10 processes, JSON) | ~103 | ~44 | **57%** |
| `Get-Help Get-ChildItem` | ~420 | ~18 | **95%** |
| `Measure-Object -Line` | ~28 | ~4 | **86%** |
| `Get-Service` (14 services) | ~82 | ~75 | **8%** |
| `Get-History` | minimal | minimal | format only |

---

## Supported Cmdlets

| PTK command | Replaces | Strategy |
|-------------|----------|----------|
| `ptk gci` | `Get-ChildItem`, `ls`, `dir`, `gci` | JSON→compact table, filters noise dirs |
| `ptk sls` | `Select-String`, `grep`, `sls` | strip filename prefix, truncate long lines |
| `ptk gps` | `Get-Process`, `ps`, `gps` | JSON→CPU-sorted table, human sizes |
| `ptk gsv` | `Get-Service`, `gsv` | strip header/separator, truncate display names |
| `ptk measure` | `Measure-Object`, `measure`, `wc` | drop null properties |
| `ptk history` | `Get-History`, `history` | strip PS table header |
| `ptk phelp` | `Get-Help`, `man` | synopsis section only (~95% savings) |

Non-PTK commands (`git`, `go`, `cargo`, etc.) pass through unchanged. PTK never blocks them.

---

## Installation

### Prerequisites

- Go 1.22+
- PowerShell 7+ (`pwsh`) recommended, PowerShell 5 (`powershell.exe`) supported
- [Claude Code](https://claude.ai/code) or another AI agent that supports Claude hooks

### Build from source

```bash
git clone https://github.com/rogerpoliver/powershell-token-killer.git
cd powershell-token-killer
go build -o ptk .
```

Move `ptk` (or `ptk.exe` on Windows) to a directory in your PATH.

### Install hooks

```bash
ptk init
```

This extracts hook files from the binary and patches `.claude/settings.json` (project-local) or `~/.claude/settings.json` (global). Restart Claude Code to activate.

---

## Configuration

PTK creates `~/.config/ptk/config.toml` (Unix) or `%APPDATA%\ptk\config.toml` (Windows) on first `ptk init`.

```toml
[hooks]
# Commands to auto-allow (agent skips confirmation prompt)
allow = [
  "Get-ChildItem *", "gci *", "ls *", "dir *",
  "Select-String *", "sls *", "grep *",
  "Get-Process", "Get-Process *",
  "git *", "go *", "cargo *",
]

# Commands blocked entirely
deny = [
  "Remove-Item * -Recurse -Force",
  "Format-Volume",
  "Set-ExecutionPolicy Bypass",
]

# Commands that require confirmation
ask = [
  "Stop-Process *",
  "Remove-Item * -Recurse",
  "Set-ExecutionPolicy *",
]

[powershell]
prefer_pwsh = true   # prefer pwsh (PS7+) over powershell.exe
max_items   = 100

[tracking]
enabled = true       # token savings metrics in SQLite
```

---

## Usage

### Direct commands

```powershell
ptk gci src/             # compressed Get-ChildItem
ptk gps                  # compressed Get-Process (top 20 by CPU)
ptk gsv -Status Running  # compressed Get-Service
ptk phelp Invoke-WebRequest  # synopsis only
ptk gain                 # token savings report
```

### Via hook (automatic)

Once `ptk init` is run, Claude Code rewrites commands automatically:

```
Agent types: gci .
Hook sees:   Get-ChildItem .
Rewrites to: ptk gci .
Agent sees:  src/  cmd/  internal/  (compressed, noise dirs hidden)
             Summary: 14 files, 7 dirs (.go ×11, .toml ×1, .md ×2)
```

### Proxy mode (passthrough with tracking)

```bash
ptk proxy Get-WinEvent -LogName System -Newest 50
```

Runs the command without PTK filtering but records it in the savings database.

---

## Architecture

```
main.go
cmd/
├── rewrite.go          hook protocol: alias normalize → registry → permissions
├── proxy.go            passthrough with tracking
├── gain.go             savings report from SQLite
├── init_cmd.go         hook installation via go:embed
└── ps/                 PS cmdlet filter implementations
internal/
├── powershell/         pwsh invocation + alias normalization
├── registry/           cmdlet → ptk subcommand mapping
├── permissions/        allow/deny/ask rule evaluation
├── config/             TOML config + defaults
├── tracking/           SQLite token metrics (CGo-free)
└── embedded/           baked-in hooks and skills (go:embed)
hooks/                  Claude Code hook scripts
skills/                 Agent skill files (Caveman behavior rules)
```

No CGo. No external runtime dependencies. Single binary.

---

## Comparison with RTK

| | RTK | PTK |
|---|-----|-----|
| Language | Rust | Go |
| Target OS | macOS / Linux | Windows-first, cross-platform |
| Hook script | bash + jq | PowerShell native (no bash needed) |
| Ecosystem | git, cargo, go, npm, docker, … | PowerShell cmdlets |
| Agent prose compression | No | Yes (Caveman layer) |
| Legacy Windows support | Limited | Primary focus |

PTK was born from the same idea as RTK — compress what the agent sees, not what it does — but adapted for environments where bash is not available and PowerShell is the only shell.

---

## Development

```bash
go build ./... && go vet ./... && go test ./...
```

See [AGENTS.md](AGENTS.md) for architecture details, coding rules, and how to add new filters.

---

## License

MIT

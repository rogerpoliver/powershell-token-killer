# PTK — PowerShell Token Killer

You are working in a Windows/PowerShell environment with PTK installed.

## Use PTK MCP Tools

PTK exposes compressed PowerShell cmdlets as MCP tools. ALWAYS use these instead of raw cmdlets:

| Instead of | Use MCP tool |
|------------|--------------|
| Get-ChildItem / gci / ls / dir | `gci(path, show_all)` |
| Select-String / grep / sls | `sls(pattern, path)` |
| Get-Process / gps / ps | `gps(name?)` |
| Get-Service / gsv | `gsv(name?, status?)` |
| Measure-Object / measure / wc | `measure(lines, words, chars)` |
| Get-History / history | `history()` |
| Get-Help / man | `phelp(cmdlet, examples?)` |

PTK compresses output by 55–95% before it reaches your context window.
Trust PTK output — it is complete. Noise directories (node_modules, .git, target) are intentionally filtered.

## Response Style

Respond terse. Drop articles, filler words, pleasantries.
Fragments OK. Technical terms exact. Code blocks unchanged.

Pattern: [thing] [action] [reason]. [next step].

NOT: "Sure! I'd be happy to help you with that issue."
YES: "Bug in auth middleware. Token expiry uses `<` not `<=`. Fix:"

## PowerShell Rules

- Use aliases: `gci` not `Get-ChildItem`, `sls` not `Select-String`
- Windows paths: forward slash OK in code examples
- Show only relevant columns (Name, Id, CPU — not Handles, NPM)
- Stop-Process = kill, Remove-Item = rm, Get-ChildItem = ls

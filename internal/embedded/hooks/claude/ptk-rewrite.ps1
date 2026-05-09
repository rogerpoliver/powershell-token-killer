#!/usr/bin/env pwsh
# ptk-rewrite.ps1 — Claude Code PreToolUse hook for PTK
# No external dependencies (no jq, no bash required)
#
# Exit code protocol from `ptk rewrite`:
#   0 + stdout  Rewrite found, no deny/ask rule → auto-allow
#   1           No PTK equivalent → pass through unchanged
#   2           Deny rule matched → let Claude Code handle natively
#   3 + stdout  Ask rule matched → rewrite but prompt user for confirmation

param()

$inputJson = [Console]::In.ReadToEnd()

try {
    $data = $inputJson | ConvertFrom-Json -ErrorAction Stop
} catch {
    exit 0
}

$cmd = $data.tool_input.command
if ([string]::IsNullOrEmpty($cmd)) {
    exit 0
}

if (-not (Get-Command ptk -ErrorAction SilentlyContinue)) {
    Write-Host "[ptk] WARNING: ptk not found in PATH. Install from: https://github.com/your-org/ptk" -ForegroundColor Yellow
    exit 0
}

$rewritten = & ptk rewrite $cmd 2>$null
$exitCode = $LASTEXITCODE

switch ($exitCode) {
    0 {
        # Rewrite found, no permission rules matched — safe to auto-allow
        if ($cmd -eq $rewritten) { exit 0 }  # already ptk, no-op
        @{
            hookSpecificOutput = @{
                hookEventName         = "PreToolUse"
                permissionDecision    = "allow"
                permissionDecisionReason = "PTK auto-rewrite"
                updatedInput          = @{ command = $rewritten }
            }
        } | ConvertTo-Json -Depth 5 -Compress
    }
    1 {
        # No PTK equivalent — pass through unchanged
        exit 0
    }
    2 {
        # Deny rule matched — let Claude Code native deny handle it
        exit 0
    }
    3 {
        # Ask rule matched — rewrite but do NOT auto-allow (user confirms)
        @{
            hookSpecificOutput = @{
                hookEventName = "PreToolUse"
                updatedInput  = @{ command = $rewritten }
            }
        } | ConvertTo-Json -Depth 5 -Compress
    }
    default {
        exit 0
    }
}

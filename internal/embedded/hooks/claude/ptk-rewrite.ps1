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

# Detect format: Claude Code / VS Code Copilot Chat vs Copilot CLI
$cmd = $null
$isCopilotCLI = $false

if ($data.tool_input -and -not [string]::IsNullOrEmpty($data.tool_input.command)) {
    # Claude Code / VS Code Copilot Chat: snake_case tool_input.command
    $cmd = $data.tool_input.command
} elseif ($data.toolName -and $data.toolArgs) {
    # GitHub Copilot CLI: camelCase toolName + toolArgs (JSON string)
    $isCopilotCLI = $true
    try {
        $toolArgs = $data.toolArgs | ConvertFrom-Json -ErrorAction Stop
        $cmd = $toolArgs.command
    } catch {
        exit 0
    }
}

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

        if ($isCopilotCLI) {
            # Copilot CLI does not support updatedInput — use deny-with-suggestion
            @{
                permissionDecision       = "deny"
                permissionDecisionReason = "Token savings: use ``$rewritten`` instead (PTK saves ~75% tokens)"
            } | ConvertTo-Json -Compress
            exit 0
        }

        $data.tool_input.command = $rewritten
        @{
            hookSpecificOutput = @{
                hookEventName            = "PreToolUse"
                permissionDecision       = "allow"
                permissionDecisionReason = "PTK auto-rewrite"
                updatedInput             = $data.tool_input
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
        if ($isCopilotCLI) {
            @{
                permissionDecision       = "deny"
                permissionDecisionReason = "Token savings: use ``$rewritten`` instead (PTK saves ~75% tokens)"
            } | ConvertTo-Json -Compress
            exit 0
        }

        $data.tool_input.command = $rewritten
        @{
            hookSpecificOutput = @{
                hookEventName = "PreToolUse"
                updatedInput  = $data.tool_input
            }
        } | ConvertTo-Json -Depth 5 -Compress
    }
    default {
        exit 0
    }
}

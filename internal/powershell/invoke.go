package powershell

import (
	"fmt"
	"os/exec"
	"strings"
)

// shell returns "pwsh" if PowerShell 7+ is available, else "powershell" (PS 5).
func shell() string {
	if _, err := exec.LookPath("pwsh"); err == nil {
		return "pwsh"
	}
	return "powershell"
}

// Available reports whether any PowerShell (pwsh or powershell) is on PATH.
func Available() bool {
	_, pwsh := exec.LookPath("pwsh")
	_, ps := exec.LookPath("powershell")
	return pwsh == nil || ps == nil
}

// InvokePwsh runs a PowerShell script and returns stdout bytes.
// Uses pwsh (PS 7+) if available, falls back to powershell (PS 5).
func InvokePwsh(script string) ([]byte, error) {
	sh := shell()
	out, err := exec.Command(sh,
		"-NoProfile", "-NonInteractive", "-Command", script,
	).Output()
	if err != nil {
		return nil, fmt.Errorf("powershell exec: %w", err)
	}
	return out, nil
}

// InvokePwshJSON runs a cmdlet and appends "| ConvertTo-Json -Depth 3 -Compress".
// Returns JSON bytes.
func InvokePwshJSON(cmdlet string) ([]byte, error) {
	script := cmdlet + " | ConvertTo-Json -Depth 3 -Compress"
	return InvokePwsh(script)
}

// InvokePwshText runs a cmdlet and returns trimmed text output.
func InvokePwshText(cmdlet string) (string, error) {
	out, err := InvokePwsh(cmdlet)
	if err != nil {
		return "", err
	}
	return strings.TrimRight(string(out), "\r\n"), nil
}

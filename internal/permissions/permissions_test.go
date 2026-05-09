package permissions

import (
	"testing"

	"ptk/internal/config"
)

func testCfg() config.HooksConfig {
	return config.HooksConfig{
		Deny: []string{
			"Remove-Item * -Recurse -Force",
			"Format-Volume",
			"Set-ExecutionPolicy Bypass",
		},
		Ask: []string{
			"Stop-Process *",
			"Remove-Item * -Recurse",
			"Set-ExecutionPolicy *",
		},
	}
}

func TestCheck_deny(t *testing.T) {
	cfg := testCfg()
	tests := []string{
		"Remove-Item * -Recurse -Force",
		"Format-Volume",
		"Set-ExecutionPolicy Bypass",
	}
	for _, cmd := range tests {
		if v := Check(cmd, cfg); v != VerdictDeny {
			t.Errorf("Check(%q) = %v, want VerdictDeny", cmd, v)
		}
	}
}

func TestCheck_ask(t *testing.T) {
	cfg := testCfg()
	tests := []string{
		"Stop-Process chrome",
		"Remove-Item src/ -Recurse",
		"Set-ExecutionPolicy RemoteSigned",
	}
	for _, cmd := range tests {
		if v := Check(cmd, cfg); v != VerdictAsk {
			t.Errorf("Check(%q) = %v, want VerdictAsk", cmd, v)
		}
	}
}

func TestCheck_default(t *testing.T) {
	cfg := testCfg()
	tests := []string{
		"Get-ChildItem .",
		"git status",
		"cargo build",
		"gci src/",
	}
	for _, cmd := range tests {
		if v := Check(cmd, cfg); v != VerdictDefault {
			t.Errorf("Check(%q) = %v, want VerdictDefault", cmd, v)
		}
	}
}

func TestMatchesPattern(t *testing.T) {
	tests := []struct {
		cmd     string
		pattern string
		want    bool
	}{
		{"Stop-Process chrome", "Stop-Process *", true},
		{"stop-process chrome", "Stop-Process *", true},   // case-insensitive
		{"Format-Volume", "Format-Volume", true},
		{"Format-Volume C:", "Format-Volume", true},       // prefix match
		{"Get-ChildItem", "Format-Volume", false},
		{"Remove-Item src/ -Recurse", "Remove-Item * -Recurse", true},
	}
	for _, tt := range tests {
		got := matchesPattern(tt.cmd, tt.pattern)
		if got != tt.want {
			t.Errorf("matchesPattern(%q, %q) = %v, want %v", tt.cmd, tt.pattern, got, tt.want)
		}
	}
}

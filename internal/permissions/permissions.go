package permissions

import (
	"strings"

	"ptk/internal/config"
)

// Verdict represents the permission decision for a command.
type Verdict int

const (
	VerdictDefault Verdict = iota // no rule matched — defer to agent
	VerdictAllow                  // explicit allow rule matched
	VerdictAsk                    // ask rule matched
	VerdictDeny                   // deny rule matched
)

// Check evaluates a command against the configured deny/ask rules.
func Check(cmd string, cfg config.HooksConfig) Verdict {
	for _, pattern := range cfg.Deny {
		if matchesPattern(cmd, pattern) {
			return VerdictDeny
		}
	}
	for _, pattern := range cfg.Ask {
		if matchesPattern(cmd, pattern) {
			return VerdictAsk
		}
	}
	return VerdictDefault
}

// matchesPattern does simple glob-like matching: "*" matches any suffix.
func matchesPattern(cmd, pattern string) bool {
	cmd = strings.ToLower(strings.TrimSpace(cmd))
	pattern = strings.ToLower(strings.TrimSpace(pattern))
	if !strings.Contains(pattern, "*") {
		return strings.HasPrefix(cmd, pattern)
	}
	// Split on * and check each part appears in order
	parts := strings.Split(pattern, "*")
	pos := 0
	for i, part := range parts {
		if part == "" {
			continue
		}
		idx := strings.Index(cmd[pos:], part)
		if idx == -1 {
			return false
		}
		if i == 0 && idx != 0 {
			return false // first part must match from start
		}
		pos += idx + len(part)
	}
	return true
}

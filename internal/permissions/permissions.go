package permissions

import (
	"strings"

	"ptk/internal/config"
)

// Verdict represents the permission decision for a command.
type Verdict int

const (
	VerdictDefault Verdict = iota // no rule matched — defer to agent
	VerdictAllow                  // explicit allow rule matched → exit 0
	VerdictAsk                    // ask rule matched → exit 3
	VerdictDeny                   // deny rule matched → exit 2
)

// Check evaluates cmd against deny → ask → allow rules, in that priority order.
// Deny takes precedence over everything; allow beats the default ask.
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
	for _, pattern := range cfg.Allow {
		if matchesPattern(cmd, pattern) {
			return VerdictAllow
		}
	}
	return VerdictDefault
}

// matchesPattern does case-insensitive glob matching where "*" matches any suffix.
// A pattern without "*" matches if cmd starts with the pattern.
func matchesPattern(cmd, pattern string) bool {
	cmd = strings.ToLower(strings.TrimSpace(cmd))
	pattern = strings.ToLower(strings.TrimSpace(pattern))

	if !strings.Contains(pattern, "*") {
		// Exact prefix match
		return cmd == pattern || strings.HasPrefix(cmd, pattern+" ")
	}

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
			return false // first segment must match from start
		}
		pos += idx + len(part)
	}
	return true
}

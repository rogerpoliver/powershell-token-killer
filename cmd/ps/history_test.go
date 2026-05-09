package ps

import (
	"strings"
	"testing"
)

func TestFilterHistory_basic(t *testing.T) {
	raw := `  Id CommandLine
  -- -----------
   1 Get-ChildItem
   2 cargo build --release
   3 git status
`
	out := FilterHistory(raw)
	if !strings.Contains(out, "1  Get-ChildItem") {
		t.Errorf("expected '1  Get-ChildItem', got:\n%s", out)
	}
	if !strings.Contains(out, "2  cargo build --release") {
		t.Errorf("expected '2  cargo build --release', got:\n%s", out)
	}
	if strings.Contains(out, "CommandLine") {
		t.Error("header should be stripped")
	}
	if strings.Contains(out, "---") {
		t.Error("separator should be stripped")
	}
}

func TestFilterHistory_empty(t *testing.T) {
	out := FilterHistory("")
	if out != "(no history)\n" {
		t.Errorf("expected (no history)\\n, got %q", out)
	}
}

func TestFilterHistory_strips_header(t *testing.T) {
	// History output is already compact (id + command) so savings are minimal.
	// The value is format normalization, not compression ratio.
	raw := `  Id CommandLine
  -- -----------
   1 Get-ChildItem -Path . -Force
   2 cargo build --release
  10 Get-Service | Where-Object Status -eq Running
`
	out := FilterHistory(raw)
	// Header and separator must be gone
	if strings.Contains(out, "CommandLine") || strings.Contains(out, "---") {
		t.Errorf("header/separator should be stripped, got:\n%s", out)
	}
	// All 3 entries must appear
	if !strings.Contains(out, "1  Get-ChildItem") {
		t.Errorf("entry 1 missing, got:\n%s", out)
	}
	if !strings.Contains(out, "10  Get-Service") {
		t.Errorf("entry 10 missing, got:\n%s", out)
	}
}

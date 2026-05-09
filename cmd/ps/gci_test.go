package ps

import (
	"strings"
	"testing"
)

func TestFormatGCI_basic(t *testing.T) {
	entries := []gciEntry{
		{Name: "src", Attributes: "Directory"},
		{Name: "main.go", Length: 4234, Attributes: "Archive"},
		{Name: "go.mod", Length: 800, Attributes: "Archive"},
		{Name: "README.md", Length: 2100, Attributes: "Archive"},
	}
	out := formatGCI(entries, false)
	if !strings.Contains(out, "src/") {
		t.Error("expected src/ in output")
	}
	if !strings.Contains(out, "main.go") {
		t.Error("expected main.go in output")
	}
	if !strings.Contains(out, "4.1K") {
		t.Errorf("expected human size 4.1K, got:\n%s", out)
	}
	if strings.Contains(out, "Directory") || strings.Contains(out, "Archive") {
		t.Error("attributes should not appear in output")
	}
}

func TestFormatGCI_noise_filtered(t *testing.T) {
	entries := []gciEntry{
		{Name: "node_modules", Attributes: "Directory"},
		{Name: ".git", Attributes: "Directory"},
		{Name: "target", Attributes: "Directory"},
		{Name: "src", Attributes: "Directory"},
		{Name: "main.go", Length: 1000, Attributes: "Archive"},
	}
	out := formatGCI(entries, false)
	if strings.Contains(out, "node_modules") {
		t.Error("node_modules should be filtered")
	}
	if strings.Contains(out, ".git/") {
		t.Error(".git should be filtered")
	}
	if strings.Contains(out, "target/") {
		t.Error("target should be filtered")
	}
	if !strings.Contains(out, "src/") {
		t.Error("src/ should appear")
	}
}

func TestFormatGCI_showAll(t *testing.T) {
	entries := []gciEntry{
		{Name: "node_modules", Attributes: "Directory"},
		{Name: "src", Attributes: "Directory"},
	}
	out := formatGCI(entries, true)
	if !strings.Contains(out, "node_modules") {
		t.Error("with showAll, node_modules should appear")
	}
}

func TestFormatGCI_empty(t *testing.T) {
	out := formatGCI([]gciEntry{}, false)
	if out != "(empty)\n" {
		t.Errorf("expected (empty)\\n, got %q", out)
	}
}

func TestFormatGCI_token_savings(t *testing.T) {
	// Simulate realistic PS output for a project with many files.
	// PS table format: Mode(5) + spaces + LastWriteTime(19) + spaces + Length(8) + spaces + Name
	// Each row ~12 tokens. Header + separator ~8 tokens. Directory banner ~4 tokens.
	psVerboseOutput := `
    Directory: C:\Users\user\project

Mode                 LastWriteTime         Length Name
----                 -------------         ------ ----
d----          2024-01-01    12:00                src
d----          2024-01-01    12:00                cmd
d----          2024-01-01    12:00                internal
d----          2024-01-01    12:00                node_modules
d----          2024-01-01    12:00                .git
d----          2024-01-01    12:00                target
-a---          2024-01-01    12:00           1234 main.go
-a---          2024-01-01    12:00           4567 server.go
-a---          2024-01-01    12:00           2345 handler.go
-a---          2024-01-01    12:00           1890 middleware.go
-a---          2024-01-01    12:00            800 go.mod
-a---          2024-01-01    12:00          12000 go.sum
-a---          2024-01-01    12:00           2100 README.md
-a---          2024-01-01    12:00            450 Makefile
-a---          2024-01-01    12:00           3200 Dockerfile
-a---          2024-01-01    12:00            650 .gitignore
-a---          2024-01-01    12:00            320 .env.example
-a---          2024-01-01    12:00           8900 package-lock.json
-a---          2024-01-01    12:00           1200 tsconfig.json
-a---          2024-01-01    12:00            780 jest.config.js
`
	entries := []gciEntry{
		{Name: "src", Attributes: "Directory"},
		{Name: "cmd", Attributes: "Directory"},
		{Name: "internal", Attributes: "Directory"},
		{Name: "node_modules", Attributes: "Directory"},
		{Name: ".git", Attributes: "Directory"},
		{Name: "target", Attributes: "Directory"},
		{Name: "main.go", Length: 1234, Attributes: "Archive"},
		{Name: "server.go", Length: 4567, Attributes: "Archive"},
		{Name: "handler.go", Length: 2345, Attributes: "Archive"},
		{Name: "middleware.go", Length: 1890, Attributes: "Archive"},
		{Name: "go.mod", Length: 800, Attributes: "Archive"},
		{Name: "go.sum", Length: 12000, Attributes: "Archive"},
		{Name: "README.md", Length: 2100, Attributes: "Archive"},
		{Name: "Makefile", Length: 450, Attributes: "Archive"},
		{Name: "Dockerfile", Length: 3200, Attributes: "Archive"},
		{Name: ".gitignore", Length: 650, Attributes: "Archive"},
		{Name: ".env.example", Length: 320, Attributes: "Archive"},
		{Name: "package-lock.json", Length: 8900, Attributes: "Archive"},
		{Name: "tsconfig.json", Length: 1200, Attributes: "Archive"},
		{Name: "jest.config.js", Length: 780, Attributes: "Archive"},
	}
	compressed := formatGCI(entries, false)

	inTokens := countTokens(psVerboseOutput)
	outTokens := countTokens(compressed)
	savings := (1 - float64(outTokens)/float64(inTokens)) * 100

	// 50% threshold for 20-entry list; real savings grow with more entries
	// (summary line overhead is amortized over more rows in production)
	if savings < 50 {
		t.Errorf("expected ≥50%% token savings, got %.1f%%\nInput tokens: %d\nOutput tokens: %d\nOutput:\n%s",
			savings, inTokens, outTokens, compressed)
	}
}

func TestHumanSize(t *testing.T) {
	tests := []struct{ n int64; want string }{
		{0, "0B"},
		{500, "500B"},
		{1024, "1.0K"},
		{1234, "1.2K"},
		{1048576, "1.0M"},
		{2500000, "2.4M"},
		{1073741824, "1.0G"},
	}
	for _, tt := range tests {
		got := humanSize(tt.n)
		if got != tt.want {
			t.Errorf("humanSize(%d) = %q, want %q", tt.n, got, tt.want)
		}
	}
}

func countTokens(s string) int {
	return len(strings.Fields(s))
}

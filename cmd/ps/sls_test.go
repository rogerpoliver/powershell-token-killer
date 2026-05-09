package ps

import (
	"strings"
	"testing"
)

func TestFilterSLS_single_file(t *testing.T) {
	raw := `src/main.go:12:func main() {
src/main.go:45:func handleMain(w http.ResponseWriter, r *http.Request) {
src/main.go:78:// main entry point
`
	out := FilterSLS(raw)
	// Should drop "src/main.go:" prefix
	if strings.Contains(out, "src/main.go:") {
		t.Errorf("single-file: filename prefix should be stripped, got:\n%s", out)
	}
	if !strings.Contains(out, "12:func main()") {
		t.Errorf("line number and content should remain, got:\n%s", out)
	}
}

func TestFilterSLS_multi_file(t *testing.T) {
	raw := `src/main.go:12:func main() {
src/server.go:8:func startServer() {
src/handler.go:34:func handleRequest() {
`
	out := FilterSLS(raw)
	// Multiple files — keep filename prefix
	if !strings.Contains(out, "src/main.go:") {
		t.Errorf("multi-file: filename should be kept, got:\n%s", out)
	}
	if !strings.Contains(out, "src/server.go:") {
		t.Errorf("multi-file: all filenames should appear, got:\n%s", out)
	}
}

func TestFilterSLS_truncates_long_lines(t *testing.T) {
	long := strings.Repeat("x", 130)
	raw := "file.go:1:" + long + "\n"
	out := FilterSLS(raw)
	// Should be truncated at 120 chars + "…"
	for _, line := range strings.Split(out, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if len(line) > 125 { // some slack for single-file prefix strip
			t.Errorf("line too long (%d chars), expected ≤125:\n%s", len(line), line)
		}
		if !strings.HasSuffix(line, "…") {
			t.Errorf("truncated line should end with …, got: %q", line)
		}
	}
}

func TestFilterSLS_summary_when_many(t *testing.T) {
	var lines []string
	for i := 1; i <= 8; i++ {
		lines = append(lines, "file.go:"+strings.Repeat("1", 1)+":match line")
	}
	raw := strings.Join(lines, "\n") + "\n"
	out := FilterSLS(raw)
	if !strings.Contains(out, "Found 8 matches") {
		t.Errorf("expected summary for >5 matches, got:\n%s", out)
	}
}

func TestFilterSLS_no_summary_few_matches(t *testing.T) {
	raw := `file.go:1:match one
file.go:2:match two
file.go:3:match three
`
	out := FilterSLS(raw)
	if strings.Contains(out, "Found") {
		t.Errorf("should not show summary for ≤5 matches, got:\n%s", out)
	}
}

func TestFilterSLS_empty(t *testing.T) {
	out := FilterSLS("")
	if out != "(no matches)\n" {
		t.Errorf("expected (no matches)\\n, got %q", out)
	}
}

func TestFilterSLS_token_savings(t *testing.T) {
	// Savings come from truncating long lines (>120 chars).
	// file:line:content format: prefix stripping doesn't reduce token count
	// because "file:line:" is a single whitespace-delimited token.
	// Build lines where content is 150+ chars to exercise truncation.
	longContent := strings.Repeat("word ", 35) // 35 words = 175 chars of content
	var lines []string
	for i := 1; i <= 10; i++ {
		lines = append(lines, "src/main.go:"+strings.Repeat("1", 1)+":"+longContent)
	}
	raw := strings.Join(lines, "\n") + "\n"

	out := FilterSLS(raw)
	inTok := countTokens(raw)
	outTok := countTokens(out)
	savings := (1 - float64(outTok)/float64(inTok)) * 100

	// Truncation cuts from 35 words/line to ~22 words/line (~37% per line).
	// Whole-output savings are lower due to prefix stripping and summary overhead.
	if savings < 25 {
		t.Errorf("expected ≥25%% savings from long-line truncation, got %.1f%%\nInput: %d tokens\nOutput: %d tokens",
			savings, inTok, outTok)
	}
}

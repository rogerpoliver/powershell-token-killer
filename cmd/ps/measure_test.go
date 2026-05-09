package ps

import (
	"strings"
	"testing"
)

// rawMeasureLine is typical Measure-Object -Line output
var rawMeasureLine = `
Count    : 42
Average  :
Sum      :
Maximum  :
Minimum  :
StandardDeviation :
Property :
`

// rawMeasureAllProps is Measure-Object -Line -Word -Character output
var rawMeasureAllProps = `
Lines      : 156
Words      : 892
Characters : 6234
Property   :
`

func TestFilterMeasure_drops_nulls(t *testing.T) {
	requested := map[string]string{"Count": "Lines"}
	out := FilterMeasure(rawMeasureLine, requested)

	if strings.Contains(out, "Average") {
		t.Errorf("null Average should be dropped, got:\n%s", out)
	}
	if strings.Contains(out, "Sum") {
		t.Errorf("null Sum should be dropped, got:\n%s", out)
	}
	if !strings.Contains(out, "Lines: 42") {
		t.Errorf("expected 'Lines: 42' (friendly label), got:\n%s", out)
	}
}

func TestFilterMeasure_friendly_labels(t *testing.T) {
	requested := map[string]string{
		"Lines":      "Lines",
		"Words":      "Words",
		"Characters": "Chars",
	}
	out := FilterMeasure(rawMeasureAllProps, requested)

	if !strings.Contains(out, "Lines: 156") {
		t.Errorf("expected Lines: 156, got:\n%s", out)
	}
	if !strings.Contains(out, "Words: 892") {
		t.Errorf("expected Words: 892, got:\n%s", out)
	}
	if !strings.Contains(out, "Chars: 6234") {
		t.Errorf("expected Chars: 6234 (friendly label), got:\n%s", out)
	}
}

func TestFilterMeasure_empty_input(t *testing.T) {
	out := FilterMeasure("", map[string]string{})
	if out != "(no output)\n" {
		t.Errorf("expected (no output)\\n, got %q", out)
	}
}

func TestFilterMeasure_all_null(t *testing.T) {
	raw := `
Count    :
Average  :
Sum      :
`
	out := FilterMeasure(raw, map[string]string{})
	if out != "(no output)\n" {
		t.Errorf("all-null props should give (no output)\\n, got %q", out)
	}
}

func TestFilterMeasure_single_line_output(t *testing.T) {
	// Output should be compact — all props on one line separated by spaces
	requested := map[string]string{"Count": "Lines"}
	out := FilterMeasure(rawMeasureLine, requested)
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 1 {
		t.Errorf("expected single-line output, got %d lines:\n%s", len(lines), out)
	}
}

func TestFilterMeasure_token_savings(t *testing.T) {
	// Full Measure-Object output has many null properties
	rawFull := `
Count    : 1284
Average  :
Sum      :
Maximum  :
Minimum  :
StandardDeviation :
Property :
`
	requested := map[string]string{"Count": "Lines"}
	out := FilterMeasure(rawFull, requested)
	inTok := countTokens(rawFull)
	outTok := countTokens(out)
	savings := (1 - float64(outTok)/float64(inTok)) * 100

	if savings < 70 {
		t.Errorf("expected ≥70%% savings (null props dropped), got %.1f%%\nInput: %d tokens\nOutput: %d tokens\nOutput: %s",
			savings, inTok, outTok, out)
	}
}

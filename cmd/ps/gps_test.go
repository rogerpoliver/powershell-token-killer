package ps

import (
	"strings"
	"testing"
)

func TestFormatGPS_basic(t *testing.T) {
	entries := []gpsEntry{
		{Name: "chrome", Id: 1234, CPU: 45.2, WorkingSet: 512 * 1024 * 1024},
		{Name: "code", Id: 5678, CPU: 12.1, WorkingSet: 256 * 1024 * 1024},
		{Name: "idle", Id: 0, CPU: 0.0, WorkingSet: 0},
	}
	out := formatGPS(entries)

	// Sorted by CPU desc
	chromeIdx := strings.Index(out, "chrome")
	codeIdx := strings.Index(out, "code")
	if chromeIdx > codeIdx {
		t.Error("expected chrome (higher CPU) to appear before code")
	}

	if !strings.Contains(out, "512.0M") {
		t.Errorf("expected human size 512.0M for chrome, got:\n%s", out)
	}
	if !strings.Contains(out, "45.2s") {
		t.Errorf("expected CPU 45.2s, got:\n%s", out)
	}
	if !strings.Contains(out, "Total: 3 processes") {
		t.Errorf("expected total line, got:\n%s", out)
	}
}

func TestFormatGPS_long_name_truncated(t *testing.T) {
	entries := []gpsEntry{
		{Name: "verylongprocessnamethatexceedslimit", Id: 999, CPU: 1.0, WorkingSet: 1024},
	}
	out := formatGPS(entries)
	if strings.Contains(out, "verylongprocessnamethatexceedslimit") {
		t.Errorf("long name should be truncated, got:\n%s", out)
	}
	if !strings.Contains(out, "...") {
		t.Errorf("truncated name should end with ..., got:\n%s", out)
	}
}

func TestFormatGPS_empty(t *testing.T) {
	out := formatGPS([]gpsEntry{})
	if out != "\nTotal: 0 processes\n" {
		t.Errorf("empty processes, got: %q", out)
	}
}

func TestFormatGPS_negative_cpu(t *testing.T) {
	entries := []gpsEntry{
		{Name: "idle", Id: 0, CPU: -1.0, WorkingSet: 0},
	}
	out := formatGPS(entries)
	if !strings.Contains(out, "0.0s") {
		t.Errorf("negative CPU should render as 0.0s, got:\n%s", out)
	}
}

func TestParseGPSJSON_array(t *testing.T) {
	raw := `[{"Name":"chrome","Id":1234,"CPU":45.2,"WorkingSet":536870912},{"Name":"code","Id":5678,"CPU":12.1,"WorkingSet":268435456}]`
	entries, err := parseGPSJSON(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Name != "chrome" {
		t.Errorf("expected chrome, got %q", entries[0].Name)
	}
}

func TestParseGPSJSON_single_object(t *testing.T) {
	raw := `{"Name":"pwsh","Id":1111,"CPU":3.5,"WorkingSet":104857600}`
	entries, err := parseGPSJSON(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
}

func TestFormatGPS_token_savings(t *testing.T) {
	// Raw JSON from Get-Process | Select Name,Id,CPU,WorkingSet | ConvertTo-Json
	rawJSON := `[
  {
    "Name": "chrome",
    "Id": 12340,
    "CPU": 145.234567,
    "WorkingSet": 1073741824
  },
  {
    "Name": "Code",
    "Id": 5678,
    "CPU": 45.123456,
    "WorkingSet": 524288000
  },
  {
    "Name": "pwsh",
    "Id": 9012,
    "CPU": 12.345678,
    "WorkingSet": 104857600
  },
  {
    "Name": "node",
    "Id": 3456,
    "CPU": 8.9,
    "WorkingSet": 209715200
  },
  {
    "Name": "Finder",
    "Id": 789,
    "CPU": 2.1,
    "WorkingSet": 52428800
  },
  {
    "Name": "WindowServer",
    "Id": 111,
    "CPU": 1.5,
    "WorkingSet": 78643200
  },
  {
    "Name": "Slack",
    "Id": 222,
    "CPU": 0.9,
    "WorkingSet": 419430400
  },
  {
    "Name": "Activity Monitor",
    "Id": 333,
    "CPU": 0.4,
    "WorkingSet": 31457280
  },
  {
    "Name": "Spotlight",
    "Id": 444,
    "CPU": 0.1,
    "WorkingSet": 15728640
  },
  {
    "Name": "kernel_task",
    "Id": 0,
    "CPU": 0.0,
    "WorkingSet": 2147483648
  }
]`
	entries, err := parseGPSJSON(rawJSON)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	compressed := formatGPS(entries)

	inTok := countTokens(rawJSON)
	outTok := countTokens(compressed)
	savings := (1 - float64(outTok)/float64(inTok)) * 100

	if savings < 55 {
		t.Errorf("expected ≥55%% token savings, got %.1f%%\nInput tokens: %d\nOutput tokens: %d\nOutput:\n%s",
			savings, inTok, outTok, compressed)
	}
}

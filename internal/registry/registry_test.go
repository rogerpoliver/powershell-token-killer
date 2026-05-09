package registry

import "testing"

func TestRewrite(t *testing.T) {
	tests := []struct {
		input   string
		want    string
		wantOK  bool
	}{
		{"Get-ChildItem", "ptk gci", true},
		{"Get-ChildItem .", "ptk gci .", true},
		{"Get-ChildItem src/ -Recurse", "ptk gci src/ -Recurse", true},
		{"get-childitem", "ptk gci", true},       // case-insensitive
		{"GET-CHILDITEM src/", "ptk gci src/", true},
		{"Select-String", "ptk sls", true},
		{"Select-String pattern file.txt", "ptk sls pattern file.txt", true},
		{"Get-Process", "ptk gps", true},
		{"Get-Service", "ptk gsv", true},
		{"Measure-Object", "ptk measure", true},
		{"Get-History", "ptk history", true},
		// No PTK filter for Get-Content or Get-WinEvent — passthrough
		{"Get-Content", "", false},
		{"Get-Help", "ptk phelp", true},
		// No PTK equivalent
		{"git status", "", false},
		{"cargo build", "", false},
		{"npm install", "", false},
		{"htop", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, ok := Rewrite(tt.input)
			if ok != tt.wantOK {
				t.Errorf("Rewrite(%q) ok=%v, want %v", tt.input, ok, tt.wantOK)
			}
			if got != tt.want {
				t.Errorf("Rewrite(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

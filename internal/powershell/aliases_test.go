package powershell

import "testing"

func TestNormalize(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"ls", "Get-ChildItem"},
		{"ls src/", "Get-ChildItem src/"},
		{"dir", "Get-ChildItem"},
		{"gci .", "Get-ChildItem ."},
		{"GCI src/", "Get-ChildItem src/"},   // case-insensitive
		{"cat file.txt", "Get-Content file.txt"},
		{"gc file.txt", "Get-Content file.txt"},
		{"grep pattern", "Select-String pattern"},
		{"sls pattern file.txt", "Select-String pattern file.txt"},
		{"ps", "Get-Process"},
		{"gps", "Get-Process"},
		{"gsv", "Get-Service"},
		{"measure", "Measure-Object"},
		{"wc", "Measure-Object"},
		{"history", "Get-History"},
		{"man Get-ChildItem", "Get-Help Get-ChildItem"},
		{"kill 1234", "Stop-Process 1234"},
		// No alias — unchanged
		{"git status", "git status"},
		{"cargo build", "cargo build"},
		{"docker ps", "docker ps"},
		{"Get-ChildItem", "Get-ChildItem"}, // already canonical
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := Normalize(tt.input)
			if got != tt.want {
				t.Errorf("Normalize(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

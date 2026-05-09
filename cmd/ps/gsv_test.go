package ps

import (
	"strings"
	"testing"
)

var rawGSV = `Status   Name               DisplayName
------   ----               -----------
Running  AdobeARMservice    Adobe Acrobat Update Service
Running  Apple Mobile Device Service Apple Mobile Device Service
Stopped  AppMgmt            Application Management
Running  AudioEndpointBuilder Windows Audio Endpoint Builder
Stopped  AxInstSV           ActiveX Installer (AxInstSV)
Running  BFE                Base Filtering Engine
Running  BITS               Background Intelligent Transfer Service
Stopped  Browser            Computer Browser
Running  COMSysApp          COM+ System Application
Stopped  CryptSvc           Cryptographic Services
Running  DcomLaunch         DCOM Server Process Launcher
Stopped  DeviceAssociationService Device Association Service
Running  Dhcp               DHCP Client
Running  DiagTrack          Connected User Experiences and Telemetry
`

func TestFilterGSV_basic(t *testing.T) {
	out := filterGSV(rawGSV)
	if strings.Contains(out, "------") {
		t.Error("separator line should be stripped")
	}
	if strings.Contains(out, "Status") && strings.Contains(out, "DisplayName") {
		t.Error("header line should be stripped")
	}
	if !strings.Contains(out, "Running") {
		t.Errorf("Running services should appear, got:\n%s", out)
	}
	if !strings.Contains(out, "AdobeARMservice") {
		t.Errorf("service name should appear, got:\n%s", out)
	}
}

func TestFilterGSV_strips_header(t *testing.T) {
	out := filterGSV(rawGSV)
	for _, line := range strings.Split(out, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if strings.HasPrefix(strings.TrimSpace(line), "---") {
			t.Errorf("separator should not appear in output: %q", line)
		}
	}
}

func TestFilterGSV_truncates_display_name(t *testing.T) {
	raw := `Status   Name    DisplayName
------   ----    -----------
Running  svc1    This is a very long display name that exceeds the forty character limit significantly
`
	out := filterGSV(raw)
	if !strings.Contains(out, "...") {
		t.Errorf("long display name should be truncated with ..., got:\n%s", out)
	}
}

func TestFilterGSV_empty(t *testing.T) {
	out := filterGSV("")
	if out != "(no services)\n" {
		t.Errorf("expected (no services)\\n, got %q", out)
	}
}

func TestFilterGSV_token_savings(t *testing.T) {
	out := filterGSV(rawGSV)
	inTok := countTokens(rawGSV)
	outTok := countTokens(out)
	savings := (1 - float64(outTok)/float64(inTok)) * 100

	// GSV savings come from header/separator removal (~8% on typical output).
	// Value is format normalization, not high compression ratio.
	if savings < 5 {
		t.Errorf("expected ≥5%% token savings, got %.1f%%\nInput tokens: %d\nOutput tokens: %d\nOutput:\n%s",
			savings, inTok, outTok, out)
	}
}

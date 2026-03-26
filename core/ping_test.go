package core

import (
	"testing"
	"time"
)

func TestSplitAndTrim(t *testing.T) {
	tests := []struct {
		input    string
		sep      string
		expected []string
	}{
		{"a,b,c", ",", []string{"a", "b", "c"}},
		{" a , b , c ", ",", []string{"a", "b", "c"}},
		{"", ",", nil},
		{"  ,  ,  ", ",", nil},
		{"single", ",", []string{"single"}},
		{"a,,b", ",", []string{"a", "b"}},
	}

	for _, tt := range tests {
		result := splitAndTrim(tt.input, tt.sep)
		if len(result) != len(tt.expected) {
			t.Errorf("splitAndTrim(%q, %q) = %v, want %v", tt.input, tt.sep, result, tt.expected)
			continue
		}
		for i := range result {
			if result[i] != tt.expected[i] {
				t.Errorf("splitAndTrim(%q, %q)[%d] = %q, want %q", tt.input, tt.sep, i, result[i], tt.expected[i])
			}
		}
	}
}

func TestSplitTargets(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"google.com,baidu.com", []string{"google.com", "baidu.com"}},
		{"1.1.1.1,8.8.8.8", []string{"1.1.1.1", "8.8.8.8"}},
		{"google.com, 1.1.1.1", []string{"google.com", "1.1.1.1"}},
		{"", nil},
		{"invalid host name!", nil},
		{"valid.com,invalid host!", []string{"valid.com"}},
	}

	for _, tt := range tests {
		result := splitTargets(tt.input)
		if len(result) != len(tt.expected) {
			t.Errorf("splitTargets(%q) = %v, want %v", tt.input, result, tt.expected)
			continue
		}
		for i := range result {
			if result[i] != tt.expected[i] {
				t.Errorf("splitTargets(%q)[%d] = %q, want %q", tt.input, i, result[i], tt.expected[i])
			}
		}
	}
}

func TestIsValidHostname(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"google.com", true},
		{"sub.domain.com", true},
		{"localhost", true},
		{"", false},
		{"has space", false},
		{"a", true},
	}

	for _, tt := range tests {
		result := isValidHostname(tt.input)
		if result != tt.expected {
			t.Errorf("isValidHostname(%q) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestIsValidHostname_TooLong(t *testing.T) {
	long := ""
	for i := 0; i < 256; i++ {
		long += "a"
	}
	if isValidHostname(long) {
		t.Error("expected false for hostname > 255 chars")
	}
}

func TestPingResultCalculateStats(t *testing.T) {
	r := &PingResult{
		RTTs: []time.Duration{100 * time.Millisecond, 200 * time.Millisecond, 300 * time.Millisecond},
	}
	r.calculateStats()

	if r.MinRTT != 100*time.Millisecond {
		t.Errorf("MinRTT = %v, want 100ms", r.MinRTT)
	}
	if r.MaxRTT != 300*time.Millisecond {
		t.Errorf("MaxRTT = %v, want 300ms", r.MaxRTT)
	}
	if r.AvgRTT != 200*time.Millisecond {
		t.Errorf("AvgRTT = %v, want 200ms", r.AvgRTT)
	}
}

func TestPingResultCalculateStats_Empty(t *testing.T) {
	r := &PingResult{}
	r.calculateStats() // Should not panic
	if r.MinRTT != 0 || r.MaxRTT != 0 || r.AvgRTT != 0 {
		t.Error("expected zero values for empty RTTs")
	}
}

func TestPingResultCalculateStats_Single(t *testing.T) {
	r := &PingResult{
		RTTs: []time.Duration{42 * time.Millisecond},
	}
	r.calculateStats()
	if r.MinRTT != 42*time.Millisecond || r.MaxRTT != 42*time.Millisecond || r.AvgRTT != 42*time.Millisecond {
		t.Errorf("single RTT: min=%v, max=%v, avg=%v", r.MinRTT, r.MaxRTT, r.AvgRTT)
	}
}

func TestNewPingConfig_DefaultRegion(t *testing.T) {
	config, err := NewPingConfig([]string{"-targets=google.com"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config.Region != RegionAuto {
		t.Errorf("expected region %s, got %s", RegionAuto, config.Region)
	}
}

func TestNewPingConfig_CNRegion(t *testing.T) {
	config, err := NewPingConfig([]string{"-region=cn"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config.Region != RegionCN {
		t.Errorf("expected region %s, got %s", RegionCN, config.Region)
	}
	if len(config.Targets) == 0 {
		t.Error("expected default CN targets when no targets specified")
	}
}

func TestNewPingConfig_CustomTargets(t *testing.T) {
	config, err := NewPingConfig([]string{"-targets=1.1.1.1,8.8.8.8"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(config.Targets) != 2 {
		t.Errorf("expected 2 targets, got %d", len(config.Targets))
	}
}

func TestNewPingConfig_InvalidTargetsOnly(t *testing.T) {
	_, err := NewPingConfig([]string{"-targets=invalid host!"})
	if err == nil {
		t.Error("expected error for all-invalid targets")
	}
}


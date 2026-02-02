package logger

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"testing"
)

func TestInit(t *testing.T) {
	// Test non-verbose (default)
	Init(false)
	if GetLevel() != LevelWarn {
		t.Errorf("Init(false) should set level to LevelWarn, got %v", GetLevel())
	}

	// Test verbose
	Init(true)
	if GetLevel() != LevelDebug {
		t.Errorf("Init(true) should set level to LevelDebug, got %v", GetLevel())
	}

	// Reset
	Init(false)
}

func TestSetLevel(t *testing.T) {
	tests := []Level{LevelDebug, LevelInfo, LevelWarn, LevelError}

	for _, level := range tests {
		t.Run(level.String(), func(t *testing.T) {
			SetLevel(level)
			if GetLevel() != level {
				t.Errorf("SetLevel(%v) failed, got %v", level, GetLevel())
			}
		})
	}

	// Reset
	SetLevel(LevelWarn)
}

func TestLevel_String(t *testing.T) {
	tests := []struct {
		level    Level
		expected string
	}{
		{LevelDebug, "DEBUG"},
		{LevelInfo, "INFO"},
		{LevelWarn, "WARN"},
		{LevelError, "ERROR"},
		{Level(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if tt.level.String() != tt.expected {
				t.Errorf("Level(%d).String() = %v, want %v", tt.level, tt.level.String(), tt.expected)
			}
		})
	}
}

func TestLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	SetOutput(&buf)
	defer SetOutput(nil) // Reset to default

	tests := []struct {
		name       string
		level      Level
		logFunc    func(string, ...interface{})
		shouldShow bool
	}{
		{"debug at debug level", LevelDebug, Debug, true},
		{"info at debug level", LevelDebug, Info, true},
		{"warn at debug level", LevelDebug, Warn, true},
		{"error at debug level", LevelDebug, Error, true},
		{"debug at info level", LevelInfo, Debug, false},
		{"info at info level", LevelInfo, Info, true},
		{"debug at warn level", LevelWarn, Debug, false},
		{"info at warn level", LevelWarn, Info, false},
		{"warn at warn level", LevelWarn, Warn, true},
		{"error at warn level", LevelWarn, Error, true},
		{"debug at error level", LevelError, Debug, false},
		{"warn at error level", LevelError, Warn, false},
		{"error at error level", LevelError, Error, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			SetLevel(tt.level)

			tt.logFunc("test message")

			hasOutput := buf.Len() > 0
			if hasOutput != tt.shouldShow {
				t.Errorf("got output=%v, want output=%v", hasOutput, tt.shouldShow)
			}
		})
	}

	// Reset
	SetLevel(LevelWarn)
}

func TestLogFormatting(t *testing.T) {
	var buf bytes.Buffer
	SetOutput(&buf)
	SetLevel(LevelDebug)
	defer func() {
		SetOutput(nil)
		SetLevel(LevelWarn)
	}()

	Debug("test %s %d", "message", 42)
	output := buf.String()

	// Check format: [LEVEL] timestamp message
	if !strings.HasPrefix(output, "[DEBUG]") {
		t.Errorf("Missing [DEBUG] prefix: %s", output)
	}

	if !strings.Contains(output, "test message 42") {
		t.Errorf("Missing formatted message: %s", output)
	}

	if !strings.HasSuffix(strings.TrimSpace(output), "test message 42") {
		t.Errorf("Message not at end: %s", output)
	}
}

func TestLogFields(t *testing.T) {
	var buf bytes.Buffer
	SetOutput(&buf)
	SetLevel(LevelDebug)
	defer func() {
		SetOutput(nil)
		SetLevel(LevelWarn)
	}()

	DebugFields("test message", map[string]interface{}{
		"key1": "value1",
		"key2": 42,
	})
	output := buf.String()

	// Check that fields are present
	if !strings.Contains(output, "key1=value1") {
		t.Errorf("Missing key1 field: %s", output)
	}

	if !strings.Contains(output, "key2=42") {
		t.Errorf("Missing key2 field: %s", output)
	}

	if !strings.Contains(output, "test message") {
		t.Errorf("Missing message: %s", output)
	}
}

func TestLogFieldsSorted(t *testing.T) {
	var buf bytes.Buffer
	SetOutput(&buf)
	SetLevel(LevelDebug)
	defer func() {
		SetOutput(nil)
		SetLevel(LevelWarn)
	}()

	// Fields should be sorted alphabetically
	DebugFields("test", map[string]interface{}{
		"zebra": 1,
		"alpha": 2,
		"beta":  3,
	})
	output := buf.String()

	// Check order: alpha should come before beta, beta before zebra
	alphaIdx := strings.Index(output, "alpha=")
	betaIdx := strings.Index(output, "beta=")
	zebraIdx := strings.Index(output, "zebra=")

	if alphaIdx == -1 || betaIdx == -1 || zebraIdx == -1 {
		t.Fatalf("Missing fields in output: %s", output)
	}

	if !(alphaIdx < betaIdx && betaIdx < zebraIdx) {
		t.Errorf("Fields not sorted alphabetically: %s", output)
	}
}

func TestLogError(t *testing.T) {
	var buf bytes.Buffer
	SetOutput(&buf)
	SetLevel(LevelError)
	defer func() {
		SetOutput(nil)
		SetLevel(LevelWarn)
	}()

	// Test with nil error
	buf.Reset()
	LogError(nil, "should not log")
	if buf.Len() > 0 {
		t.Error("LogError with nil should not produce output")
	}

	// Test with actual error
	buf.Reset()
	testErr := fmt.Errorf("test error")
	LogError(testErr, "operation failed")
	output := buf.String()
	if !strings.Contains(output, "[ERROR]") {
		t.Errorf("LogError should produce ERROR level: %s", output)
	}
	if !strings.Contains(output, "operation failed") {
		t.Errorf("LogError should contain message: %s", output)
	}
	if !strings.Contains(output, "test error") {
		t.Errorf("LogError should contain error: %s", output)
	}
}

func TestConcurrentLogging(t *testing.T) {
	var buf bytes.Buffer
	SetOutput(&buf)
	SetLevel(LevelDebug)
	defer func() {
		SetOutput(nil)
		SetLevel(LevelWarn)
	}()

	// Run multiple goroutines logging concurrently
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			Debug("goroutine %d", n)
			Info("info from %d", n)
			DebugFields("fields", map[string]interface{}{"n": n})
		}(i)
	}
	wg.Wait()

	// Count log lines
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	expected := 300 // 100 goroutines * 3 log calls each

	if len(lines) != expected {
		t.Errorf("Expected %d log lines, got %d", expected, len(lines))
	}

	// Check for corrupted lines (each line should have a level prefix)
	for i, line := range lines {
		if !strings.HasPrefix(line, "[DEBUG]") && !strings.HasPrefix(line, "[INFO]") {
			t.Errorf("Line %d may be corrupted: %s", i, line)
		}
	}
}

func TestEmptyFields(t *testing.T) {
	var buf bytes.Buffer
	SetOutput(&buf)
	SetLevel(LevelDebug)
	defer func() {
		SetOutput(nil)
		SetLevel(LevelWarn)
	}()

	DebugFields("no fields", nil)
	output := buf.String()

	if !strings.Contains(output, "no fields") {
		t.Errorf("Message should be present: %s", output)
	}

	// Should not have trailing fields separator
	trimmed := strings.TrimSpace(output)
	if strings.HasSuffix(trimmed, " ") {
		t.Errorf("Should not have trailing space: %q", trimmed)
	}
}

func TestAllLogFunctions(t *testing.T) {
	var buf bytes.Buffer
	SetOutput(&buf)
	SetLevel(LevelDebug)
	defer func() {
		SetOutput(nil)
		SetLevel(LevelWarn)
	}()

	// Test all basic log functions
	Debug("debug")
	Info("info")
	Warn("warn")
	Error("error")

	output := buf.String()
	if !strings.Contains(output, "[DEBUG]") {
		t.Error("Missing DEBUG output")
	}
	if !strings.Contains(output, "[INFO]") {
		t.Error("Missing INFO output")
	}
	if !strings.Contains(output, "[WARN]") {
		t.Error("Missing WARN output")
	}
	if !strings.Contains(output, "[ERROR]") {
		t.Error("Missing ERROR output")
	}

	// Test all field log functions
	buf.Reset()
	InfoFields("info", map[string]interface{}{"test": 1})
	WarnFields("warn", map[string]interface{}{"test": 2})
	ErrorFields("error", map[string]interface{}{"test": 3})

	output = buf.String()
	if !strings.Contains(output, "[INFO]") || !strings.Contains(output, "test=1") {
		t.Error("InfoFields output incorrect")
	}
	if !strings.Contains(output, "[WARN]") || !strings.Contains(output, "test=2") {
		t.Error("WarnFields output incorrect")
	}
	if !strings.Contains(output, "[ERROR]") || !strings.Contains(output, "test=3") {
		t.Error("ErrorFields output incorrect")
	}
}

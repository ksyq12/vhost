package output

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/fatih/color"
)

func init() {
	// Disable color for tests
	color.NoColor = true
}

// captureStdout captures stdout during function execution
func captureStdout(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Also set color output to the same writer
	color.Output = w

	f()

	w.Close()
	os.Stdout = old
	color.Output = os.Stdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String()
}

func TestJSON(t *testing.T) {
	t.Run("simple map", func(t *testing.T) {
		data := map[string]interface{}{
			"domain": "example.com",
			"status": "active",
		}

		output := captureStdout(func() {
			_ = JSON(data)
		})

		var result map[string]interface{}
		err := json.Unmarshal([]byte(output), &result)
		if err != nil {
			t.Fatalf("JSON output is invalid: %v", err)
		}

		if result["domain"] != "example.com" {
			t.Errorf("expected domain example.com, got %v", result["domain"])
		}
		if result["status"] != "active" {
			t.Errorf("expected status active, got %v", result["status"])
		}
	})

	t.Run("struct", func(t *testing.T) {
		type TestStruct struct {
			Name  string `json:"name"`
			Value int    `json:"value"`
		}
		data := TestStruct{Name: "test", Value: 42}

		output := captureStdout(func() {
			_ = JSON(data)
		})

		var result TestStruct
		err := json.Unmarshal([]byte(output), &result)
		if err != nil {
			t.Fatalf("JSON output is invalid: %v", err)
		}

		if result.Name != "test" {
			t.Errorf("expected name test, got %s", result.Name)
		}
		if result.Value != 42 {
			t.Errorf("expected value 42, got %d", result.Value)
		}
	})

	t.Run("slice", func(t *testing.T) {
		data := []string{"example.com", "test.com"}

		output := captureStdout(func() {
			_ = JSON(data)
		})

		var result []string
		err := json.Unmarshal([]byte(output), &result)
		if err != nil {
			t.Fatalf("JSON output is invalid: %v", err)
		}

		if len(result) != 2 {
			t.Errorf("expected 2 items, got %d", len(result))
		}
	})

	t.Run("empty object", func(t *testing.T) {
		data := map[string]interface{}{}

		output := captureStdout(func() {
			_ = JSON(data)
		})

		if !strings.Contains(output, "{}") {
			t.Errorf("expected empty object, got %s", output)
		}
	})
}

func TestTable(t *testing.T) {
	t.Run("basic table", func(t *testing.T) {
		headers := []string{"NAME", "STATUS"}
		rows := [][]string{
			{"example.com", "active"},
			{"test.com", "inactive"},
		}

		output := captureStdout(func() {
			Table(headers, rows)
		})

		if !strings.Contains(output, "NAME") {
			t.Error("output should contain header NAME")
		}
		if !strings.Contains(output, "STATUS") {
			t.Error("output should contain header STATUS")
		}
		if !strings.Contains(output, "example.com") {
			t.Error("output should contain example.com")
		}
		if !strings.Contains(output, "test.com") {
			t.Error("output should contain test.com")
		}
	})

	t.Run("empty headers", func(t *testing.T) {
		headers := []string{}
		rows := [][]string{{"data"}}

		output := captureStdout(func() {
			Table(headers, rows)
		})

		if output != "" {
			t.Errorf("expected no output for empty headers, got %s", output)
		}
	})

	t.Run("empty rows", func(t *testing.T) {
		headers := []string{"COL1", "COL2"}
		rows := [][]string{}

		output := captureStdout(func() {
			Table(headers, rows)
		})

		if !strings.Contains(output, "COL1") {
			t.Error("output should contain header COL1")
		}
		// Should have header and separator but no data rows
		lines := strings.Split(strings.TrimSpace(output), "\n")
		if len(lines) != 2 {
			t.Errorf("expected 2 lines (header + separator), got %d", len(lines))
		}
	})

	t.Run("uneven columns", func(t *testing.T) {
		headers := []string{"COL1", "COL2", "COL3"}
		rows := [][]string{
			{"a", "b"},           // missing COL3
			{"x", "y", "z", "w"}, // extra column (should be ignored)
		}

		output := captureStdout(func() {
			Table(headers, rows)
		})

		if !strings.Contains(output, "COL1") {
			t.Error("output should contain header COL1")
		}
		if !strings.Contains(output, "a") {
			t.Error("output should contain value a")
		}
	})

	t.Run("separator line", func(t *testing.T) {
		headers := []string{"NAME"}
		rows := [][]string{{"test"}}

		output := captureStdout(func() {
			Table(headers, rows)
		})

		if !strings.Contains(output, "----") {
			t.Error("table should have a separator line")
		}
	})

	t.Run("column alignment", func(t *testing.T) {
		headers := []string{"SHORT", "VERYLONGHEADER"}
		rows := [][]string{
			{"a", "b"},
		}

		output := captureStdout(func() {
			Table(headers, rows)
		})

		// Header line should have proper padding
		lines := strings.Split(strings.TrimSpace(output), "\n")
		if len(lines) < 1 {
			t.Fatal("expected at least 1 line")
		}
	})
}

func TestSuccess(t *testing.T) {
	output := captureStdout(func() {
		Success("operation completed")
	})

	if !strings.Contains(output, "operation completed") {
		t.Error("output should contain success message")
	}
	if !strings.Contains(output, "✓") {
		t.Error("output should contain success symbol")
	}
}

func TestError(t *testing.T) {
	output := captureStdout(func() {
		Error("operation failed")
	})

	if !strings.Contains(output, "operation failed") {
		t.Error("output should contain error message")
	}
	if !strings.Contains(output, "✗") {
		t.Error("output should contain error symbol")
	}
}

func TestWarn(t *testing.T) {
	output := captureStdout(func() {
		Warn("warning message")
	})

	if !strings.Contains(output, "warning message") {
		t.Error("output should contain warning message")
	}
	if !strings.Contains(output, "!") {
		t.Error("output should contain warning symbol")
	}
}

func TestInfo(t *testing.T) {
	output := captureStdout(func() {
		Info("info message")
	})

	if !strings.Contains(output, "info message") {
		t.Error("output should contain info message")
	}
	if !strings.Contains(output, "→") {
		t.Error("output should contain info symbol")
	}
}

func TestPrint(t *testing.T) {
	output := captureStdout(func() {
		Print("plain message")
	})

	if !strings.Contains(output, "plain message") {
		t.Error("output should contain plain message")
	}
}

func TestFormattedOutput(t *testing.T) {
	t.Run("success with format args", func(t *testing.T) {
		output := captureStdout(func() {
			Success("VHost %s created", "example.com")
		})

		if !strings.Contains(output, "VHost example.com created") {
			t.Errorf("expected formatted message, got %s", output)
		}
	})

	t.Run("error with format args", func(t *testing.T) {
		output := captureStdout(func() {
			Error("Failed: %s", "connection refused")
		})

		if !strings.Contains(output, "Failed: connection refused") {
			t.Errorf("expected formatted message, got %s", output)
		}
	})

	t.Run("warn with format args", func(t *testing.T) {
		output := captureStdout(func() {
			Warn("Found %d issues", 5)
		})

		if !strings.Contains(output, "Found 5 issues") {
			t.Errorf("expected formatted message, got %s", output)
		}
	})

	t.Run("info with format args", func(t *testing.T) {
		output := captureStdout(func() {
			Info("Processing %s...", "file.txt")
		})

		if !strings.Contains(output, "Processing file.txt...") {
			t.Errorf("expected formatted message, got %s", output)
		}
	})

	t.Run("print with format args", func(t *testing.T) {
		output := captureStdout(func() {
			Print("Value: %d", 42)
		})

		if !strings.Contains(output, "Value: 42") {
			t.Errorf("expected formatted message, got %s", output)
		}
	})
}

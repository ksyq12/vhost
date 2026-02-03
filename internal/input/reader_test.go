package input

import (
	"io"
	"testing"
)

func TestStringReader_ReadString(t *testing.T) {
	t.Run("single input", func(t *testing.T) {
		reader := NewStringReader("yes\n")
		result, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("ReadString failed: %v", err)
		}
		if result != "yes\n" {
			t.Errorf("expected 'yes\\n', got '%s'", result)
		}
	})

	t.Run("multiple inputs", func(t *testing.T) {
		reader := NewStringReader("first\n", "second\n", "third\n")

		result1, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("ReadString for first failed: %v", err)
		}
		if result1 != "first\n" {
			t.Errorf("expected 'first\\n', got '%s'", result1)
		}

		result2, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("ReadString for second failed: %v", err)
		}
		if result2 != "second\n" {
			t.Errorf("expected 'second\\n', got '%s'", result2)
		}

		result3, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("ReadString for third failed: %v", err)
		}
		if result3 != "third\n" {
			t.Errorf("expected 'third\\n', got '%s'", result3)
		}
	})

	t.Run("EOF after all inputs consumed", func(t *testing.T) {
		reader := NewStringReader("yes\n")
		_, err := reader.ReadString('\n') // consume the input
		if err != nil {
			t.Fatalf("ReadString failed: %v", err)
		}

		result, err := reader.ReadString('\n')
		if err != io.EOF {
			t.Errorf("expected io.EOF, got %v", err)
		}
		if result != "" {
			t.Errorf("expected empty string, got '%s'", result)
		}
	})

	t.Run("EOF on empty reader", func(t *testing.T) {
		reader := NewStringReader()
		result, err := reader.ReadString('\n')
		if err != io.EOF {
			t.Errorf("expected io.EOF, got %v", err)
		}
		if result != "" {
			t.Errorf("expected empty string, got '%s'", result)
		}
	})
}

func TestNewStdinReader(t *testing.T) {
	reader := NewStdinReader()
	if reader == nil {
		t.Fatal("expected non-nil reader")
	}
	if reader.reader == nil {
		t.Error("expected non-nil bufio.Reader")
	}
}

package executor

import (
	"errors"
	"testing"
)

func TestSystemExecutor_Execute(t *testing.T) {
	exec := NewSystemExecutor()

	t.Run("echo command", func(t *testing.T) {
		output, err := exec.Execute("echo", "hello")
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}
		if string(output) != "hello\n" {
			t.Errorf("expected 'hello\\n', got '%s'", string(output))
		}
	})

	t.Run("nonexistent command", func(t *testing.T) {
		_, err := exec.Execute("nonexistent-command-xyz-12345")
		if err == nil {
			t.Error("expected error for nonexistent command")
		}
	})
}

func TestSystemExecutor_LookPath(t *testing.T) {
	exec := NewSystemExecutor()

	t.Run("find sh", func(t *testing.T) {
		path, err := exec.LookPath("sh")
		if err != nil {
			t.Fatalf("LookPath failed: %v", err)
		}
		if path == "" {
			t.Error("expected non-empty path")
		}
	})

	t.Run("nonexistent command", func(t *testing.T) {
		_, err := exec.LookPath("nonexistent-command-xyz-12345")
		if err == nil {
			t.Error("expected error for nonexistent command")
		}
	})
}

func TestMockExecutor_Execute(t *testing.T) {
	t.Run("default behavior", func(t *testing.T) {
		mock := &MockExecutor{}
		output, err := mock.Execute("test", "arg1", "arg2")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if string(output) != "" {
			t.Errorf("expected empty output, got '%s'", string(output))
		}
		// Verify call was recorded
		if len(mock.Calls) != 1 {
			t.Errorf("expected 1 call, got %d", len(mock.Calls))
		}
		if mock.Calls[0].Name != "test" {
			t.Errorf("expected command 'test', got '%s'", mock.Calls[0].Name)
		}
	})

	t.Run("custom function", func(t *testing.T) {
		mock := &MockExecutor{
			ExecuteFunc: func(name string, args ...string) ([]byte, error) {
				return []byte("mocked output"), nil
			},
		}
		output, err := mock.Execute("test")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if string(output) != "mocked output" {
			t.Errorf("expected 'mocked output', got '%s'", string(output))
		}
	})

	t.Run("error case", func(t *testing.T) {
		mock := &MockExecutor{
			ExecuteFunc: func(name string, args ...string) ([]byte, error) {
				return []byte("error output"), errors.New("mock error")
			},
		}
		output, err := mock.Execute("test")
		if err == nil {
			t.Error("expected error")
		}
		if string(output) != "error output" {
			t.Errorf("expected 'error output', got '%s'", string(output))
		}
	})
}

func TestMockExecutor_LookPath(t *testing.T) {
	t.Run("default behavior", func(t *testing.T) {
		mock := &MockExecutor{}
		path, err := mock.LookPath("certbot")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if path != "/usr/bin/certbot" {
			t.Errorf("expected '/usr/bin/certbot', got '%s'", path)
		}
	})

	t.Run("custom function", func(t *testing.T) {
		mock := &MockExecutor{
			LookPathFunc: func(file string) (string, error) {
				if file == "certbot" {
					return "/usr/local/bin/certbot", nil
				}
				return "", errors.New("not found")
			},
		}

		path, err := mock.LookPath("certbot")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if path != "/usr/local/bin/certbot" {
			t.Errorf("expected '/usr/local/bin/certbot', got '%s'", path)
		}

		_, err = mock.LookPath("unknown")
		if err == nil {
			t.Error("expected error for unknown command")
		}
	})
}

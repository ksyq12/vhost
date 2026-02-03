package driver

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/ksyq12/vhost/internal/config"
	"github.com/ksyq12/vhost/internal/executor"
)

func TestCaddyDriver(t *testing.T) {
	// Create temp directories for testing
	tempDir := t.TempDir()
	availableDir := filepath.Join(tempDir, "sites-available")
	enabledDir := filepath.Join(tempDir, "sites-enabled")

	if err := os.MkdirAll(availableDir, 0755); err != nil {
		t.Fatalf("failed to create sites-available: %v", err)
	}
	if err := os.MkdirAll(enabledDir, 0755); err != nil {
		t.Fatalf("failed to create sites-enabled: %v", err)
	}

	// Create driver with test paths
	drv := NewCaddyWithPaths(availableDir, enabledDir)

	t.Run("Name", func(t *testing.T) {
		if drv.Name() != "caddy" {
			t.Errorf("expected caddy, got %s", drv.Name())
		}
	})

	t.Run("Paths", func(t *testing.T) {
		paths := drv.Paths()
		if paths.Available != availableDir {
			t.Errorf("expected %s, got %s", availableDir, paths.Available)
		}
		if paths.Enabled != enabledDir {
			t.Errorf("expected %s, got %s", enabledDir, paths.Enabled)
		}
	})

	t.Run("Add", func(t *testing.T) {
		vhost := &config.VHost{
			Domain: "test.example.com",
			Type:   "static",
			Root:   filepath.Join(tempDir, "www", "test.example.com"),
		}

		configContent := "test.example.com {\n    root * /var/www/html\n    file_server\n}"

		if err := drv.Add(vhost, configContent); err != nil {
			t.Fatalf("Add failed: %v", err)
		}

		// Check config file exists (no extension for Caddy)
		configPath := filepath.Join(availableDir, vhost.Domain)
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Error("config file was not created")
		}

		// Check content
		content, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("failed to read config: %v", err)
		}
		if string(content) != configContent {
			t.Errorf("config content mismatch")
		}

		// Check document root was created
		if _, err := os.Stat(vhost.Root); os.IsNotExist(err) {
			t.Error("document root was not created")
		}
	})

	t.Run("List", func(t *testing.T) {
		domains, err := drv.List()
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}

		if len(domains) != 1 {
			t.Errorf("expected 1 domain, got %d", len(domains))
		}

		if domains[0] != "test.example.com" {
			t.Errorf("expected test.example.com, got %s", domains[0])
		}
	})

	t.Run("Enable", func(t *testing.T) {
		domain := "test.example.com"

		if err := drv.Enable(domain); err != nil {
			t.Fatalf("Enable failed: %v", err)
		}

		// Check symlink exists (no extension for Caddy)
		symlinkPath := filepath.Join(enabledDir, domain)
		info, err := os.Lstat(symlinkPath)
		if err != nil {
			t.Fatalf("symlink not found: %v", err)
		}

		if info.Mode()&os.ModeSymlink == 0 {
			t.Error("expected symlink, got regular file")
		}
	})

	t.Run("IsEnabled", func(t *testing.T) {
		enabled, err := drv.IsEnabled("test.example.com")
		if err != nil {
			t.Fatalf("IsEnabled failed: %v", err)
		}
		if !enabled {
			t.Error("expected enabled to be true")
		}

		enabled, err = drv.IsEnabled("nonexistent.example.com")
		if err != nil {
			t.Fatalf("IsEnabled failed: %v", err)
		}
		if enabled {
			t.Error("expected enabled to be false for nonexistent domain")
		}
	})

	t.Run("Disable", func(t *testing.T) {
		domain := "test.example.com"

		if err := drv.Disable(domain); err != nil {
			t.Fatalf("Disable failed: %v", err)
		}

		// Check symlink was removed
		symlinkPath := filepath.Join(enabledDir, domain)
		if _, err := os.Lstat(symlinkPath); !os.IsNotExist(err) {
			t.Error("symlink should have been removed")
		}
	})

	t.Run("Remove", func(t *testing.T) {
		domain := "test.example.com"

		if err := drv.Remove(domain); err != nil {
			t.Fatalf("Remove failed: %v", err)
		}

		// Check config file was removed
		configPath := filepath.Join(availableDir, domain)
		if _, err := os.Stat(configPath); !os.IsNotExist(err) {
			t.Error("config file should have been removed")
		}
	})

	t.Run("RemoveNonexistent", func(t *testing.T) {
		err := drv.Remove("nonexistent.example.com")
		if err == nil {
			t.Error("expected error for nonexistent domain")
		}
	})
}

func TestCaddyDriverListFiltersCorrectly(t *testing.T) {
	tempDir := t.TempDir()
	availableDir := filepath.Join(tempDir, "sites-available")
	enabledDir := filepath.Join(tempDir, "sites-enabled")

	if err := os.MkdirAll(availableDir, 0755); err != nil {
		t.Fatalf("failed to create available dir: %v", err)
	}
	if err := os.MkdirAll(enabledDir, 0755); err != nil {
		t.Fatalf("failed to create enabled dir: %v", err)
	}

	drv := NewCaddyWithPaths(availableDir, enabledDir)

	// Create various files
	if err := os.WriteFile(filepath.Join(availableDir, "example.com"), []byte("config"), 0644); err != nil {
		t.Fatalf("failed to write example.com: %v", err)
	}
	if err := os.WriteFile(filepath.Join(availableDir, "test.org"), []byte("config"), 0644); err != nil {
		t.Fatalf("failed to write test.org: %v", err)
	}
	if err := os.WriteFile(filepath.Join(availableDir, ".hidden"), []byte("config"), 0644); err != nil { // hidden file
		t.Fatalf("failed to write .hidden: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(availableDir, "directory"), 0755); err != nil { // directory
		t.Fatalf("failed to create directory: %v", err)
	}

	domains, err := drv.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	// Should only include example.com and test.org (not hidden or directories)
	if len(domains) != 2 {
		t.Errorf("expected 2 domains, got %d: %v", len(domains), domains)
	}
}

func TestCaddyDriver_WithExecutor(t *testing.T) {
	tempDir := t.TempDir()
	availableDir := filepath.Join(tempDir, "sites-available")
	enabledDir := filepath.Join(tempDir, "sites-enabled")

	if err := os.MkdirAll(availableDir, 0755); err != nil {
		t.Fatalf("failed to create available dir: %v", err)
	}
	if err := os.MkdirAll(enabledDir, 0755); err != nil {
		t.Fatalf("failed to create enabled dir: %v", err)
	}

	t.Run("Test_success", func(t *testing.T) {
		mock := &executor.MockExecutor{
			ExecuteFunc: func(name string, args ...string) ([]byte, error) {
				if name == "caddy" && len(args) > 0 && args[0] == "validate" {
					return []byte("Valid configuration"), nil
				}
				return nil, errors.New("unexpected command")
			},
		}

		drv := NewCaddyWithExecutor(availableDir, enabledDir, mock)
		err := drv.Test()
		if err != nil {
			t.Errorf("Test should succeed: %v", err)
		}

		// Verify the correct command was called
		if len(mock.Calls) != 1 {
			t.Errorf("expected 1 call, got %d", len(mock.Calls))
		}
		if mock.Calls[0].Name != "caddy" || mock.Calls[0].Args[0] != "validate" {
			t.Errorf("expected caddy validate, got %s %v", mock.Calls[0].Name, mock.Calls[0].Args)
		}
	})

	t.Run("Test_failure", func(t *testing.T) {
		mock := &executor.MockExecutor{
			ExecuteFunc: func(name string, args ...string) ([]byte, error) {
				return []byte("Error: invalid config"), errors.New("exit status 1")
			},
		}

		drv := NewCaddyWithExecutor(availableDir, enabledDir, mock)
		err := drv.Test()
		if err == nil {
			t.Error("Test should fail for invalid config")
		}
	})

	t.Run("Reload_systemctl_success", func(t *testing.T) {
		mock := &executor.MockExecutor{
			ExecuteFunc: func(name string, args ...string) ([]byte, error) {
				if name == "systemctl" && len(args) >= 2 && args[0] == "reload" && args[1] == "caddy" {
					return []byte(""), nil
				}
				return nil, errors.New("unexpected command")
			},
		}

		drv := NewCaddyWithExecutor(availableDir, enabledDir, mock)
		err := drv.Reload()
		if err != nil {
			t.Errorf("Reload should succeed: %v", err)
		}
	})

	t.Run("Reload_fallback_success", func(t *testing.T) {
		callCount := 0
		mock := &executor.MockExecutor{
			ExecuteFunc: func(name string, args ...string) ([]byte, error) {
				callCount++
				if callCount == 1 {
					// First call: systemctl fails
					return []byte("systemctl not available"), errors.New("systemctl not found")
				}
				// Second call: caddy reload succeeds
				if name == "caddy" && len(args) > 0 && args[0] == "reload" {
					return []byte(""), nil
				}
				return nil, errors.New("unexpected command")
			},
		}

		drv := NewCaddyWithExecutor(availableDir, enabledDir, mock)
		err := drv.Reload()
		if err != nil {
			t.Errorf("Reload should succeed with fallback: %v", err)
		}

		if callCount != 2 {
			t.Errorf("expected 2 calls, got %d", callCount)
		}
	})

	t.Run("Reload_both_fail", func(t *testing.T) {
		mock := &executor.MockExecutor{
			ExecuteFunc: func(name string, args ...string) ([]byte, error) {
				return []byte("error"), errors.New("command failed")
			},
		}

		drv := NewCaddyWithExecutor(availableDir, enabledDir, mock)
		err := drv.Reload()
		if err == nil {
			t.Error("Reload should fail when both methods fail")
		}
	})
}

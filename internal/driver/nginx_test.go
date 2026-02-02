package driver

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/ksyq12/vhost/internal/config"
	"github.com/ksyq12/vhost/internal/executor"
)

func TestNginxDriver(t *testing.T) {
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
	drv := NewNginxWithPaths(availableDir, enabledDir)

	t.Run("Name", func(t *testing.T) {
		if drv.Name() != "nginx" {
			t.Errorf("expected nginx, got %s", drv.Name())
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

		configContent := "server { listen 80; server_name test.example.com; }"

		if err := drv.Add(vhost, configContent); err != nil {
			t.Fatalf("Add failed: %v", err)
		}

		// Check config file exists
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

		// Check symlink exists
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

func TestNginxDriver_WithExecutor(t *testing.T) {
	tempDir := t.TempDir()
	availableDir := filepath.Join(tempDir, "sites-available")
	enabledDir := filepath.Join(tempDir, "sites-enabled")

	os.MkdirAll(availableDir, 0755)
	os.MkdirAll(enabledDir, 0755)

	t.Run("Test_success", func(t *testing.T) {
		mock := &executor.MockExecutor{
			ExecuteFunc: func(name string, args ...string) ([]byte, error) {
				if name == "nginx" && len(args) > 0 && args[0] == "-t" {
					return []byte("nginx: configuration file test is successful"), nil
				}
				return nil, errors.New("unexpected command")
			},
		}

		drv := NewNginxWithExecutor(availableDir, enabledDir, mock)
		err := drv.Test()
		if err != nil {
			t.Errorf("Test should succeed: %v", err)
		}

		// Verify the correct command was called
		if len(mock.Calls) != 1 {
			t.Errorf("expected 1 call, got %d", len(mock.Calls))
		}
		if mock.Calls[0].Name != "nginx" || mock.Calls[0].Args[0] != "-t" {
			t.Errorf("expected nginx -t, got %s %v", mock.Calls[0].Name, mock.Calls[0].Args)
		}
	})

	t.Run("Test_failure", func(t *testing.T) {
		mock := &executor.MockExecutor{
			ExecuteFunc: func(name string, args ...string) ([]byte, error) {
				return []byte("nginx: [emerg] invalid config"), errors.New("exit status 1")
			},
		}

		drv := NewNginxWithExecutor(availableDir, enabledDir, mock)
		err := drv.Test()
		if err == nil {
			t.Error("Test should fail for invalid config")
		}
	})

	t.Run("Reload_systemctl_success", func(t *testing.T) {
		mock := &executor.MockExecutor{
			ExecuteFunc: func(name string, args ...string) ([]byte, error) {
				if name == "systemctl" && len(args) >= 2 && args[0] == "reload" && args[1] == "nginx" {
					return []byte(""), nil
				}
				return nil, errors.New("unexpected command")
			},
		}

		drv := NewNginxWithExecutor(availableDir, enabledDir, mock)
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
				// Second call: nginx -s reload succeeds
				if name == "nginx" && len(args) >= 2 && args[0] == "-s" && args[1] == "reload" {
					return []byte(""), nil
				}
				return nil, errors.New("unexpected command")
			},
		}

		drv := NewNginxWithExecutor(availableDir, enabledDir, mock)
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

		drv := NewNginxWithExecutor(availableDir, enabledDir, mock)
		err := drv.Reload()
		if err == nil {
			t.Error("Reload should fail when both methods fail")
		}
	})
}

func TestNginxDriver_EdgeCases(t *testing.T) {
	t.Run("EnableAlreadyEnabled", func(t *testing.T) {
		tempDir := t.TempDir()
		availableDir := filepath.Join(tempDir, "sites-available")
		enabledDir := filepath.Join(tempDir, "sites-enabled")

		os.MkdirAll(availableDir, 0755)
		os.MkdirAll(enabledDir, 0755)

		drv := NewNginxWithPaths(availableDir, enabledDir)
		domain := "test.com"

		// Create config file
		os.WriteFile(filepath.Join(availableDir, domain), []byte("config"), 0644)

		// Enable once
		if err := drv.Enable(domain); err != nil {
			t.Fatalf("first Enable failed: %v", err)
		}

		// Try to enable again
		err := drv.Enable(domain)
		if err == nil {
			t.Error("expected error when enabling already enabled domain")
		}
	})

	t.Run("DisableNotEnabled", func(t *testing.T) {
		tempDir := t.TempDir()
		availableDir := filepath.Join(tempDir, "sites-available")
		enabledDir := filepath.Join(tempDir, "sites-enabled")

		os.MkdirAll(availableDir, 0755)
		os.MkdirAll(enabledDir, 0755)

		drv := NewNginxWithPaths(availableDir, enabledDir)

		err := drv.Disable("nonexistent.com")
		if err == nil {
			t.Error("expected error when disabling non-enabled domain")
		}
	})

	t.Run("DisableNonSymlink", func(t *testing.T) {
		tempDir := t.TempDir()
		availableDir := filepath.Join(tempDir, "sites-available")
		enabledDir := filepath.Join(tempDir, "sites-enabled")

		os.MkdirAll(availableDir, 0755)
		os.MkdirAll(enabledDir, 0755)

		domain := "test.com"
		// Create regular file (not symlink) in enabled dir
		os.WriteFile(filepath.Join(enabledDir, domain), []byte("config"), 0644)

		drv := NewNginxWithPaths(availableDir, enabledDir)

		err := drv.Disable(domain)
		if err == nil {
			t.Error("expected error when trying to disable non-symlink")
		}
	})

	t.Run("EnableNonexistentSource", func(t *testing.T) {
		tempDir := t.TempDir()
		availableDir := filepath.Join(tempDir, "sites-available")
		enabledDir := filepath.Join(tempDir, "sites-enabled")

		os.MkdirAll(availableDir, 0755)
		os.MkdirAll(enabledDir, 0755)

		drv := NewNginxWithPaths(availableDir, enabledDir)

		err := drv.Enable("nonexistent.com")
		if err == nil {
			t.Error("expected error when source doesn't exist")
		}
	})

	t.Run("ListEmptyDirectory", func(t *testing.T) {
		tempDir := t.TempDir()
		availableDir := filepath.Join(tempDir, "sites-available")
		enabledDir := filepath.Join(tempDir, "sites-enabled")

		os.MkdirAll(availableDir, 0755)
		os.MkdirAll(enabledDir, 0755)

		drv := NewNginxWithPaths(availableDir, enabledDir)

		domains, err := drv.List()
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(domains) != 0 {
			t.Errorf("expected 0 domains, got %d", len(domains))
		}
	})

	t.Run("ListNonexistentDirectory", func(t *testing.T) {
		tempDir := t.TempDir()
		availableDir := filepath.Join(tempDir, "nonexistent")
		enabledDir := filepath.Join(tempDir, "sites-enabled")

		drv := NewNginxWithPaths(availableDir, enabledDir)

		domains, err := drv.List()
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(domains) != 0 {
			t.Errorf("expected 0 domains, got %d", len(domains))
		}
	})

	t.Run("ListSkipsHiddenFiles", func(t *testing.T) {
		tempDir := t.TempDir()
		availableDir := filepath.Join(tempDir, "sites-available")
		enabledDir := filepath.Join(tempDir, "sites-enabled")

		os.MkdirAll(availableDir, 0755)
		os.MkdirAll(enabledDir, 0755)

		// Create normal file and hidden file
		os.WriteFile(filepath.Join(availableDir, "test.com"), []byte("config"), 0644)
		os.WriteFile(filepath.Join(availableDir, ".hidden"), []byte("config"), 0644)

		drv := NewNginxWithPaths(availableDir, enabledDir)

		domains, err := drv.List()
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(domains) != 1 {
			t.Errorf("expected 1 domain, got %d", len(domains))
		}
		if domains[0] != "test.com" {
			t.Errorf("expected test.com, got %s", domains[0])
		}
	})

	t.Run("ListSkipsDirectories", func(t *testing.T) {
		tempDir := t.TempDir()
		availableDir := filepath.Join(tempDir, "sites-available")
		enabledDir := filepath.Join(tempDir, "sites-enabled")

		os.MkdirAll(availableDir, 0755)
		os.MkdirAll(enabledDir, 0755)

		// Create normal file and directory
		os.WriteFile(filepath.Join(availableDir, "test.com"), []byte("config"), 0644)
		os.MkdirAll(filepath.Join(availableDir, "subdir"), 0755)

		drv := NewNginxWithPaths(availableDir, enabledDir)

		domains, err := drv.List()
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(domains) != 1 {
			t.Errorf("expected 1 domain, got %d", len(domains))
		}
	})
}

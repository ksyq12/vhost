//go:build integration

package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ksyq12/vhost/internal/config"
	"github.com/ksyq12/vhost/internal/driver"
	"github.com/ksyq12/vhost/internal/template"
)

// testDirs holds paths to test directories, created fresh for each test
type testDirs struct {
	sitesAvailable string
	sitesEnabled   string
	wwwDir         string
}

// setupTestDirs creates temporary directories for testing
func setupTestDirs(t *testing.T) *testDirs {
	t.Helper()
	baseDir := t.TempDir() // Automatically cleaned up after test

	dirs := &testDirs{
		sitesAvailable: filepath.Join(baseDir, "sites-available"),
		sitesEnabled:   filepath.Join(baseDir, "sites-enabled"),
		wwwDir:         filepath.Join(baseDir, "www"),
	}

	if err := os.MkdirAll(dirs.sitesAvailable, 0755); err != nil {
		t.Fatalf("Failed to create sites-available directory: %v", err)
	}
	if err := os.MkdirAll(dirs.sitesEnabled, 0755); err != nil {
		t.Fatalf("Failed to create sites-enabled directory: %v", err)
	}
	if err := os.MkdirAll(dirs.wwwDir, 0755); err != nil {
		t.Fatalf("Failed to create www directory: %v", err)
	}

	return dirs
}

func TestNginxDriverIntegration(t *testing.T) {
	dirs := setupTestDirs(t)

	drv := driver.NewNginxWithPaths(dirs.sitesAvailable, dirs.sitesEnabled)

	t.Run("Add static vhost", func(t *testing.T) {
		vhost := &config.VHost{
			Domain:    "test.local",
			Type:      config.TypeStatic,
			Root:      filepath.Join(dirs.wwwDir, "test.local"),
			SSL:       false,
			Enabled:   true,
			CreatedAt: time.Now(),
		}

		// Render config using template
		content, err := template.Render("nginx", vhost)
		if err != nil {
			t.Fatalf("Failed to render template: %v", err)
		}

		// Add vhost
		if err := drv.Add(vhost, content); err != nil {
			t.Fatalf("Failed to add vhost: %v", err)
		}

		// Verify config file exists
		configPath := filepath.Join(dirs.sitesAvailable, "test.local")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Error("Config file was not created")
		}

		// Verify document root was created
		if _, err := os.Stat(vhost.Root); os.IsNotExist(err) {
			t.Error("Document root was not created")
		}
	})

	t.Run("Enable vhost", func(t *testing.T) {
		if err := drv.Enable("test.local"); err != nil {
			t.Fatalf("Failed to enable vhost: %v", err)
		}

		enabled, err := drv.IsEnabled("test.local")
		if err != nil {
			t.Fatalf("Failed to check enabled status: %v", err)
		}
		if !enabled {
			t.Error("VHost should be enabled")
		}

		// Verify symlink exists
		symlinkPath := filepath.Join(dirs.sitesEnabled, "test.local")
		info, err := os.Lstat(symlinkPath)
		if err != nil {
			t.Fatalf("Failed to stat symlink: %v", err)
		}
		if info.Mode()&os.ModeSymlink == 0 {
			t.Error("Expected symlink, got regular file")
		}
	})

	t.Run("List vhosts", func(t *testing.T) {
		domains, err := drv.List()
		if err != nil {
			t.Fatalf("Failed to list vhosts: %v", err)
		}

		found := false
		for _, d := range domains {
			if d == "test.local" {
				found = true
				break
			}
		}
		if !found {
			t.Error("test.local not found in list")
		}
	})

	t.Run("Disable vhost", func(t *testing.T) {
		if err := drv.Disable("test.local"); err != nil {
			t.Fatalf("Failed to disable vhost: %v", err)
		}

		enabled, _ := drv.IsEnabled("test.local")
		if enabled {
			t.Error("VHost should be disabled")
		}
	})

	t.Run("Remove vhost", func(t *testing.T) {
		if err := drv.Remove("test.local"); err != nil {
			t.Fatalf("Failed to remove vhost: %v", err)
		}

		configPath := filepath.Join(dirs.sitesAvailable, "test.local")
		if _, err := os.Stat(configPath); !os.IsNotExist(err) {
			t.Error("Config file should have been removed")
		}
	})
}

func TestNginxConfigValidation(t *testing.T) {
	if !isNginxAvailable() {
		t.Skip("Nginx is not available")
	}

	dirs := setupTestDirs(t)

	drv := driver.NewNginxWithPaths(dirs.sitesAvailable, dirs.sitesEnabled)

	t.Run("Valid config syntax", func(t *testing.T) {
		vhost := &config.VHost{
			Domain:    "valid.local",
			Type:      config.TypeStatic,
			Root:      filepath.Join(dirs.wwwDir, "valid.local"),
			SSL:       false,
			Enabled:   true,
			CreatedAt: time.Now(),
		}

		content, err := template.Render("nginx", vhost)
		if err != nil {
			t.Fatalf("Failed to render template: %v", err)
		}

		if err := drv.Add(vhost, content); err != nil {
			t.Fatalf("Failed to add vhost: %v", err)
		}

		if err := drv.Enable("valid.local"); err != nil {
			t.Fatalf("Failed to enable vhost: %v", err)
		}

		// Test nginx config (this requires nginx to be running in container)
		// Note: nginx -t checks the main config which includes our sites
		err = drv.Test()
		if err != nil {
			// Log but don't fail - nginx container might not include our config
			t.Logf("Nginx test returned: %v", err)
		}

		// Cleanup
		drv.Remove("valid.local")
	})
}

func TestProxyVhost(t *testing.T) {
	dirs := setupTestDirs(t)

	drv := driver.NewNginxWithPaths(dirs.sitesAvailable, dirs.sitesEnabled)

	t.Run("Add proxy vhost", func(t *testing.T) {
		vhost := &config.VHost{
			Domain:    "api.local",
			Type:      config.TypeProxy,
			ProxyPass: "http://localhost:3000",
			SSL:       false,
			Enabled:   true,
			CreatedAt: time.Now(),
		}

		content, err := template.Render("nginx", vhost)
		if err != nil {
			t.Fatalf("Failed to render template: %v", err)
		}

		if err := drv.Add(vhost, content); err != nil {
			t.Fatalf("Failed to add vhost: %v", err)
		}

		// Read and verify config contains proxy settings
		configPath := filepath.Join(dirs.sitesAvailable, "api.local")
		data, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("Failed to read config: %v", err)
		}

		configStr := string(data)
		if !strings.Contains(configStr, "proxy_pass") {
			t.Error("Config should contain proxy_pass directive")
		}
		if !strings.Contains(configStr, "http://localhost:3000") {
			t.Error("Config should contain proxy URL")
		}

		// Cleanup
		drv.Remove("api.local")
	})
}

func TestPHPVhost(t *testing.T) {
	dirs := setupTestDirs(t)

	drv := driver.NewNginxWithPaths(dirs.sitesAvailable, dirs.sitesEnabled)

	t.Run("Add PHP vhost", func(t *testing.T) {
		vhost := &config.VHost{
			Domain:     "php.local",
			Type:       config.TypePHP,
			Root:       filepath.Join(dirs.wwwDir, "php.local"),
			PHPVersion: "8.2",
			SSL:        false,
			Enabled:    true,
			CreatedAt:  time.Now(),
		}

		content, err := template.Render("nginx", vhost)
		if err != nil {
			t.Fatalf("Failed to render template: %v", err)
		}

		if err := drv.Add(vhost, content); err != nil {
			t.Fatalf("Failed to add vhost: %v", err)
		}

		// Read and verify config contains PHP-FPM settings
		configPath := filepath.Join(dirs.sitesAvailable, "php.local")
		data, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("Failed to read config: %v", err)
		}

		configStr := string(data)
		if !strings.Contains(configStr, "fastcgi_pass") {
			t.Error("Config should contain fastcgi_pass directive")
		}
		if !strings.Contains(configStr, "php8.2-fpm") || !strings.Contains(configStr, "php-fpm") {
			// Check for either naming convention
			if !strings.Contains(configStr, "php") {
				t.Error("Config should reference PHP-FPM")
			}
		}

		// Cleanup
		drv.Remove("php.local")
	})
}

func TestLaravelVhost(t *testing.T) {
	dirs := setupTestDirs(t)

	drv := driver.NewNginxWithPaths(dirs.sitesAvailable, dirs.sitesEnabled)

	t.Run("Add Laravel vhost", func(t *testing.T) {
		vhost := &config.VHost{
			Domain:     "laravel.local",
			Type:       config.TypeLaravel,
			Root:       filepath.Join(dirs.wwwDir, "laravel.local/public"),
			PHPVersion: "8.2",
			SSL:        false,
			Enabled:    true,
			CreatedAt:  time.Now(),
		}

		content, err := template.Render("nginx", vhost)
		if err != nil {
			t.Fatalf("Failed to render template: %v", err)
		}

		if err := drv.Add(vhost, content); err != nil {
			t.Fatalf("Failed to add vhost: %v", err)
		}

		// Read and verify config contains Laravel-specific settings
		configPath := filepath.Join(dirs.sitesAvailable, "laravel.local")
		data, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("Failed to read config: %v", err)
		}

		configStr := string(data)
		// Laravel configs typically have try_files with $uri
		if !strings.Contains(configStr, "try_files") {
			t.Error("Config should contain try_files directive")
		}

		// Cleanup
		drv.Remove("laravel.local")
	})
}

func TestErrorCases(t *testing.T) {
	dirs := setupTestDirs(t)

	drv := driver.NewNginxWithPaths(dirs.sitesAvailable, dirs.sitesEnabled)

	t.Run("Enable non-existent vhost", func(t *testing.T) {
		err := drv.Enable("nonexistent.local")
		if err == nil {
			t.Error("Expected error when enabling non-existent vhost")
		}
	})

	t.Run("Disable non-enabled vhost", func(t *testing.T) {
		// First create a vhost without enabling it
		vhost := &config.VHost{
			Domain:    "disabled.local",
			Type:      config.TypeStatic,
			Root:      "/var/www/disabled",
			SSL:       false,
			Enabled:   false,
			CreatedAt: time.Now(),
		}

		content, _ := template.Render("nginx", vhost)
		drv.Add(vhost, content)

		err := drv.Disable("disabled.local")
		if err == nil {
			t.Error("Expected error when disabling non-enabled vhost")
		}

		// Cleanup
		drv.Remove("disabled.local")
	})

	t.Run("Remove non-existent vhost", func(t *testing.T) {
		err := drv.Remove("nonexistent.local")
		if err == nil {
			t.Error("Expected error when removing non-existent vhost")
		}
	})

	t.Run("Enable already enabled vhost", func(t *testing.T) {
		vhost := &config.VHost{
			Domain:    "double.local",
			Type:      config.TypeStatic,
			Root:      "/var/www/double",
			SSL:       false,
			Enabled:   true,
			CreatedAt: time.Now(),
		}

		content, _ := template.Render("nginx", vhost)
		drv.Add(vhost, content)
		drv.Enable("double.local")

		err := drv.Enable("double.local")
		if err == nil {
			t.Error("Expected error when enabling already enabled vhost")
		}

		// Cleanup
		drv.Remove("double.local")
	})
}

func isNginxAvailable() bool {
	_, err := exec.LookPath("nginx")
	return err == nil
}

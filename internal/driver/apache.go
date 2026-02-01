package driver

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ksyq12/vhost/internal/config"
)

// ApacheDriver implements the Driver interface for Apache2
type ApacheDriver struct {
	paths Paths
}

// NewApache creates a new Apache driver with default paths
func NewApache() *ApacheDriver {
	return &ApacheDriver{
		paths: Paths{
			Available: "/etc/apache2/sites-available",
			Enabled:   "/etc/apache2/sites-enabled",
		},
	}
}

// NewApacheWithPaths creates a new Apache driver with custom paths
func NewApacheWithPaths(available, enabled string) *ApacheDriver {
	return &ApacheDriver{
		paths: Paths{
			Available: available,
			Enabled:   enabled,
		},
	}
}

// Name returns the driver name
func (a *ApacheDriver) Name() string {
	return "apache"
}

// Paths returns the config paths
func (a *ApacheDriver) Paths() Paths {
	return a.paths
}

// configFileName returns the config file name with .conf extension
func (a *ApacheDriver) configFileName(domain string) string {
	return domain + ".conf"
}

// Add creates a vhost config file
func (a *ApacheDriver) Add(vhost *config.VHost, configContent string) error {
	// Create sites-available directory if it doesn't exist
	if err := os.MkdirAll(a.paths.Available, 0755); err != nil {
		return fmt.Errorf("failed to create sites-available directory: %w", err)
	}

	// Create sites-enabled directory if it doesn't exist
	if err := os.MkdirAll(a.paths.Enabled, 0755); err != nil {
		return fmt.Errorf("failed to create sites-enabled directory: %w", err)
	}

	// Write config file to sites-available with .conf extension
	configPath := filepath.Join(a.paths.Available, a.configFileName(vhost.Domain))
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	// Create document root if specified and doesn't exist
	if vhost.Root != "" {
		if err := os.MkdirAll(vhost.Root, 0755); err != nil {
			return fmt.Errorf("failed to create document root: %w", err)
		}
	}

	return nil
}

// Remove deletes a vhost config
func (a *ApacheDriver) Remove(domain string) error {
	// First disable the site
	if enabled, _ := a.IsEnabled(domain); enabled {
		if err := a.Disable(domain); err != nil {
			return err
		}
	}

	// Remove config file from sites-available
	configPath := filepath.Join(a.paths.Available, a.configFileName(domain))
	if err := os.Remove(configPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("vhost %s not found", domain)
		}
		return fmt.Errorf("failed to remove config file: %w", err)
	}

	return nil
}

// Enable activates a vhost by creating a symlink
func (a *ApacheDriver) Enable(domain string) error {
	source := filepath.Join(a.paths.Available, a.configFileName(domain))
	target := filepath.Join(a.paths.Enabled, a.configFileName(domain))

	// Check if source exists
	if _, err := os.Stat(source); os.IsNotExist(err) {
		return fmt.Errorf("vhost %s not found in sites-available", domain)
	}

	// Check if already enabled
	if _, err := os.Lstat(target); err == nil {
		return fmt.Errorf("vhost %s is already enabled", domain)
	}

	// Create symlink
	if err := os.Symlink(source, target); err != nil {
		return fmt.Errorf("failed to enable vhost: %w", err)
	}

	return nil
}

// Disable deactivates a vhost by removing the symlink
func (a *ApacheDriver) Disable(domain string) error {
	target := filepath.Join(a.paths.Enabled, a.configFileName(domain))

	// Check if symlink exists
	info, err := os.Lstat(target)
	if os.IsNotExist(err) {
		return fmt.Errorf("vhost %s is not enabled", domain)
	}
	if err != nil {
		return fmt.Errorf("failed to check vhost status: %w", err)
	}

	// Verify it's a symlink
	if info.Mode()&os.ModeSymlink == 0 {
		return fmt.Errorf("vhost %s is not a symlink, refusing to remove", domain)
	}

	// Remove symlink
	if err := os.Remove(target); err != nil {
		return fmt.Errorf("failed to disable vhost: %w", err)
	}

	return nil
}

// List returns all vhost domains from sites-available
func (a *ApacheDriver) List() ([]string, error) {
	entries, err := os.ReadDir(a.paths.Available)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read sites-available: %w", err)
	}

	domains := make([]string, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		// Only include .conf files (not directories or hidden files)
		if !entry.IsDir() && !strings.HasPrefix(name, ".") && strings.HasSuffix(name, ".conf") {
			// Strip .conf extension to get domain name
			domain := strings.TrimSuffix(name, ".conf")
			domains = append(domains, domain)
		}
	}

	return domains, nil
}

// IsEnabled checks if a vhost is enabled
func (a *ApacheDriver) IsEnabled(domain string) (bool, error) {
	target := filepath.Join(a.paths.Enabled, a.configFileName(domain))
	_, err := os.Lstat(target)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check vhost status: %w", err)
	}
	return true, nil
}

// Test validates the apache config syntax
func (a *ApacheDriver) Test() error {
	cmd := exec.Command("apache2ctl", "configtest")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("apache config test failed: %s", string(output))
	}
	return nil
}

// Reload reloads apache to apply changes
func (a *ApacheDriver) Reload() error {
	cmd := exec.Command("systemctl", "reload", "apache2")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Try apache2ctl graceful as fallback
		cmd = exec.Command("apache2ctl", "graceful")
		output, err = cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to reload apache: %s", string(output))
		}
	}
	return nil
}

// init registers the apache driver
func init() {
	Register(NewApache())
}

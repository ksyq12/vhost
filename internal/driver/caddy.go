package driver

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ksyq12/vhost/internal/config"
)

// CaddyDriver implements the Driver interface for Caddy
type CaddyDriver struct {
	paths Paths
}

// NewCaddy creates a new Caddy driver with default paths
func NewCaddy() *CaddyDriver {
	return &CaddyDriver{
		paths: Paths{
			Available: "/etc/caddy/sites-available",
			Enabled:   "/etc/caddy/sites-enabled",
		},
	}
}

// NewCaddyWithPaths creates a new Caddy driver with custom paths
func NewCaddyWithPaths(available, enabled string) *CaddyDriver {
	return &CaddyDriver{
		paths: Paths{
			Available: available,
			Enabled:   enabled,
		},
	}
}

// Name returns the driver name
func (c *CaddyDriver) Name() string {
	return "caddy"
}

// Paths returns the config paths
func (c *CaddyDriver) Paths() Paths {
	return c.paths
}

// Add creates a vhost config file
func (c *CaddyDriver) Add(vhost *config.VHost, configContent string) error {
	// Create sites-available directory if it doesn't exist
	if err := os.MkdirAll(c.paths.Available, 0755); err != nil {
		return fmt.Errorf("failed to create sites-available directory: %w", err)
	}

	// Create sites-enabled directory if it doesn't exist
	if err := os.MkdirAll(c.paths.Enabled, 0755); err != nil {
		return fmt.Errorf("failed to create sites-enabled directory: %w", err)
	}

	// Write config file to sites-available
	configPath := filepath.Join(c.paths.Available, vhost.Domain)
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
func (c *CaddyDriver) Remove(domain string) error {
	// First disable the site
	if enabled, _ := c.IsEnabled(domain); enabled {
		if err := c.Disable(domain); err != nil {
			return err
		}
	}

	// Remove config file from sites-available
	configPath := filepath.Join(c.paths.Available, domain)
	if err := os.Remove(configPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("vhost %s not found", domain)
		}
		return fmt.Errorf("failed to remove config file: %w", err)
	}

	return nil
}

// Enable activates a vhost by creating a symlink
func (c *CaddyDriver) Enable(domain string) error {
	source := filepath.Join(c.paths.Available, domain)
	target := filepath.Join(c.paths.Enabled, domain)

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
func (c *CaddyDriver) Disable(domain string) error {
	target := filepath.Join(c.paths.Enabled, domain)

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
func (c *CaddyDriver) List() ([]string, error) {
	entries, err := os.ReadDir(c.paths.Available)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read sites-available: %w", err)
	}

	domains := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			domains = append(domains, entry.Name())
		}
	}

	return domains, nil
}

// IsEnabled checks if a vhost is enabled
func (c *CaddyDriver) IsEnabled(domain string) (bool, error) {
	target := filepath.Join(c.paths.Enabled, domain)
	_, err := os.Lstat(target)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check vhost status: %w", err)
	}
	return true, nil
}

// Test validates the caddy config syntax
func (c *CaddyDriver) Test() error {
	cmd := exec.Command("caddy", "validate", "--config", "/etc/caddy/Caddyfile")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("caddy config test failed: %s", string(output))
	}
	return nil
}

// Reload reloads caddy to apply changes
func (c *CaddyDriver) Reload() error {
	cmd := exec.Command("systemctl", "reload", "caddy")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Try caddy reload as fallback
		cmd = exec.Command("caddy", "reload", "--config", "/etc/caddy/Caddyfile")
		output, err = cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to reload caddy: %s", string(output))
		}
	}
	return nil
}

// init registers the caddy driver
func init() {
	Register(NewCaddy())
}

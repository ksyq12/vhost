package driver

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ksyq12/vhost/internal/config"
	"github.com/ksyq12/vhost/internal/executor"
)

// NginxDriver implements the Driver interface for Nginx
type NginxDriver struct {
	paths Paths
	exec  executor.CommandExecutor
}

// NewNginx creates a new Nginx driver with default paths
func NewNginx() *NginxDriver {
	return &NginxDriver{
		paths: Paths{
			Available: "/etc/nginx/sites-available",
			Enabled:   "/etc/nginx/sites-enabled",
		},
		exec: executor.NewSystemExecutor(),
	}
}

// NewNginxWithPaths creates a new Nginx driver with custom paths
func NewNginxWithPaths(available, enabled string) *NginxDriver {
	return &NginxDriver{
		paths: Paths{
			Available: available,
			Enabled:   enabled,
		},
		exec: executor.NewSystemExecutor(),
	}
}

// NewNginxWithExecutor creates a new Nginx driver with custom paths and executor (for testing)
func NewNginxWithExecutor(available, enabled string, exec executor.CommandExecutor) *NginxDriver {
	return &NginxDriver{
		paths: Paths{
			Available: available,
			Enabled:   enabled,
		},
		exec: exec,
	}
}

// Name returns the driver name
func (n *NginxDriver) Name() string {
	return "nginx"
}

// Paths returns the config paths
func (n *NginxDriver) Paths() Paths {
	return n.paths
}

// Add creates and enables a vhost config
func (n *NginxDriver) Add(vhost *config.VHost, configContent string) error {
	// Create sites-available directory if it doesn't exist
	if err := os.MkdirAll(n.paths.Available, 0755); err != nil {
		return fmt.Errorf("failed to create sites-available directory: %w", err)
	}

	// Create sites-enabled directory if it doesn't exist
	if err := os.MkdirAll(n.paths.Enabled, 0755); err != nil {
		return fmt.Errorf("failed to create sites-enabled directory: %w", err)
	}

	// Write config file to sites-available
	configPath := filepath.Join(n.paths.Available, vhost.Domain)
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
func (n *NginxDriver) Remove(domain string) error {
	// First disable the site
	if enabled, _ := n.IsEnabled(domain); enabled {
		if err := n.Disable(domain); err != nil {
			return err
		}
	}

	// Remove config file from sites-available
	configPath := filepath.Join(n.paths.Available, domain)
	if err := os.Remove(configPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("vhost %s not found", domain)
		}
		return fmt.Errorf("failed to remove config file: %w", err)
	}

	return nil
}

// Enable activates a vhost by creating a symlink
func (n *NginxDriver) Enable(domain string) error {
	source := filepath.Join(n.paths.Available, domain)
	target := filepath.Join(n.paths.Enabled, domain)

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
func (n *NginxDriver) Disable(domain string) error {
	target := filepath.Join(n.paths.Enabled, domain)

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
func (n *NginxDriver) List() ([]string, error) {
	entries, err := os.ReadDir(n.paths.Available)
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
func (n *NginxDriver) IsEnabled(domain string) (bool, error) {
	target := filepath.Join(n.paths.Enabled, domain)
	_, err := os.Lstat(target)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check vhost status: %w", err)
	}
	return true, nil
}

// Test validates the nginx config syntax
func (n *NginxDriver) Test() error {
	output, err := n.exec.Execute("nginx", "-t")
	if err != nil {
		return fmt.Errorf("nginx config test failed: %s", string(output))
	}
	return nil
}

// Reload reloads nginx to apply changes
func (n *NginxDriver) Reload() error {
	output, err := n.exec.Execute("systemctl", "reload", "nginx")
	if err != nil {
		// Try nginx -s reload as fallback
		output, err = n.exec.Execute("nginx", "-s", "reload")
		if err != nil {
			return fmt.Errorf("failed to reload nginx: %s", string(output))
		}
	}
	return nil
}

// init registers the nginx driver
func init() {
	Register(NewNginx())
}

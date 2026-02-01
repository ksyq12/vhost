package cli

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/ksyq12/vhost/internal/config"
	"github.com/ksyq12/vhost/internal/driver"
	"github.com/ksyq12/vhost/internal/output"
)

// loadConfigAndDriver loads config and returns the appropriate driver
func loadConfigAndDriver() (*config.Config, driver.Driver, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load config: %w", err)
	}

	drv, ok := driver.Get(cfg.Driver)
	if !ok {
		return nil, nil, fmt.Errorf("driver %s not found", cfg.Driver)
	}

	return cfg, drv, nil
}

// testAndReload tests config and reloads the web server
// If rollback is provided, it will be called on test failure
func testAndReload(drv driver.Driver, reload bool, rollback func() error) error {
	output.Info("Testing configuration...")
	if err := drv.Test(); err != nil {
		if rollback != nil {
			if rbErr := rollback(); rbErr != nil {
				output.Warn("Rollback failed: %v", rbErr)
			}
		}
		return fmt.Errorf("configuration test failed: %w", err)
	}

	if reload {
		output.Info("Reloading %s...", drv.Name())
		if err := drv.Reload(); err != nil {
			return fmt.Errorf("failed to reload %s: %w", drv.Name(), err)
		}
	}

	return nil
}

// saveConfig saves the config and returns error instead of just warning
func saveConfig(cfg *config.Config) error {
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	return nil
}

// outputResult handles JSON or human-readable output
func outputResult(data interface{}, successMsg string, args ...interface{}) error {
	if jsonOutput {
		return output.JSON(data)
	}
	output.Success(successMsg, args...)
	return nil
}

// validateDomain checks if domain is valid
func validateDomain(domain string) error {
	if domain == "" {
		return fmt.Errorf("domain cannot be empty")
	}
	if strings.Contains(domain, " ") {
		return fmt.Errorf("domain cannot contain spaces")
	}
	if strings.HasPrefix(domain, "-") || strings.HasSuffix(domain, "-") {
		return fmt.Errorf("domain cannot start or end with hyphen")
	}
	return nil
}

// validateRoot checks if root path is valid
func validateRoot(root string) error {
	if root == "" {
		return nil // empty is allowed (will be validated elsewhere if required)
	}
	if !filepath.IsAbs(root) {
		return fmt.Errorf("root path must be absolute: %s", root)
	}
	return nil
}

// validateProxyURL checks if proxy URL is valid
func validateProxyURL(proxyURL string) error {
	if proxyURL == "" {
		return nil
	}

	// Allow host:port format without scheme
	if !strings.Contains(proxyURL, "://") {
		proxyURL = "http://" + proxyURL
	}

	_, err := url.Parse(proxyURL)
	if err != nil {
		return fmt.Errorf("invalid proxy URL: %w", err)
	}

	return nil
}

// CommandResult represents a common result structure for CLI commands
type CommandResult struct {
	Success bool   `json:"success"`
	Domain  string `json:"domain"`
	Action  string `json:"action,omitempty"`
	Message string `json:"message,omitempty"`
}

// newSuccessResult creates a success result
func newSuccessResult(domain, action string) CommandResult {
	return CommandResult{
		Success: true,
		Domain:  domain,
		Action:  action,
	}
}

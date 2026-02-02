package cli

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

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

// getCertExpiry reads an SSL certificate and returns its expiry time
func getCertExpiry(certPath string) (time.Time, error) {
	data, err := os.ReadFile(certPath)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to read certificate: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return time.Time{}, fmt.Errorf("failed to parse certificate PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse certificate: %w", err)
	}

	return cert.NotAfter, nil
}

// getEditor returns the user's preferred editor
func getEditor() string {
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}
	return "vi"
}

// parseLogPaths extracts access_log and error_log paths from a config file
func parseLogPaths(drv driver.Driver, domain string) (accessLog, errorLog string, err error) {
	configPath := filepath.Join(drv.Paths().Available, domain)
	if drv.Name() == "apache" {
		configPath = filepath.Join(drv.Paths().Available, domain+".conf")
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to read config file: %w", err)
	}

	configStr := string(content)

	switch drv.Name() {
	case "nginx":
		accessLog = parseNginxLogPath(configStr, "access_log")
		errorLog = parseNginxLogPath(configStr, "error_log")
	case "apache":
		accessLog = parseApacheLogPath(configStr, "CustomLog")
		errorLog = parseApacheLogPath(configStr, "ErrorLog")
	case "caddy":
		// Caddy uses a different log format, try to find log path
		accessLog = parseCaddyLogPath(configStr)
		errorLog = accessLog // Caddy typically uses a single log file
	}

	// Fall back to default paths if not found in config
	if accessLog == "" {
		accessLog = getDefaultLogPath(drv.Name(), domain, "access")
	}
	if errorLog == "" {
		errorLog = getDefaultLogPath(drv.Name(), domain, "error")
	}

	return accessLog, errorLog, nil
}

// parseNginxLogPath extracts log path from nginx config
func parseNginxLogPath(config, directive string) string {
	pattern := regexp.MustCompile(directive + `\s+([^\s;]+)`)
	matches := pattern.FindStringSubmatch(config)
	if len(matches) >= 2 {
		return matches[1]
	}
	return ""
}

// parseApacheLogPath extracts log path from apache config
func parseApacheLogPath(config, directive string) string {
	pattern := regexp.MustCompile(directive + `\s+([^\s]+)`)
	matches := pattern.FindStringSubmatch(config)
	if len(matches) >= 2 {
		path := matches[1]
		// Handle Apache variable like ${APACHE_LOG_DIR}
		path = strings.ReplaceAll(path, "${APACHE_LOG_DIR}", "/var/log/apache2")
		return path
	}
	return ""
}

// parseCaddyLogPath extracts log path from caddy config
func parseCaddyLogPath(config string) string {
	pattern := regexp.MustCompile(`output\s+file\s+([^\s\}]+)`)
	matches := pattern.FindStringSubmatch(config)
	if len(matches) >= 2 {
		return matches[1]
	}
	return ""
}

// getDefaultLogPath returns default log path for a driver
func getDefaultLogPath(driverName, domain, logType string) string {
	switch driverName {
	case "nginx":
		return fmt.Sprintf("/var/log/nginx/%s-%s.log", domain, logType)
	case "apache":
		return fmt.Sprintf("/var/log/apache2/%s-%s.log", domain, logType)
	case "caddy":
		return fmt.Sprintf("/var/log/caddy/%s.log", domain)
	default:
		return ""
	}
}

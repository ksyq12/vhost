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
	"github.com/ksyq12/vhost/internal/platform"
)

// loadConfigAndDriver loads config and returns the appropriate driver
func loadConfigAndDriver() (*config.Config, driver.Driver, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Resolve paths: config override > platform detection
	paths, err := resolvePaths(cfg)
	if err != nil {
		return nil, nil, err
	}

	// Create driver with resolved paths
	drv, err := createDriverWithPaths(cfg.Driver, paths)
	if err != nil {
		return nil, nil, err
	}

	return cfg, drv, nil
}

// resolvePaths determines the paths to use for the driver.
// Priority: config override > platform auto-detection
func resolvePaths(cfg *config.Config) (driver.Paths, error) {
	// Priority 1: Use config paths if provided
	if cfg.Paths != nil && cfg.Paths.Available != "" && cfg.Paths.Enabled != "" {
		// Validate that paths are absolute
		if !filepath.IsAbs(cfg.Paths.Available) {
			return driver.Paths{}, fmt.Errorf("paths.available must be an absolute path: %s", cfg.Paths.Available)
		}
		if !filepath.IsAbs(cfg.Paths.Enabled) {
			return driver.Paths{}, fmt.Errorf("paths.enabled must be an absolute path: %s", cfg.Paths.Enabled)
		}

		return driver.Paths{
			Available: cfg.Paths.Available,
			Enabled:   cfg.Paths.Enabled,
		}, nil
	}

	// Validate partial config (only one path specified)
	if cfg.Paths != nil && (cfg.Paths.Available != "" || cfg.Paths.Enabled != "") {
		return driver.Paths{}, fmt.Errorf("both paths.available and paths.enabled must be set if either is specified")
	}

	// Priority 2: Auto-detect platform paths
	platformPaths, err := platform.DetectPaths()
	if err != nil {
		return driver.Paths{}, fmt.Errorf("failed to detect platform paths: %w\n\n"+
			"To manually configure paths, add to ~/.config/vhost/config.yaml:\n"+
			"paths:\n"+
			"  available: /path/to/sites-available\n"+
			"  enabled: /path/to/sites-enabled", err)
	}

	// Get paths for the configured driver
	pathConfig, err := platformPaths.GetPathsForDriver(cfg.Driver)
	if err != nil {
		return driver.Paths{}, err
	}

	return driver.Paths{
		Available: pathConfig.Available,
		Enabled:   pathConfig.Enabled,
	}, nil
}

// createDriverWithPaths creates a driver instance with the specified paths.
func createDriverWithPaths(driverName string, paths driver.Paths) (driver.Driver, error) {
	switch driverName {
	case "nginx":
		return driver.NewNginxWithPaths(paths.Available, paths.Enabled), nil
	case "apache":
		return driver.NewApacheWithPaths(paths.Available, paths.Enabled), nil
	case "caddy":
		return driver.NewCaddyWithPaths(paths.Available, paths.Enabled), nil
	default:
		return nil, fmt.Errorf("unknown driver: %s (available: nginx, apache, caddy)", driverName)
	}
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

// Maximum domain length according to RFC 1035
const maxDomainLength = 253

// domainPattern validates domain format (RFC 1035 compliant)
var domainPattern = regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)*[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?$`)

// validateDomain checks if domain is valid and secure.
// Validates against:
//   - Empty or whitespace-only input
//   - Maximum length (253 characters per RFC 1035)
//   - Path traversal sequences (../, ./)
//   - Shell metacharacters (;|&$`<>)
//   - Null byte injection
//   - RFC 1035 domain format compliance
func validateDomain(domain string) error {
	// Check for empty domain
	if domain == "" {
		return fmt.Errorf("domain cannot be empty")
	}

	// Check for whitespace
	trimmed := strings.TrimSpace(domain)
	if trimmed != domain {
		return fmt.Errorf("domain cannot contain leading or trailing whitespace")
	}
	if strings.Contains(domain, " ") {
		return fmt.Errorf("domain cannot contain spaces")
	}

	// Check maximum length (RFC 1035)
	if len(domain) > maxDomainLength {
		return fmt.Errorf("domain exceeds maximum length of %d characters", maxDomainLength)
	}

	// Security: Prevent path traversal attacks
	if containsPathTraversal(domain) {
		return fmt.Errorf("domain contains invalid path traversal sequences")
	}

	// Security: Prevent shell injection
	if containsShellMetaChars(domain) {
		return fmt.Errorf("domain contains invalid shell metacharacters")
	}

	// Security: Prevent null byte injection
	if containsNullByte(domain) {
		return fmt.Errorf("domain contains null byte")
	}

	// Validate hyphen rules (can't start or end with hyphen)
	if strings.HasPrefix(domain, "-") || strings.HasSuffix(domain, "-") {
		return fmt.Errorf("domain cannot start or end with hyphen")
	}

	// Validate domain format (RFC 1035 compliance)
	if !isValidDomainFormat(domain) {
		return fmt.Errorf("invalid domain format: must contain only letters, numbers, hyphens, and dots")
	}

	return nil
}

// validateRoot checks if root path is valid and secure.
// Validates against:
//   - Non-absolute paths
//   - Path traversal sequences (../, ./)
//   - Null byte injection
//   - Non-canonical paths (containing /./ or /../)
func validateRoot(root string) error {
	if root == "" {
		return nil // empty is allowed (will be validated elsewhere if required)
	}

	// Must be absolute path
	if !filepath.IsAbs(root) {
		return fmt.Errorf("root path must be absolute: %s", root)
	}

	// Security: Prevent path traversal attacks
	if containsPathTraversal(root) {
		return fmt.Errorf("root path contains invalid traversal sequences: %s", root)
	}

	// Security: Prevent null byte injection
	if containsNullByte(root) {
		return fmt.Errorf("root path contains null byte")
	}

	// Verify path is canonical (Clean doesn't change it)
	cleaned := filepath.Clean(root)
	if cleaned != root {
		return fmt.Errorf("root path contains invalid sequences: use %s instead of %s", cleaned, root)
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

// Security validation helper functions

// containsPathTraversal checks for path traversal patterns that could be used
// to access files outside the intended directory.
func containsPathTraversal(s string) bool {
	patterns := []string{
		"../",  // Unix path traversal
		"..\\", // Windows path traversal
		"./",   // Current directory reference
		".\\",  // Windows current directory
	}
	for _, pattern := range patterns {
		if strings.Contains(s, pattern) {
			return true
		}
	}
	return false
}

// containsNullByte checks for null byte injection attempts.
// Null bytes can be used to truncate strings in C-based libraries.
func containsNullByte(s string) bool {
	return strings.Contains(s, "\x00")
}

// containsShellMetaChars checks for shell metacharacters that could be used
// for command injection attacks.
func containsShellMetaChars(s string) bool {
	// Characters that have special meaning in shells
	metaChars := []rune{';', '|', '&', '$', '`', '<', '>', '\n', '\r'}
	for _, r := range s {
		for _, meta := range metaChars {
			if r == meta {
				return true
			}
		}
	}
	return false
}

// isValidDomainFormat validates domain name format according to RFC 1035.
// Allows alphanumeric characters, hyphens, and dots.
func isValidDomainFormat(domain string) bool {
	if len(domain) == 0 || len(domain) > maxDomainLength {
		return false
	}
	return domainPattern.MatchString(domain)
}

// requireRoot checks if the current process is running as root (UID 0).
// Returns an error if not running as root, enforcing security policy.
func requireRoot() error {
	if os.Geteuid() != 0 {
		return fmt.Errorf("this operation requires root privileges. Please run with sudo")
	}
	return nil
}

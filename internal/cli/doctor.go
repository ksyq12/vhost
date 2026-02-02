package cli

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/ksyq12/vhost/internal/config"
	"github.com/ksyq12/vhost/internal/driver"
	"github.com/ksyq12/vhost/internal/executor"
	"github.com/ksyq12/vhost/internal/output"
	"github.com/ksyq12/vhost/internal/ssl"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check system status and diagnose issues",
	Long: `Run diagnostic checks on the system and vhost configuration.

Checks:
  - Web server installation (nginx, apache, caddy)
  - PHP-FPM status
  - Certbot installation
  - Configuration file validity
  - Virtual host status

Examples:
  vhost doctor
  vhost doctor --json`,
	RunE: runDoctor,
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}

// CheckResult represents a single diagnostic check result
type CheckResult struct {
	Status  string `json:"status"` // "success", "warning", "error"
	Message string `json:"message"`
}

// VHostStatus represents the status of a single vhost
type VHostStatus struct {
	Domain  string        `json:"domain"`
	Enabled bool          `json:"enabled"`
	Checks  []CheckResult `json:"checks"`
}

// DoctorReport contains all diagnostic results
type DoctorReport struct {
	SystemRequirements []CheckResult  `json:"system_requirements"`
	Configuration      []CheckResult  `json:"configuration"`
	VHosts             []VHostStatus  `json:"vhosts"`
}

func runDoctor(cmd *cobra.Command, args []string) error {
	// Create executor for system commands
	exec := executor.NewSystemExecutor()

	// Load config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get driver
	drv, ok := driver.Get(cfg.Driver)
	if !ok {
		return fmt.Errorf("driver %s not found", cfg.Driver)
	}

	// Run all checks
	report := &DoctorReport{}
	report.SystemRequirements = checkSystemRequirements(exec, cfg)
	report.Configuration = checkConfiguration(drv, cfg)
	report.VHosts = checkVHosts(drv, cfg)

	// Output results
	if jsonOutput {
		return output.JSON(report)
	}

	displayDoctorResults(report)
	return nil
}

func checkSystemRequirements(exec executor.CommandExecutor, cfg *config.Config) []CheckResult {
	results := []CheckResult{}

	// Version extraction patterns
	versionPatterns := map[string]*regexp.Regexp{
		"nginx":   regexp.MustCompile(`nginx/(\d+\.\d+\.\d+)`),
		"apache2": regexp.MustCompile(`Apache/(\d+\.\d+\.\d+)`),
		"caddy":   regexp.MustCompile(`v?(\d+\.\d+\.\d+)`),
	}

	// Check web servers
	webServers := []struct {
		name     string
		binary   string
		versionFlag string
		optional bool
	}{
		{"Nginx", "nginx", "-v", cfg.Driver != "nginx"},
		{"Apache", "apache2", "-v", cfg.Driver != "apache"},
		{"Caddy", "caddy", "version", cfg.Driver != "caddy"},
	}

	for _, ws := range webServers {
		if _, err := exec.LookPath(ws.binary); err == nil {
			// Get version
			versionOutput, err := exec.Execute(ws.binary, ws.versionFlag)
			version := "unknown"
			if err == nil {
				if pattern, ok := versionPatterns[ws.binary]; ok {
					if matches := pattern.FindStringSubmatch(string(versionOutput)); len(matches) >= 2 {
						version = matches[1]
					}
				}
			}
			results = append(results, CheckResult{
				Status:  "success",
				Message: fmt.Sprintf("%s installed (%s)", ws.name, version),
			})
		} else {
			status := "error"
			suffix := ""
			if ws.optional {
				status = "warning"
				suffix = " (optional)"
			}
			results = append(results, CheckResult{
				Status:  status,
				Message: fmt.Sprintf("%s not installed%s", ws.name, suffix),
			})
		}
	}

	// Check PHP-FPM
	phpVersions := []string{"8.3", "8.2", "8.1", "8.0", "7.4"}
	phpFound := false
	for _, v := range phpVersions {
		if isPHPFPMRunning(exec, v) {
			results = append(results, CheckResult{
				Status:  "success",
				Message: fmt.Sprintf("PHP-FPM %s running", v),
			})
			phpFound = true
			break
		}
	}
	if !phpFound {
		// Check if any PHP type vhosts exist
		needsPHP := false
		for _, vhost := range cfg.VHosts {
			if vhost.Type == "php" || vhost.Type == "laravel" || vhost.Type == "wordpress" {
				needsPHP = true
				break
			}
		}
		status := "warning"
		if needsPHP {
			status = "error"
		}
		results = append(results, CheckResult{
			Status:  status,
			Message: "PHP-FPM not detected",
		})
	}

	// Check Certbot
	if ssl.IsInstalled() {
		results = append(results, CheckResult{
			Status:  "success",
			Message: "Certbot installed",
		})
	} else {
		// Check if any SSL vhosts exist
		needsSSL := false
		for _, vhost := range cfg.VHosts {
			if vhost.SSL {
				needsSSL = true
				break
			}
		}
		status := "warning"
		if needsSSL {
			status = "error"
		}
		results = append(results, CheckResult{
			Status:  status,
			Message: "Certbot not installed",
		})
	}

	return results
}

func isPHPFPMRunning(exec executor.CommandExecutor, version string) bool {
	serviceName := fmt.Sprintf("php%s-fpm", version)

	// Try systemctl
	if out, err := exec.Execute("systemctl", "is-active", serviceName); err == nil {
		if strings.TrimSpace(string(out)) == "active" {
			return true
		}
	}

	// Try service command
	if out, err := exec.Execute("service", serviceName, "status"); err == nil {
		if strings.Contains(string(out), "running") || strings.Contains(string(out), "active") {
			return true
		}
	}

	// Check socket file
	socketPath := fmt.Sprintf("/run/php/php%s-fpm.sock", version)
	if _, err := os.Stat(socketPath); err == nil {
		return true
	}

	return false
}

func checkConfiguration(drv driver.Driver, cfg *config.Config) []CheckResult {
	results := []CheckResult{}

	// Check config file exists
	configPath, pathErr := config.ConfigPath()
	if pathErr == nil {
		if _, err := os.Stat(configPath); err == nil {
			// Use ~ notation for display
			displayPath := strings.Replace(configPath, os.Getenv("HOME"), "~", 1)
			results = append(results, CheckResult{
				Status:  "success",
				Message: fmt.Sprintf("Config file exists (%s)", displayPath),
			})
		} else {
			results = append(results, CheckResult{
				Status:  "error",
				Message: "Config file not found",
			})
		}
	} else {
		results = append(results, CheckResult{
			Status:  "error",
			Message: "Could not determine config path",
		})
	}

	// Test web server config syntax
	if err := drv.Test(); err == nil {
		results = append(results, CheckResult{
			Status:  "success",
			Message: fmt.Sprintf("%s config syntax OK", capitalize(drv.Name())),
		})
	} else {
		results = append(results, CheckResult{
			Status:  "error",
			Message: fmt.Sprintf("%s config syntax error", capitalize(drv.Name())),
		})
	}

	return results
}

func checkVHosts(drv driver.Driver, cfg *config.Config) []VHostStatus {
	statuses := []VHostStatus{}

	for domain, vhost := range cfg.VHosts {
		status := VHostStatus{
			Domain:  domain,
			Enabled: false,
			Checks:  []CheckResult{},
		}

		// Check enabled status
		if enabled, err := drv.IsEnabled(domain); err == nil {
			status.Enabled = enabled
		}

		// Build status message
		var checkMessages []string
		allOK := true

		// Check if enabled status matches config
		if status.Enabled != vhost.Enabled {
			checkMessages = append(checkMessages, fmt.Sprintf("enabled mismatch (config: %v, actual: %v)", vhost.Enabled, status.Enabled))
			status.Checks = append(status.Checks, CheckResult{
				Status:  "warning",
				Message: "enabled status mismatch",
			})
			allOK = false
		}

		// Check root directory exists (if applicable)
		if vhost.Root != "" {
			if _, err := os.Stat(vhost.Root); os.IsNotExist(err) {
				checkMessages = append(checkMessages, "root directory missing")
				status.Checks = append(status.Checks, CheckResult{
					Status:  "warning",
					Message: "root directory missing",
				})
				allOK = false
			}
		}

		// Check SSL certificates exist (if SSL enabled)
		if vhost.SSL {
			if vhost.SSLCert != "" {
				if _, err := os.Stat(vhost.SSLCert); os.IsNotExist(err) {
					checkMessages = append(checkMessages, "SSL certificate missing")
					status.Checks = append(status.Checks, CheckResult{
						Status:  "error",
						Message: "SSL certificate missing",
					})
					allOK = false
				}
			}
			if vhost.SSLKey != "" {
				if _, err := os.Stat(vhost.SSLKey); os.IsNotExist(err) {
					checkMessages = append(checkMessages, "SSL key missing")
					status.Checks = append(status.Checks, CheckResult{
						Status:  "error",
						Message: "SSL key missing",
					})
					allOK = false
				}
			}
		}

		// Add success check if all OK
		if allOK {
			statusText := "disabled"
			if status.Enabled {
				statusText = "enabled"
			}
			status.Checks = append(status.Checks, CheckResult{
				Status:  "success",
				Message: fmt.Sprintf("%s, config valid", statusText),
			})
		}

		statuses = append(statuses, status)
	}

	return statuses
}

func displayDoctorResults(report *DoctorReport) {
	// System requirements
	output.Print("Checking system requirements...")
	for _, check := range report.SystemRequirements {
		displayCheck(check)
	}
	output.Print("")

	// Configuration
	output.Print("Checking configuration...")
	for _, check := range report.Configuration {
		displayCheck(check)
	}
	output.Print("")

	// VHosts
	if len(report.VHosts) > 0 {
		output.Print("Checking vhosts...")
		for _, vhost := range report.VHosts {
			// Get the main check result
			if len(vhost.Checks) > 0 {
				mainCheck := vhost.Checks[len(vhost.Checks)-1]
				switch mainCheck.Status {
				case "success":
					output.Success("%s - %s", vhost.Domain, mainCheck.Message)
				case "warning":
					output.Warn("%s - %s", vhost.Domain, mainCheck.Message)
				case "error":
					output.Error("%s - %s", vhost.Domain, mainCheck.Message)
				}
			}
		}
	} else {
		output.Print("No vhosts configured")
	}
}

func displayCheck(check CheckResult) {
	switch check.Status {
	case "success":
		output.Success("%s", check.Message)
	case "warning":
		output.Warn("%s", check.Message)
	case "error":
		output.Error("%s", check.Message)
	}
}

func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

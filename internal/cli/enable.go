package cli

import (
	"fmt"
	"path/filepath"

	"github.com/ksyq12/vhost/internal/output"
	"github.com/spf13/cobra"
)

var enableCmd = &cobra.Command{
	Use:   "enable <domain>",
	Short: "Enable a virtual host",
	Long: `Enable a virtual host by creating a symlink in sites-enabled.

Examples:
  vhost enable example.com`,
	Args: cobra.ExactArgs(1),
	RunE: runEnable,
}

func init() {
	enableCmd.Flags().BoolVar(&noReload, "no-reload", false, "Don't reload web server")

	rootCmd.AddCommand(enableCmd)
}

func runEnable(cmd *cobra.Command, args []string) error {
	domain := args[0]

	// Validate domain
	if err := validateDomain(domain); err != nil {
		return err
	}

	// Load config and driver
	cfg, drv, err := loadConfigAndDriver()
	if err != nil {
		return err
	}

	// Dry-run mode: show what would be done without making changes
	if dryRun {
		return outputEnableDryRun(domain, drv.Name(), drv.Paths())
	}

	// Require root for system operations
	if err := requireRoot(); err != nil {
		return err
	}

	// Enable via driver
	output.Info("Enabling vhost...")
	if err := drv.Enable(domain); err != nil {
		return fmt.Errorf("failed to enable vhost: %w", err)
	}

	// Test and reload with rollback
	rollback := func() error {
		return drv.Disable(domain)
	}

	if err := testAndReload(drv, !noReload, rollback); err != nil {
		return err
	}

	// Update config
	if vhost, exists := cfg.VHosts[domain]; exists {
		vhost.Enabled = true
		if err := saveConfig(cfg); err != nil {
			output.Warn("VHost enabled but config save failed: %v", err)
		}
	}

	return outputResult(
		map[string]interface{}{
			"success": true,
			"domain":  domain,
			"enabled": true,
		},
		"VHost %s enabled", domain,
	)
}

// outputEnableDryRun outputs what enable command would do in dry-run mode
func outputEnableDryRun(domain string, drvName string, drvPaths struct{ Available, Enabled string }) error {
	// Determine config file name (apache uses .conf extension)
	configFileName := domain
	if drvName == "apache" {
		configFileName = domain + ".conf"
	}

	configPath := filepath.Join(drvPaths.Available, configFileName)
	enabledPath := filepath.Join(drvPaths.Enabled, configFileName)

	operations := []DryRunOperation{
		{
			Action:  "create_symlink",
			Target:  enabledPath,
			Details: fmt.Sprintf("Link to %s", configPath),
		},
	}

	// Add test and reload operations if not --no-reload
	if !noReload {
		operations = append(operations,
			DryRunOperation{
				Action:  "test_config",
				Target:  drvName,
				Details: "Validate configuration syntax",
			},
			DryRunOperation{
				Action:  "reload_server",
				Target:  drvName,
				Details: "Apply configuration changes",
			},
		)
	}

	result := &DryRunResult{
		Domain:     domain,
		Operations: operations,
	}

	return outputDryRun(result)
}

package cli

import (
	"fmt"

	"github.com/ksyq12/vhost/internal/output"
	"github.com/spf13/cobra"
)

var disableCmd = &cobra.Command{
	Use:   "disable <domain>",
	Short: "Disable a virtual host",
	Long: `Disable a virtual host by removing its symlink from sites-enabled.

Examples:
  vhost disable example.com`,
	Args: cobra.ExactArgs(1),
	RunE: runDisable,
}

func init() {
	disableCmd.Flags().BoolVar(&noReload, "no-reload", false, "Don't reload web server")

	rootCmd.AddCommand(disableCmd)
}

func runDisable(cmd *cobra.Command, args []string) error {
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

	// Disable via driver
	output.Info("Disabling vhost...")
	if err := drv.Disable(domain); err != nil {
		return fmt.Errorf("failed to disable vhost: %w", err)
	}

	// Test and reload (no rollback needed for disable)
	if err := testAndReload(drv, !noReload, nil); err != nil {
		output.Warn("Post-disable check failed: %v", err)
		// Continue anyway since vhost is already disabled
	}

	// Update config
	if vhost, exists := cfg.VHosts[domain]; exists {
		vhost.Enabled = false
		if err := saveConfig(cfg); err != nil {
			output.Warn("VHost disabled but config save failed: %v", err)
		}
	}

	return outputResult(
		map[string]interface{}{
			"success":  true,
			"domain":   domain,
			"disabled": true,
		},
		"VHost %s disabled", domain,
	)
}

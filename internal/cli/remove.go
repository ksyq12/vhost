package cli

import (
	"fmt"
	"strings"

	"github.com/ksyq12/vhost/internal/output"
	"github.com/spf13/cobra"
)

var (
	forceRemove bool
)

var removeCmd = &cobra.Command{
	Use:     "remove <domain>",
	Aliases: []string{"rm", "delete"},
	Short:   "Remove a virtual host",
	Long: `Remove a virtual host configuration.

Examples:
  vhost remove example.com
  vhost rm example.com --force`,
	Args: cobra.ExactArgs(1),
	RunE: runRemove,
}

func init() {
	removeCmd.Flags().BoolVarP(&forceRemove, "force", "f", false, "Force removal without confirmation")
	removeCmd.Flags().BoolVar(&noReload, "no-reload", false, "Don't reload web server")

	rootCmd.AddCommand(removeCmd)
}

func runRemove(cmd *cobra.Command, args []string) error {
	domain := args[0]

	// Validate domain
	if err := validateDomain(domain); err != nil {
		return err
	}

	// Require root for system operations
	if err := requireRoot(); err != nil {
		return err
	}

	// Load config and driver
	cfg, drv, err := loadConfigAndDriver()
	if err != nil {
		return err
	}

	// Confirm removal if not forced
	if !forceRemove {
		output.Print("Are you sure you want to remove vhost '%s'? [y/N]: ", domain)
		answer, _ := deps.StdinReader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "y" && answer != "yes" {
			output.Info("Removal cancelled")
			return nil
		}
	}

	// Remove via driver
	output.Info("Removing vhost configuration...")
	if err := drv.Remove(domain); err != nil {
		return fmt.Errorf("failed to remove vhost: %w", err)
	}

	// Test and reload (no rollback for remove)
	if err := testAndReload(drv, !noReload, nil); err != nil {
		output.Warn("Post-removal check failed: %v", err)
		// Continue anyway since vhost is already removed
	}

	// Remove from config
	delete(cfg.VHosts, domain)
	if err := saveConfig(cfg); err != nil {
		output.Warn("VHost removed but config save failed: %v", err)
	}

	return outputResult(
		map[string]interface{}{
			"success": true,
			"domain":  domain,
			"removed": true,
		},
		"VHost %s removed", domain,
	)
}

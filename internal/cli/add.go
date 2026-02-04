package cli

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/ksyq12/vhost/internal/config"
	"github.com/ksyq12/vhost/internal/output"
	"github.com/ksyq12/vhost/internal/template"
	"github.com/spf13/cobra"
)

var (
	vhostType  string
	vhostRoot  string
	proxyPass  string
	phpVersion string
	withSSL    bool
	noReload   bool
)

var addCmd = &cobra.Command{
	Use:   "add <domain>",
	Short: "Add a new virtual host",
	Long: `Add a new virtual host configuration.

Examples:
  vhost add example.com --type static --root /var/www/html
  vhost add example.com --type php --root /var/www/app --php 8.2
  vhost add example.com --type proxy --proxy http://localhost:3000
  vhost add example.com --type laravel --root /var/www/laravel
  vhost add example.com --type wordpress --root /var/www/wordpress`,
	Args: cobra.ExactArgs(1),
	RunE: runAdd,
}

func init() {
	addCmd.Flags().StringVarP(&vhostType, "type", "t", "static", "VHost type (static, php, proxy, laravel, wordpress)")
	addCmd.Flags().StringVarP(&vhostRoot, "root", "r", "", "Document root path")
	addCmd.Flags().StringVarP(&proxyPass, "proxy", "p", "", "Proxy pass URL (for proxy type)")
	addCmd.Flags().StringVar(&phpVersion, "php", "", "PHP version (e.g., 8.2)")
	addCmd.Flags().BoolVar(&withSSL, "ssl", false, "Enable SSL (requires certbot)")
	addCmd.Flags().BoolVar(&noReload, "no-reload", false, "Don't reload web server")

	rootCmd.AddCommand(addCmd)
}

func runAdd(cmd *cobra.Command, args []string) error {
	domain := args[0]

	// Validate domain
	if err := validateDomain(domain); err != nil {
		return err
	}

	// Validate type
	if !config.IsValidType(vhostType) {
		return fmt.Errorf("invalid type: %s. Valid types: %s", vhostType, strings.Join(config.ValidTypes(), ", "))
	}

	// Validate required options based on type
	if err := validateAddOptions(); err != nil {
		return err
	}

	// Load config and driver
	cfg, drv, err := loadConfigAndDriver()
	if err != nil {
		return err
	}

	// Check if vhost already exists
	if _, exists := cfg.VHosts[domain]; exists {
		return fmt.Errorf("vhost %s already exists", domain)
	}

	// Create vhost config
	vhost := &config.VHost{
		Domain:     domain,
		Type:       vhostType,
		Root:       vhostRoot,
		ProxyPass:  proxyPass,
		PHPVersion: phpVersion,
		SSL:        withSSL,
		Enabled:    true,
		CreatedAt:  time.Now(),
	}

	// Set default PHP version if needed
	if vhost.PHPVersion == "" && (vhost.Type == config.TypePHP || vhost.Type == config.TypeLaravel || vhost.Type == config.TypeWordPress) {
		vhost.PHPVersion = cfg.DefaultPHP
	}

	// Render template
	configContent, err := template.Render(drv.Name(), vhost)
	if err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	// Dry-run mode: show what would be done without making changes
	if dryRun {
		drvPaths := drv.Paths()
		return outputAddDryRun(domain, drv.Name(), struct{ Available, Enabled string }{drvPaths.Available, drvPaths.Enabled}, vhost, configContent)
	}

	// Require root for system operations
	if err := requireRoot(); err != nil {
		return err
	}

	// Add vhost via driver
	output.Info("Creating vhost configuration...")
	if err := drv.Add(vhost, configContent); err != nil {
		return fmt.Errorf("failed to add vhost: %w", err)
	}

	// Enable the site
	output.Info("Enabling site...")
	if err := drv.Enable(domain); err != nil {
		// Rollback: remove config file
		_ = drv.Remove(domain)
		return fmt.Errorf("failed to enable vhost: %w", err)
	}

	// Test and reload with proper rollback
	rollback := func() error {
		output.Info("Rolling back changes...")
		if err := drv.Disable(domain); err != nil {
			output.Warn("Rollback disable failed: %v", err)
		}
		if err := drv.Remove(domain); err != nil {
			return fmt.Errorf("rollback remove failed: %w", err)
		}
		return nil
	}

	if err := testAndReload(drv, !noReload, rollback); err != nil {
		return err
	}

	// Save to config
	cfg.VHosts[domain] = vhost
	if err := saveConfig(cfg); err != nil {
		output.Warn("VHost created but config save failed: %v", err)
	}

	return outputResult(
		map[string]interface{}{
			"success": true,
			"domain":  domain,
			"type":    vhostType,
			"enabled": true,
		},
		"VHost %s created and enabled", domain,
	)
}

func validateAddOptions() error {
	switch vhostType {
	case config.TypeStatic, config.TypePHP, config.TypeLaravel, config.TypeWordPress:
		if vhostRoot == "" {
			return fmt.Errorf("--root is required for type %s", vhostType)
		}
		if err := validateRoot(vhostRoot); err != nil {
			return err
		}
	case config.TypeProxy:
		if proxyPass == "" {
			return fmt.Errorf("--proxy is required for type proxy")
		}
		if err := validateProxyURL(proxyPass); err != nil {
			return err
		}
	}
	return nil
}

// outputAddDryRun outputs what add command would do in dry-run mode
func outputAddDryRun(domain string, drvName string, drvPaths struct{ Available, Enabled string }, vhost *config.VHost, configContent string) error {
	paths := drvPaths

	// Determine config file name (apache uses .conf extension)
	configFileName := domain
	if drvName == "apache" {
		configFileName = domain + ".conf"
	}

	configPath := filepath.Join(paths.Available, configFileName)
	enabledPath := filepath.Join(paths.Enabled, configFileName)

	operations := []DryRunOperation{
		{
			Action:  "create_file",
			Target:  configPath,
			Details: fmt.Sprintf("VHost configuration for %s", domain),
		},
		{
			Action:  "create_symlink",
			Target:  enabledPath,
			Details: fmt.Sprintf("Link to %s", configPath),
		},
	}

	// Add document root creation if specified
	if vhost.Root != "" {
		operations = append(operations, DryRunOperation{
			Action:  "create_directory",
			Target:  vhost.Root,
			Details: "Document root directory",
		})
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
		Domain:        domain,
		Operations:    operations,
		ConfigPreview: configContent,
	}

	return outputDryRun(result)
}

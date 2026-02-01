package cli

import (
	"fmt"

	"github.com/ksyq12/vhost/internal/output"
	"github.com/ksyq12/vhost/internal/ssl"
	"github.com/ksyq12/vhost/internal/template"
	"github.com/spf13/cobra"
)

var (
	sslEmail string
)

var sslCmd = &cobra.Command{
	Use:   "ssl",
	Short: "SSL certificate management",
	Long:  `Manage SSL certificates using Let's Encrypt.`,
}

var sslInstallCmd = &cobra.Command{
	Use:   "install <domain>",
	Short: "Install SSL certificate for a domain",
	Long: `Install a Let's Encrypt SSL certificate for a domain.

Examples:
  vhost ssl install example.com --email admin@example.com`,
	Args: cobra.ExactArgs(1),
	RunE: runSSLInstall,
}

var sslRenewCmd = &cobra.Command{
	Use:   "renew [domain]",
	Short: "Renew SSL certificate(s)",
	Long: `Renew SSL certificates.

Examples:
  vhost ssl renew example.com    # Renew specific domain
  vhost ssl renew --all          # Renew all certificates`,
	RunE: runSSLRenew,
}

var sslStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show SSL certificate status",
	Long: `Show the status of all SSL certificates.

Examples:
  vhost ssl status`,
	RunE: runSSLStatus,
}

var (
	renewAll bool
)

func init() {
	sslInstallCmd.Flags().StringVarP(&sslEmail, "email", "e", "", "Email address for Let's Encrypt (required)")
	sslInstallCmd.MarkFlagRequired("email")

	sslRenewCmd.Flags().BoolVar(&renewAll, "all", false, "Renew all certificates")

	sslCmd.AddCommand(sslInstallCmd)
	sslCmd.AddCommand(sslRenewCmd)
	sslCmd.AddCommand(sslStatusCmd)

	rootCmd.AddCommand(sslCmd)
}

func runSSLInstall(cmd *cobra.Command, args []string) error {
	domain := args[0]

	// Validate domain
	if err := validateDomain(domain); err != nil {
		return err
	}

	// Check if certbot is installed
	if !ssl.IsInstalled() {
		return fmt.Errorf("certbot is not installed. Install it with: apt install certbot python3-certbot-nginx")
	}

	// Load config and driver
	cfg, drv, err := loadConfigAndDriver()
	if err != nil {
		return err
	}

	// Get vhost
	vhost, exists := cfg.VHosts[domain]
	if !exists {
		return fmt.Errorf("vhost %s not found. Create it first with: vhost add %s", domain, domain)
	}

	// Issue certificate
	output.Info("Issuing SSL certificate for %s...", domain)
	cert, err := ssl.IssueNginx(domain, sslEmail)
	if err != nil {
		return fmt.Errorf("failed to issue certificate: %w", err)
	}

	// Update vhost config
	vhost.SSL = true
	vhost.SSLCert = cert.CertPath
	vhost.SSLKey = cert.KeyPath

	// Re-render template with SSL
	configContent, err := template.Render(drv.Name(), vhost)
	if err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	// Update config file - disable first, then remove and re-add
	output.Info("Updating vhost configuration with SSL...")
	if enabled, _ := drv.IsEnabled(domain); enabled {
		if err := drv.Disable(domain); err != nil {
			output.Warn("Failed to disable before update: %v", err)
		}
	}

	if err := drv.Remove(domain); err != nil {
		// Not fatal - file might not exist
		output.Warn("Could not remove old config: %v", err)
	}

	if err := drv.Add(vhost, configContent); err != nil {
		return fmt.Errorf("failed to update vhost config: %w", err)
	}

	if err := drv.Enable(domain); err != nil {
		return fmt.Errorf("failed to enable vhost: %w", err)
	}

	// Test and reload
	if err := testAndReload(drv, true, nil); err != nil {
		return err
	}

	// Save config
	if err := saveConfig(cfg); err != nil {
		output.Warn("SSL installed but config save failed: %v", err)
	}

	if jsonOutput {
		return output.JSON(map[string]interface{}{
			"success":   true,
			"domain":    domain,
			"cert_path": cert.CertPath,
			"key_path":  cert.KeyPath,
		})
	}

	output.Success("SSL certificate installed for %s", domain)
	output.Print("  Certificate: %s", cert.CertPath)
	output.Print("  Private Key: %s", cert.KeyPath)

	return nil
}

func runSSLRenew(cmd *cobra.Command, args []string) error {
	if !ssl.IsInstalled() {
		return fmt.Errorf("certbot is not installed")
	}

	if renewAll {
		output.Info("Renewing all certificates...")
		if err := ssl.RenewAll(); err != nil {
			return err
		}
		return outputResult(
			map[string]interface{}{
				"success": true,
				"renewed": "all",
			},
			"All certificates renewed",
		)
	}

	if len(args) == 0 {
		return fmt.Errorf("specify a domain or use --all to renew all certificates")
	}

	domain := args[0]
	if err := validateDomain(domain); err != nil {
		return err
	}

	output.Info("Renewing certificate for %s...", domain)
	if err := ssl.Renew(domain); err != nil {
		return err
	}

	return outputResult(
		map[string]interface{}{
			"success": true,
			"domain":  domain,
			"renewed": true,
		},
		"Certificate renewed for %s", domain,
	)
}

func runSSLStatus(cmd *cobra.Command, args []string) error {
	if !ssl.IsInstalled() {
		return fmt.Errorf("certbot is not installed")
	}

	domains, err := ssl.List()
	if err != nil {
		return err
	}

	if len(domains) == 0 {
		output.Info("No SSL certificates found")
		return nil
	}

	if jsonOutput {
		return output.JSON(domains)
	}

	output.Print("Managed SSL certificates:")
	for _, domain := range domains {
		output.Print("  - %s", domain)
	}

	return nil
}

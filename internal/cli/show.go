package cli

import (
	"fmt"
	"time"

	"github.com/ksyq12/vhost/internal/output"
	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:   "show <domain>",
	Short: "Show details of a virtual host",
	Long: `Show detailed information about a virtual host.

Examples:
  vhost show example.com
  vhost show example.com --json`,
	Args: cobra.ExactArgs(1),
	RunE: runShow,
}

func init() {
	rootCmd.AddCommand(showCmd)
}

// showDetail represents the detailed vhost information for output
type showDetail struct {
	Domain     string     `json:"domain"`
	Type       string     `json:"type"`
	Root       string     `json:"root,omitempty"`
	ProxyPass  string     `json:"proxy_pass,omitempty"`
	PHPVersion string     `json:"php_version,omitempty"`
	SSL        bool       `json:"ssl"`
	SSLCert    string     `json:"ssl_cert,omitempty"`
	SSLKey     string     `json:"ssl_key,omitempty"`
	SSLExpires *time.Time `json:"ssl_expires,omitempty"`
	Enabled    bool       `json:"enabled"`
	CreatedAt  time.Time  `json:"created_at"`
}

func runShow(cmd *cobra.Command, args []string) error {
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

	// Get vhost from config
	vhost, exists := cfg.VHosts[domain]
	if !exists {
		return fmt.Errorf("vhost %s not found", domain)
	}

	// Check enabled status from driver
	enabled, err := drv.IsEnabled(domain)
	if err != nil {
		output.Warn("Could not determine enabled status: %v", err)
	}

	// Build detail struct
	detail := showDetail{
		Domain:     vhost.Domain,
		Type:       vhost.Type,
		Root:       vhost.Root,
		ProxyPass:  vhost.ProxyPass,
		PHPVersion: vhost.PHPVersion,
		SSL:        vhost.SSL,
		SSLCert:    vhost.SSLCert,
		SSLKey:     vhost.SSLKey,
		Enabled:    enabled,
		CreatedAt:  vhost.CreatedAt,
	}

	// Get SSL expiry if SSL is enabled
	if vhost.SSL && vhost.SSLCert != "" {
		if expiry, err := getCertExpiry(vhost.SSLCert); err == nil {
			detail.SSLExpires = &expiry
		}
	}

	// Output JSON if requested
	if jsonOutput {
		return output.JSON(detail)
	}

	// Human-readable output
	output.Print("")
	output.Print("Domain:     %s", detail.Domain)
	output.Print("Type:       %s", detail.Type)

	if detail.Root != "" {
		output.Print("Root:       %s", detail.Root)
	}
	if detail.ProxyPass != "" {
		output.Print("ProxyPass:  %s", detail.ProxyPass)
	}
	if detail.PHPVersion != "" {
		output.Print("PHP:        %s", detail.PHPVersion)
	}

	if detail.SSL {
		output.Print("SSL:        enabled")
		if detail.SSLCert != "" {
			output.Print("  Cert:     %s", detail.SSLCert)
		}
		if detail.SSLKey != "" {
			output.Print("  Key:      %s", detail.SSLKey)
		}
		if detail.SSLExpires != nil {
			output.Print("  Expires:  %s", detail.SSLExpires.Format("2006-01-02"))
		}
	} else {
		output.Print("SSL:        disabled")
	}

	if detail.Enabled {
		output.Print("Enabled:    yes")
	} else {
		output.Print("Enabled:    no")
	}

	output.Print("Created:    %s", detail.CreatedAt.Format("2006-01-02 15:04:05"))
	output.Print("")

	return nil
}

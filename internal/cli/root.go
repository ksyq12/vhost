package cli

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	jsonOutput bool
	version    = "dev"
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "vhost",
	Short: "Virtual host management CLI",
	Long: `vhost is a CLI tool for managing virtual hosts on Nginx, Apache, and Caddy.

It provides commands to add, remove, enable, disable, and list virtual hosts,
as well as SSL certificate management through Let's Encrypt.`,
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// SetVersion sets the version string for the CLI
func SetVersion(v string) {
	version = v
	rootCmd.Version = v
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
}

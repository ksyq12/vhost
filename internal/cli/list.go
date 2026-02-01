package cli

import (
	"sort"

	"github.com/ksyq12/vhost/internal/output"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all virtual hosts",
	Long: `List all configured virtual hosts.

Examples:
  vhost list
  vhost ls
  vhost list --json`,
	RunE: runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

type vhostListItem struct {
	Domain  string `json:"domain"`
	Type    string `json:"type"`
	Root    string `json:"root,omitempty"`
	Proxy   string `json:"proxy,omitempty"`
	SSL     bool   `json:"ssl"`
	Enabled bool   `json:"enabled"`
}

func runList(cmd *cobra.Command, args []string) error {
	// Load config and driver
	cfg, drv, err := loadConfigAndDriver()
	if err != nil {
		return err
	}

	// Get list from driver (to check enabled status)
	driverDomains, err := drv.List()
	if err != nil {
		output.Warn("Could not read from %s: %v", drv.Name(), err)
	}

	// Build list items
	items := make([]vhostListItem, 0)
	for domain, vhost := range cfg.VHosts {
		enabled, _ := drv.IsEnabled(domain)
		items = append(items, vhostListItem{
			Domain:  domain,
			Type:    vhost.Type,
			Root:    vhost.Root,
			Proxy:   vhost.ProxyPass,
			SSL:     vhost.SSL,
			Enabled: enabled,
		})
	}

	// Also add domains found in driver but not in config
	for _, domain := range driverDomains {
		if _, exists := cfg.VHosts[domain]; !exists {
			enabled, _ := drv.IsEnabled(domain)
			items = append(items, vhostListItem{
				Domain:  domain,
				Type:    "unknown",
				Enabled: enabled,
			})
		}
	}

	// Sort by domain
	sort.Slice(items, func(i, j int) bool {
		return items[i].Domain < items[j].Domain
	})

	if len(items) == 0 {
		if jsonOutput {
			return output.JSON([]vhostListItem{})
		}
		output.Info("No virtual hosts configured")
		return nil
	}

	if jsonOutput {
		return output.JSON(items)
	}

	// Build table
	headers := []string{"DOMAIN", "TYPE", "ROOT/PROXY", "SSL", "ENABLED"}
	rows := make([][]string, 0, len(items))

	for _, item := range items {
		rootOrProxy := item.Root
		if item.Proxy != "" {
			rootOrProxy = item.Proxy
		}

		ssl := "no"
		if item.SSL {
			ssl = "yes"
		}

		enabled := "no"
		if item.Enabled {
			enabled = "yes"
		}

		rows = append(rows, []string{
			item.Domain,
			item.Type,
			rootOrProxy,
			ssl,
			enabled,
		})
	}

	output.Table(headers, rows)
	return nil
}

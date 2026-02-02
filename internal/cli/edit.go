package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ksyq12/vhost/internal/output"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit <domain>",
	Short: "Edit a virtual host configuration file",
	Long: `Open the virtual host configuration file in an editor.

Uses $EDITOR environment variable or defaults to vi.

Examples:
  vhost edit example.com
  EDITOR=nano vhost edit example.com`,
	Args: cobra.ExactArgs(1),
	RunE: runEdit,
}

func init() {
	rootCmd.AddCommand(editCmd)
}

func runEdit(cmd *cobra.Command, args []string) error {
	domain := args[0]

	// Validate domain
	if err := validateDomain(domain); err != nil {
		return err
	}

	// Load config and driver
	_, drv, err := loadConfigAndDriver()
	if err != nil {
		return err
	}

	// Build config file path
	configPath := filepath.Join(drv.Paths().Available, domain)
	if drv.Name() == "apache" {
		configPath = filepath.Join(drv.Paths().Available, domain+".conf")
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("config file not found: %s", configPath)
	}

	// Get editor
	editor := getEditor()

	// Check if editor exists
	editorPath, err := exec.LookPath(editor)
	if err != nil {
		return fmt.Errorf("editor not found: %s", editor)
	}

	output.Info("Opening %s with %s...", configPath, editor)

	// Create and run editor command
	editCmd := exec.Command(editorPath, configPath)
	editCmd.Stdin = os.Stdin
	editCmd.Stdout = os.Stdout
	editCmd.Stderr = os.Stderr

	if err := editCmd.Run(); err != nil {
		return fmt.Errorf("editor exited with error: %w", err)
	}

	output.Success("Editor closed")
	output.Info("Run 'vhost test' or reload your web server to apply changes")

	return nil
}

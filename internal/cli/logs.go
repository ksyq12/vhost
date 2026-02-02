package cli

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/ksyq12/vhost/internal/output"
	"github.com/spf13/cobra"
)

var (
	logsAccess bool
	logsError  bool
	logsFollow bool
	logsLines  int
)

var logsCmd = &cobra.Command{
	Use:   "logs <domain>",
	Short: "View logs for a virtual host",
	Long: `View access and error logs for a virtual host.

By default, shows both access and error logs.
Use --access or --error to show only one log type.

Examples:
  vhost logs example.com           # Show both logs
  vhost logs example.com --access  # Show only access log
  vhost logs example.com --error   # Show only error log
  vhost logs example.com -f        # Follow logs in real-time
  vhost logs example.com -n 50     # Show last 50 lines`,
	Args: cobra.ExactArgs(1),
	RunE: runLogs,
}

func init() {
	logsCmd.Flags().BoolVar(&logsAccess, "access", false, "Show access log only")
	logsCmd.Flags().BoolVar(&logsError, "error", false, "Show error log only")
	logsCmd.Flags().BoolVarP(&logsFollow, "follow", "f", false, "Follow log output (like tail -f)")
	logsCmd.Flags().IntVarP(&logsLines, "lines", "n", 20, "Number of lines to show")

	rootCmd.AddCommand(logsCmd)
}

func runLogs(cmd *cobra.Command, args []string) error {
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

	// Check if vhost exists
	if _, exists := cfg.VHosts[domain]; !exists {
		output.Warn("VHost %s not found in config, trying to parse logs anyway", domain)
	}

	// Parse log paths from config
	accessLog, errorLog, err := parseLogPaths(drv, domain)
	if err != nil {
		return fmt.Errorf("failed to get log paths: %w", err)
	}

	// Determine which logs to show
	showAccess := true
	showError := true
	if logsAccess && !logsError {
		showError = false
	} else if logsError && !logsAccess {
		showAccess = false
	}

	// Collect log files to tail
	var logFiles []string
	if showAccess && accessLog != "" {
		if _, err := os.Stat(accessLog); err == nil {
			logFiles = append(logFiles, accessLog)
		} else {
			output.Warn("Access log not found: %s", accessLog)
		}
	}
	if showError && errorLog != "" {
		if _, err := os.Stat(errorLog); err == nil {
			logFiles = append(logFiles, errorLog)
		} else {
			output.Warn("Error log not found: %s", errorLog)
		}
	}

	if len(logFiles) == 0 {
		return fmt.Errorf("no log files found for %s", domain)
	}

	// Build tail command
	tailArgs := []string{}
	if logsFollow {
		tailArgs = append(tailArgs, "-f")
	}
	tailArgs = append(tailArgs, "-n", fmt.Sprintf("%d", logsLines))
	tailArgs = append(tailArgs, logFiles...)

	// Find tail command
	tailPath, err := exec.LookPath("tail")
	if err != nil {
		return fmt.Errorf("tail command not found")
	}

	// Print info about which logs we're showing
	if len(logFiles) == 1 {
		output.Info("Showing logs from: %s", logFiles[0])
	} else {
		output.Info("Showing logs from:")
		for _, f := range logFiles {
			output.Print("  - %s", f)
		}
	}
	output.Print("")

	// Run tail command
	tailCmd := exec.Command(tailPath, tailArgs...)
	tailCmd.Stdin = os.Stdin
	tailCmd.Stdout = os.Stdout
	tailCmd.Stderr = os.Stderr

	if err := tailCmd.Run(); err != nil {
		// Check for interrupt signals (130 = SIGINT/Ctrl+C, 143 = SIGTERM)
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode := exitErr.ExitCode()
			if exitCode == 130 || exitCode == 143 {
				return nil
			}
		}
		return fmt.Errorf("failed to read logs: %w", err)
	}

	return nil
}

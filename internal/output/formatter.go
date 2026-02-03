package output

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
)

var (
	successColor = color.New(color.FgGreen)
	errorColor   = color.New(color.FgRed)
	warnColor    = color.New(color.FgYellow)
	infoColor    = color.New(color.FgCyan)
)

// JSON outputs data as JSON
func JSON(data interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// Table outputs data as a formatted table
func Table(headers []string, rows [][]string) {
	if len(headers) == 0 {
		return
	}

	// Calculate column widths
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Print headers
	headerLine := make([]string, len(headers))
	for i, h := range headers {
		headerLine[i] = fmt.Sprintf("%-*s", widths[i], h)
	}
	fmt.Println(strings.Join(headerLine, "  "))

	// Print separator
	sepLine := make([]string, len(headers))
	for i, w := range widths {
		sepLine[i] = strings.Repeat("-", w)
	}
	fmt.Println(strings.Join(sepLine, "  "))

	// Print rows
	for _, row := range rows {
		rowLine := make([]string, len(headers))
		for i := range headers {
			cell := ""
			if i < len(row) {
				cell = row[i]
			}
			rowLine[i] = fmt.Sprintf("%-*s", widths[i], cell)
		}
		fmt.Println(strings.Join(rowLine, "  "))
	}
}

// Success prints a success message
func Success(format string, args ...interface{}) {
	_, _ = successColor.Printf("✓ "+format+"\n", args...)
}

// Error prints an error message
func Error(format string, args ...interface{}) {
	_, _ = errorColor.Printf("✗ "+format+"\n", args...)
}

// Warn prints a warning message
func Warn(format string, args ...interface{}) {
	_, _ = warnColor.Printf("! "+format+"\n", args...)
}

// Info prints an info message
func Info(format string, args ...interface{}) {
	_, _ = infoColor.Printf("→ "+format+"\n", args...)
}

// Print prints a plain message
func Print(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

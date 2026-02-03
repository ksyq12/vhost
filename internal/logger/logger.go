// Package logger provides leveled logging for the vhost CLI tool.
//
// The logger package outputs debug information to stderr, separate from
// the user-facing output that goes to stdout. This allows for verbose
// debugging without interfering with normal CLI output or JSON formatting.
//
// # Log Levels
//
// Four log levels are supported, in order of severity:
//   - Debug: Detailed information for debugging
//   - Info: General operational information
//   - Warn: Warning conditions that don't prevent operation
//   - Error: Error conditions that affect operation
//
// # Initialization
//
// Initialize the logger based on the --verbose flag:
//
//	logger.Init(verbose)  // verbose=true enables Debug level
//
// By default (verbose=false), only Warn and Error messages are shown.
// When verbose=true, all levels including Debug and Info are shown.
//
// # Usage
//
// Basic logging:
//
//	logger.Debug("Loading config from %s", path)
//	logger.Info("Creating vhost for %s", domain)
//	logger.Warn("Config file not found, using defaults")
//	logger.Error("Failed to reload nginx: %v", err)
//
// Structured logging with fields:
//
//	logger.DebugFields("Config loaded", map[string]interface{}{
//	    "driver": "nginx",
//	    "vhosts": 5,
//	})
//
// # Output Format
//
// Log messages are formatted as:
//
//	[LEVEL] YYYY-MM-DD HH:MM:SS message
//	[DEBUG] 2026-02-03 10:30:45 Loading configuration
//	[INFO] 2026-02-03 10:30:45 Creating vhost for example.com
//
// Structured logs append key=value pairs:
//
//	[DEBUG] 2026-02-03 10:30:45 Config loaded driver=nginx vhosts=5
//
// # Separation of Concerns
//
// The logger is for debugging output (stderr), while the output package
// is for user-facing messages (stdout with colors). Use logger for:
//   - Internal operation details
//   - Debug information
//   - Diagnostic messages
//
// Use output package for:
//   - Success/error messages shown to users
//   - Table output
//   - JSON output
package logger

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

// Level represents a logging severity level.
type Level int

// Log levels from least to most severe.
const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

// String returns the string representation of the log level.
func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Logger handles leveled logging with thread-safe output.
type Logger struct {
	level  Level
	output io.Writer
	mu     sync.Mutex
}

// Global logger instance.
var std = &Logger{
	level:  LevelWarn, // Default: only warnings and errors
	output: os.Stderr,
}

// Init initializes the global logger with the specified verbosity.
// When verbose is true, Debug and Info levels are enabled.
// When verbose is false, only Warn and Error are shown.
func Init(verbose bool) {
	std.mu.Lock()
	defer std.mu.Unlock()

	if verbose {
		std.level = LevelDebug
	} else {
		std.level = LevelWarn
	}
}

// SetLevel sets the minimum log level for the global logger.
func SetLevel(level Level) {
	std.mu.Lock()
	defer std.mu.Unlock()
	std.level = level
}

// SetOutput sets the output destination for the global logger.
// Useful for testing. Default is os.Stderr.
func SetOutput(w io.Writer) {
	std.mu.Lock()
	defer std.mu.Unlock()
	std.output = w
}

// GetLevel returns the current log level.
func GetLevel() Level {
	std.mu.Lock()
	defer std.mu.Unlock()
	return std.level
}

// log writes a formatted message at the specified level.
func (l *Logger) log(level Level, format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if level < l.level {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	msg := fmt.Sprintf(format, args...)
	_, _ = fmt.Fprintf(l.output, "[%s] %s %s\n", level.String(), timestamp, msg)
}

// logFields writes a message with structured key-value fields.
func (l *Logger) logFields(level Level, msg string, fields map[string]interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if level < l.level {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")

	// Sort field keys for consistent output
	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var fieldParts []string
	for _, k := range keys {
		fieldParts = append(fieldParts, fmt.Sprintf("%s=%v", k, fields[k]))
	}

	fieldsStr := ""
	if len(fieldParts) > 0 {
		fieldsStr = " " + strings.Join(fieldParts, " ")
	}

	_, _ = fmt.Fprintf(l.output, "[%s] %s %s%s\n", level.String(), timestamp, msg, fieldsStr)
}

// Debug logs a debug message.
// Only shown when verbose mode is enabled.
func Debug(format string, args ...interface{}) {
	std.log(LevelDebug, format, args...)
}

// Info logs an informational message.
// Only shown when verbose mode is enabled.
func Info(format string, args ...interface{}) {
	std.log(LevelInfo, format, args...)
}

// Warn logs a warning message.
// Always shown regardless of verbose mode.
func Warn(format string, args ...interface{}) {
	std.log(LevelWarn, format, args...)
}

// Error logs an error message.
// Always shown regardless of verbose mode.
func Error(format string, args ...interface{}) {
	std.log(LevelError, format, args...)
}

// DebugFields logs a debug message with structured fields.
func DebugFields(msg string, fields map[string]interface{}) {
	std.logFields(LevelDebug, msg, fields)
}

// InfoFields logs an informational message with structured fields.
func InfoFields(msg string, fields map[string]interface{}) {
	std.logFields(LevelInfo, msg, fields)
}

// WarnFields logs a warning message with structured fields.
func WarnFields(msg string, fields map[string]interface{}) {
	std.logFields(LevelWarn, msg, fields)
}

// ErrorFields logs an error message with structured fields.
func ErrorFields(msg string, fields map[string]interface{}) {
	std.logFields(LevelError, msg, fields)
}

// LogError logs an error with additional context message.
// This is a convenience function for logging errors with extra information.
func LogError(err error, msg string) {
	if err == nil {
		return
	}
	std.log(LevelError, "%s: %v", msg, err)
}

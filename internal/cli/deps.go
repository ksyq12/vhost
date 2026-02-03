package cli

import (
	"bufio"
	"os"

	"github.com/ksyq12/vhost/internal/config"
	"github.com/ksyq12/vhost/internal/driver"
	"github.com/ksyq12/vhost/internal/platform"
)

// Dependencies aggregates all CLI external dependencies for testability
type Dependencies struct {
	ConfigLoader     ConfigLoader
	PlatformDetector PlatformDetector
	DriverFactory    DriverFactory
	RootChecker      RootChecker
	StdinReader      StdinReader
}

// ConfigLoader handles configuration loading and saving
type ConfigLoader interface {
	Load() (*config.Config, error)
	Save(cfg *config.Config) error
}

// PlatformDetector handles platform path detection
type PlatformDetector interface {
	DetectPaths() (*platform.PlatformPaths, error)
}

// DriverFactory creates driver instances
type DriverFactory interface {
	Create(name string, paths driver.Paths) (driver.Driver, error)
}

// RootChecker checks root privileges
type RootChecker interface {
	RequireRoot() error
}

// StdinReader reads from stdin
type StdinReader interface {
	ReadString(delim byte) (string, error)
}

// Package-level dependencies (can be overridden for testing)
var deps = &Dependencies{
	ConfigLoader:     &realConfigLoader{},
	PlatformDetector: &realPlatformDetector{},
	DriverFactory:    &realDriverFactory{},
	RootChecker:      &realRootChecker{},
	StdinReader:      &realStdinReader{},
}

// SetDeps replaces the package dependencies (for testing)
func SetDeps(d *Dependencies) {
	deps = d
}

// GetDeps returns the current dependencies (for testing)
func GetDeps() *Dependencies {
	return deps
}

// Real implementations that delegate to existing functions

type realConfigLoader struct{}

func (r *realConfigLoader) Load() (*config.Config, error) {
	return config.Load()
}

func (r *realConfigLoader) Save(cfg *config.Config) error {
	return cfg.Save()
}

type realPlatformDetector struct{}

func (r *realPlatformDetector) DetectPaths() (*platform.PlatformPaths, error) {
	return platform.DetectPaths()
}

type realDriverFactory struct{}

func (r *realDriverFactory) Create(name string, paths driver.Paths) (driver.Driver, error) {
	return createDriverWithPaths(name, paths)
}

type realRootChecker struct{}

func (r *realRootChecker) RequireRoot() error {
	if os.Geteuid() != 0 {
		return errRootRequired
	}
	return nil
}

type realStdinReader struct {
	reader *bufio.Reader
}

func (r *realStdinReader) ReadString(delim byte) (string, error) {
	if r.reader == nil {
		r.reader = bufio.NewReader(os.Stdin)
	}
	return r.reader.ReadString(delim)
}

// Command runner for edit and logs commands
type CommandRunner interface {
	Run(name string, args ...string) error
	RunInteractive(name string, args ...string) error
	LookPath(file string) (string, error)
}

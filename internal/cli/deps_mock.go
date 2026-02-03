package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/ksyq12/vhost/internal/config"
	"github.com/ksyq12/vhost/internal/driver"
	"github.com/ksyq12/vhost/internal/platform"
)

// MockConfigLoader is a test double for ConfigLoader
type MockConfigLoader struct {
	Cfg       *config.Config
	LoadErr   error
	SaveErr   error
	SaveCalls int
}

func (m *MockConfigLoader) Load() (*config.Config, error) {
	if m.LoadErr != nil {
		return nil, m.LoadErr
	}
	if m.Cfg == nil {
		m.Cfg = config.New()
	}
	return m.Cfg, nil
}

func (m *MockConfigLoader) Save(cfg *config.Config) error {
	m.SaveCalls++
	if m.SaveErr != nil {
		return m.SaveErr
	}
	m.Cfg = cfg
	return nil
}

// MockPlatformDetector is a test double for PlatformDetector
type MockPlatformDetector struct {
	Paths *platform.PlatformPaths
	Err   error
}

func (m *MockPlatformDetector) DetectPaths() (*platform.PlatformPaths, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	if m.Paths != nil {
		return m.Paths, nil
	}
	// Return default mock paths
	return &platform.PlatformPaths{
		Nginx: platform.PathConfig{
			Available: "/etc/nginx/sites-available",
			Enabled:   "/etc/nginx/sites-enabled",
		},
		Apache: platform.PathConfig{
			Available: "/etc/apache2/sites-available",
			Enabled:   "/etc/apache2/sites-enabled",
		},
		Caddy: platform.PathConfig{
			Available: "/etc/caddy/sites-available",
			Enabled:   "/etc/caddy/sites-enabled",
		},
	}, nil
}

// MockDriverFactory is a test double for DriverFactory
type MockDriverFactory struct {
	Driver driver.Driver
	Err    error
}

func (m *MockDriverFactory) Create(name string, paths driver.Paths) (driver.Driver, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	if m.Driver != nil {
		return m.Driver, nil
	}
	// Return a default mock driver if none provided
	return driver.NewMockDriver(name, paths.Available, paths.Enabled), nil
}

// MockRootChecker is a test double for RootChecker
type MockRootChecker struct {
	IsRoot bool
	Calls  int
}

func (m *MockRootChecker) RequireRoot() error {
	m.Calls++
	if !m.IsRoot {
		return errors.New("this operation requires root privileges. Please run with sudo")
	}
	return nil
}

// MockStdinReader is a test double for StdinReader
type MockStdinReader struct {
	Input string
	pos   int
}

func (m *MockStdinReader) ReadString(delim byte) (string, error) {
	if m.pos >= len(m.Input) {
		return "", errors.New("EOF")
	}
	idx := strings.IndexByte(m.Input[m.pos:], delim)
	if idx == -1 {
		result := m.Input[m.pos:]
		m.pos = len(m.Input)
		return result, nil
	}
	result := m.Input[m.pos : m.pos+idx+1]
	m.pos += idx + 1
	return result, nil
}

// MockCommandRunner is a test double for CommandRunner
type MockCommandRunner struct {
	Calls        [][]string
	LookPathFunc func(file string) (string, error)
	RunFunc      func(name string, args ...string) error
	Err          error
}

func (m *MockCommandRunner) Run(name string, args ...string) error {
	m.Calls = append(m.Calls, append([]string{name}, args...))
	if m.RunFunc != nil {
		return m.RunFunc(name, args...)
	}
	return m.Err
}

func (m *MockCommandRunner) RunInteractive(name string, args ...string) error {
	return m.Run(name, args...)
}

func (m *MockCommandRunner) LookPath(file string) (string, error) {
	if m.LookPathFunc != nil {
		return m.LookPathFunc(file)
	}
	if m.Err != nil {
		return "", m.Err
	}
	return "/usr/bin/" + file, nil
}

// MockDependenciesBuilder helps create mock dependencies for tests
type MockDependenciesBuilder struct {
	deps *Dependencies
}

// NewMockDeps creates a new MockDependenciesBuilder with sensible defaults
func NewMockDeps() *MockDependenciesBuilder {
	return &MockDependenciesBuilder{
		deps: &Dependencies{
			ConfigLoader:     &MockConfigLoader{Cfg: config.New()},
			PlatformDetector: &MockPlatformDetector{},
			DriverFactory:    &MockDriverFactory{},
			RootChecker:      &MockRootChecker{IsRoot: true},
			StdinReader:      &MockStdinReader{Input: "y\n"},
		},
	}
}

// WithConfig sets the config for the mock
func (b *MockDependenciesBuilder) WithConfig(cfg *config.Config) *MockDependenciesBuilder {
	b.deps.ConfigLoader = &MockConfigLoader{Cfg: cfg}
	return b
}

// WithConfigLoader sets a custom config loader
func (b *MockDependenciesBuilder) WithConfigLoader(loader ConfigLoader) *MockDependenciesBuilder {
	b.deps.ConfigLoader = loader
	return b
}

// WithDriver sets the driver for the mock
func (b *MockDependenciesBuilder) WithDriver(drv driver.Driver) *MockDependenciesBuilder {
	b.deps.DriverFactory = &MockDriverFactory{Driver: drv}
	return b
}

// WithDriverFactory sets a custom driver factory
func (b *MockDependenciesBuilder) WithDriverFactory(factory DriverFactory) *MockDependenciesBuilder {
	b.deps.DriverFactory = factory
	return b
}

// WithRootAccess sets whether root access is available
func (b *MockDependenciesBuilder) WithRootAccess(isRoot bool) *MockDependenciesBuilder {
	b.deps.RootChecker = &MockRootChecker{IsRoot: isRoot}
	return b
}

// WithStdinInput sets the stdin input for the mock
func (b *MockDependenciesBuilder) WithStdinInput(input string) *MockDependenciesBuilder {
	b.deps.StdinReader = &MockStdinReader{Input: input}
	return b
}

// WithPlatformPaths sets custom platform paths
func (b *MockDependenciesBuilder) WithPlatformPaths(paths *platform.PlatformPaths) *MockDependenciesBuilder {
	b.deps.PlatformDetector = &MockPlatformDetector{Paths: paths}
	return b
}

// WithPlatformError sets an error for platform detection
func (b *MockDependenciesBuilder) WithPlatformError(err error) *MockDependenciesBuilder {
	b.deps.PlatformDetector = &MockPlatformDetector{Err: err}
	return b
}

// Build returns the configured Dependencies
func (b *MockDependenciesBuilder) Build() *Dependencies {
	return b.deps
}

// TestHelper provides utilities for CLI tests
type TestHelper struct {
	T interface {
		Helper()
		Cleanup(func())
	}
	OldDeps    *Dependencies
	MockDriver *driver.MockDriver
	MockConfig *MockConfigLoader
}

// NewTestHelper creates a new test helper with mock dependencies
func NewTestHelper(t interface {
	Helper()
	Cleanup(func())
}, availableDir, enabledDir string) *TestHelper {
	t.Helper()

	mockDriver := driver.NewMockDriver("nginx", availableDir, enabledDir)
	mockConfig := &MockConfigLoader{Cfg: config.New()}

	helper := &TestHelper{
		T:          t,
		OldDeps:    deps,
		MockDriver: mockDriver,
		MockConfig: mockConfig,
	}

	// Set up mock dependencies
	mockDeps := NewMockDeps().
		WithDriver(mockDriver).
		WithConfigLoader(mockConfig).
		Build()

	deps = mockDeps

	// Cleanup function to restore original deps
	t.Cleanup(func() {
		deps = helper.OldDeps
	})

	return helper
}

// SetRootAccess sets whether root access is available
func (h *TestHelper) SetRootAccess(isRoot bool) {
	deps.RootChecker = &MockRootChecker{IsRoot: isRoot}
}

// SetStdinInput sets the stdin input
func (h *TestHelper) SetStdinInput(input string) {
	deps.StdinReader = &MockStdinReader{Input: input}
}

// AddVHost adds a vhost to the mock config
func (h *TestHelper) AddVHost(domain string, vhost *config.VHost) {
	h.MockConfig.Cfg.VHosts[domain] = vhost
}

// GetConfig returns the current mock config
func (h *TestHelper) GetConfig() *config.Config {
	return h.MockConfig.Cfg
}

// errRootRequired is the sentinel error for root privilege check
var errRootRequired = fmt.Errorf("this operation requires root privileges. Please run with sudo")

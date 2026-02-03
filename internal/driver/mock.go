package driver

import (
	"github.com/ksyq12/vhost/internal/config"
)

// MockDriver is a test double for Driver interface
type MockDriver struct {
	name  string
	paths Paths

	// Function mocks - set these to customize behavior
	AddFunc       func(vhost *config.VHost, configContent string) error
	RemoveFunc    func(domain string) error
	EnableFunc    func(domain string) error
	DisableFunc   func(domain string) error
	ListFunc      func() ([]string, error)
	IsEnabledFunc func(domain string) (bool, error)
	TestFunc      func() error
	ReloadFunc    func() error

	// Call tracking - check these to verify interactions
	AddCalls       []AddCall
	RemoveCalls    []string
	EnableCalls    []string
	DisableCalls   []string
	ListCalls      int
	IsEnabledCalls []string
	TestCalls      int
	ReloadCalls    int
}

// AddCall records arguments passed to Add
type AddCall struct {
	VHost   *config.VHost
	Content string
}

// NewMockDriver creates a new MockDriver with default no-op implementations
func NewMockDriver(name, availableDir, enabledDir string) *MockDriver {
	return &MockDriver{
		name: name,
		paths: Paths{
			Available: availableDir,
			Enabled:   enabledDir,
		},
		AddCalls:       make([]AddCall, 0),
		RemoveCalls:    make([]string, 0),
		EnableCalls:    make([]string, 0),
		DisableCalls:   make([]string, 0),
		IsEnabledCalls: make([]string, 0),
	}
}

// Name returns the driver name
func (m *MockDriver) Name() string {
	return m.name
}

// Paths returns the configured paths
func (m *MockDriver) Paths() Paths {
	return m.paths
}

// Add records the call and invokes the mock function if set
func (m *MockDriver) Add(vhost *config.VHost, configContent string) error {
	m.AddCalls = append(m.AddCalls, AddCall{VHost: vhost, Content: configContent})
	if m.AddFunc != nil {
		return m.AddFunc(vhost, configContent)
	}
	return nil
}

// Remove records the call and invokes the mock function if set
func (m *MockDriver) Remove(domain string) error {
	m.RemoveCalls = append(m.RemoveCalls, domain)
	if m.RemoveFunc != nil {
		return m.RemoveFunc(domain)
	}
	return nil
}

// Enable records the call and invokes the mock function if set
func (m *MockDriver) Enable(domain string) error {
	m.EnableCalls = append(m.EnableCalls, domain)
	if m.EnableFunc != nil {
		return m.EnableFunc(domain)
	}
	return nil
}

// Disable records the call and invokes the mock function if set
func (m *MockDriver) Disable(domain string) error {
	m.DisableCalls = append(m.DisableCalls, domain)
	if m.DisableFunc != nil {
		return m.DisableFunc(domain)
	}
	return nil
}

// List records the call and invokes the mock function if set
func (m *MockDriver) List() ([]string, error) {
	m.ListCalls++
	if m.ListFunc != nil {
		return m.ListFunc()
	}
	return []string{}, nil
}

// IsEnabled records the call and invokes the mock function if set
func (m *MockDriver) IsEnabled(domain string) (bool, error) {
	m.IsEnabledCalls = append(m.IsEnabledCalls, domain)
	if m.IsEnabledFunc != nil {
		return m.IsEnabledFunc(domain)
	}
	return false, nil
}

// Test records the call and invokes the mock function if set
func (m *MockDriver) Test() error {
	m.TestCalls++
	if m.TestFunc != nil {
		return m.TestFunc()
	}
	return nil
}

// Reload records the call and invokes the mock function if set
func (m *MockDriver) Reload() error {
	m.ReloadCalls++
	if m.ReloadFunc != nil {
		return m.ReloadFunc()
	}
	return nil
}

// Reset clears all call tracking
func (m *MockDriver) Reset() {
	m.AddCalls = make([]AddCall, 0)
	m.RemoveCalls = make([]string, 0)
	m.EnableCalls = make([]string, 0)
	m.DisableCalls = make([]string, 0)
	m.IsEnabledCalls = make([]string, 0)
	m.ListCalls = 0
	m.TestCalls = 0
	m.ReloadCalls = 0
}

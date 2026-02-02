package executor

import "os/exec"

// CommandExecutor is an interface for executing system commands
type CommandExecutor interface {
	// Execute runs a command with the given name and arguments
	Execute(name string, args ...string) ([]byte, error)

	// LookPath searches for an executable in the directories named by the PATH
	LookPath(file string) (string, error)
}

// SystemExecutor implements CommandExecutor using os/exec
type SystemExecutor struct{}

// NewSystemExecutor creates a new SystemExecutor
func NewSystemExecutor() *SystemExecutor {
	return &SystemExecutor{}
}

// Execute runs a command and returns combined output
func (e *SystemExecutor) Execute(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	return cmd.CombinedOutput()
}

// LookPath searches for an executable
func (e *SystemExecutor) LookPath(file string) (string, error) {
	return exec.LookPath(file)
}

// MockExecutor is a mock implementation for testing
type MockExecutor struct {
	ExecuteFunc  func(name string, args ...string) ([]byte, error)
	LookPathFunc func(file string) (string, error)
	Calls        []CommandCall
}

// CommandCall records a command execution for verification
type CommandCall struct {
	Name string
	Args []string
}

// Execute calls the mock function
func (m *MockExecutor) Execute(name string, args ...string) ([]byte, error) {
	m.Calls = append(m.Calls, CommandCall{Name: name, Args: args})
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(name, args...)
	}
	return []byte(""), nil
}

// LookPath calls the mock function
func (m *MockExecutor) LookPath(file string) (string, error) {
	if m.LookPathFunc != nil {
		return m.LookPathFunc(file)
	}
	return "/usr/bin/" + file, nil
}

package errors

import (
	"errors"
	"fmt"
	"testing"
)

func TestVHostError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *VHostError
		expected string
	}{
		{
			name: "message only",
			err: &VHostError{
				Code:    ErrCodeValidation,
				Message: "invalid input",
			},
			expected: "invalid input",
		},
		{
			name: "with domain",
			err: &VHostError{
				Code:    ErrCodeNotFound,
				Message: "vhost not found",
				Domain:  "example.com",
			},
			expected: "vhost example.com: vhost not found",
		},
		{
			name: "with underlying error",
			err: &VHostError{
				Code:    ErrCodeConfig,
				Message: "failed to load",
				Err:     fmt.Errorf("file not found"),
			},
			expected: "failed to load: file not found",
		},
		{
			name: "with domain and underlying error",
			err: &VHostError{
				Code:    ErrCodeDriver,
				Message: "failed to enable",
				Domain:  "test.com",
				Err:     fmt.Errorf("permission denied"),
			},
			expected: "vhost test.com: failed to enable: permission denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Error()
			if result != tt.expected {
				t.Errorf("Error() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestVHostError_Unwrap(t *testing.T) {
	underlying := fmt.Errorf("underlying error")
	err := &VHostError{
		Code:    ErrCodeConfig,
		Message: "wrapped error",
		Err:     underlying,
	}

	if err.Unwrap() != underlying {
		t.Errorf("Unwrap() did not return underlying error")
	}

	errNoWrap := &VHostError{
		Code:    ErrCodeValidation,
		Message: "no underlying",
	}

	if errNoWrap.Unwrap() != nil {
		t.Errorf("Unwrap() should return nil when no underlying error")
	}
}

func TestVHostError_Is(t *testing.T) {
	tests := []struct {
		name     string
		err      *VHostError
		target   error
		expected bool
	}{
		{
			name:     "matches sentinel error",
			err:      &VHostError{Code: ErrCodeNotFound, Message: "custom message"},
			target:   ErrVHostNotFound,
			expected: true,
		},
		{
			name:     "different code",
			err:      &VHostError{Code: ErrCodeNotFound},
			target:   ErrVHostExists,
			expected: false,
		},
		{
			name:     "non-VHostError target",
			err:      &VHostError{Code: ErrCodeNotFound},
			target:   fmt.Errorf("regular error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if errors.Is(tt.err, tt.target) != tt.expected {
				t.Errorf("Is() = %v, want %v", !tt.expected, tt.expected)
			}
		})
	}
}

func TestNotFound(t *testing.T) {
	err := NotFound("example.com")

	var vhostErr *VHostError
	if !errors.As(err, &vhostErr) {
		t.Fatal("NotFound() should return *VHostError")
	}

	if vhostErr.Code != ErrCodeNotFound {
		t.Errorf("Code = %v, want %v", vhostErr.Code, ErrCodeNotFound)
	}

	if vhostErr.Domain != "example.com" {
		t.Errorf("Domain = %v, want %v", vhostErr.Domain, "example.com")
	}

	if !errors.Is(err, ErrVHostNotFound) {
		t.Error("NotFound() should match ErrVHostNotFound")
	}
}

func TestAlreadyExists(t *testing.T) {
	err := AlreadyExists("test.com")

	var vhostErr *VHostError
	if !errors.As(err, &vhostErr) {
		t.Fatal("AlreadyExists() should return *VHostError")
	}

	if vhostErr.Code != ErrCodeAlreadyExists {
		t.Errorf("Code = %v, want %v", vhostErr.Code, ErrCodeAlreadyExists)
	}

	if vhostErr.Domain != "test.com" {
		t.Errorf("Domain = %v, want %v", vhostErr.Domain, "test.com")
	}

	if !errors.Is(err, ErrVHostExists) {
		t.Error("AlreadyExists() should match ErrVHostExists")
	}
}

func TestValidation(t *testing.T) {
	err := Validation("domain cannot be empty")

	var vhostErr *VHostError
	if !errors.As(err, &vhostErr) {
		t.Fatal("Validation() should return *VHostError")
	}

	if vhostErr.Code != ErrCodeValidation {
		t.Errorf("Code = %v, want %v", vhostErr.Code, ErrCodeValidation)
	}

	if vhostErr.Message != "domain cannot be empty" {
		t.Errorf("Message = %v, want %v", vhostErr.Message, "domain cannot be empty")
	}
}

func TestWrap(t *testing.T) {
	underlying := fmt.Errorf("file not found")
	err := Wrap(ErrCodeConfig, "failed to load config", underlying)

	var vhostErr *VHostError
	if !errors.As(err, &vhostErr) {
		t.Fatal("Wrap() should return *VHostError")
	}

	if vhostErr.Code != ErrCodeConfig {
		t.Errorf("Code = %v, want %v", vhostErr.Code, ErrCodeConfig)
	}

	if vhostErr.Err != underlying {
		t.Error("Wrap() should preserve underlying error")
	}

	if !errors.Is(err, underlying) {
		t.Error("Wrapped error should contain underlying error in chain")
	}
}

func TestWrapDomain(t *testing.T) {
	underlying := fmt.Errorf("symlink failed")
	err := WrapDomain(ErrCodeDriver, "example.com", underlying)

	var vhostErr *VHostError
	if !errors.As(err, &vhostErr) {
		t.Fatal("WrapDomain() should return *VHostError")
	}

	if vhostErr.Code != ErrCodeDriver {
		t.Errorf("Code = %v, want %v", vhostErr.Code, ErrCodeDriver)
	}

	if vhostErr.Domain != "example.com" {
		t.Errorf("Domain = %v, want %v", vhostErr.Domain, "example.com")
	}

	if vhostErr.Err != underlying {
		t.Error("WrapDomain() should preserve underlying error")
	}
}

func TestSentinelErrors(t *testing.T) {
	sentinels := []struct {
		name string
		err  *VHostError
		code ErrorCode
	}{
		{"ErrVHostNotFound", ErrVHostNotFound, ErrCodeNotFound},
		{"ErrVHostExists", ErrVHostExists, ErrCodeAlreadyExists},
		{"ErrInvalidDomain", ErrInvalidDomain, ErrCodeValidation},
		{"ErrInvalidType", ErrInvalidType, ErrCodeValidation},
		{"ErrInvalidPath", ErrInvalidPath, ErrCodeValidation},
		{"ErrPermissionDenied", ErrPermissionDenied, ErrCodePermission},
		{"ErrConfigInvalid", ErrConfigInvalid, ErrCodeConfig},
		{"ErrDriverNotFound", ErrDriverNotFound, ErrCodeDriver},
		{"ErrSSLNotInstalled", ErrSSLNotInstalled, ErrCodeSSL},
		{"ErrRootRequired", ErrRootRequired, ErrCodePermission},
	}

	for _, tt := range sentinels {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Code != tt.code {
				t.Errorf("%s.Code = %v, want %v", tt.name, tt.err.Code, tt.code)
			}
			if tt.err.Message == "" {
				t.Errorf("%s.Message should not be empty", tt.name)
			}
		})
	}
}

func TestErrorChain(t *testing.T) {
	// Create a chain of errors
	root := fmt.Errorf("file not found")
	wrapped := Wrap(ErrCodeConfig, "failed to load", root)
	doubleWrapped := Wrap(ErrCodeInternal, "initialization failed", wrapped)

	// Should be able to unwrap to root
	if !errors.Is(doubleWrapped, root) {
		t.Error("Should be able to find root error in chain")
	}

	// Should match intermediate VHostError
	var configErr *VHostError
	if !errors.As(doubleWrapped, &configErr) {
		t.Error("Should be able to extract VHostError from chain")
	}
}

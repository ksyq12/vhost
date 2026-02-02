// Package errors provides standardized error types for the vhost CLI tool.
//
// The errors package defines domain-specific error types that enable
// structured error handling and consistent error messages throughout
// the application.
//
// # Error Types
//
// VHostError is the primary error type, containing:
//   - Code: Categorizes the error (NOT_FOUND, ALREADY_EXISTS, etc.)
//   - Message: Human-readable error description
//   - Domain: The domain name involved (if applicable)
//   - Err: The underlying wrapped error (if any)
//
// # Sentinel Errors
//
// Common error scenarios have pre-defined sentinel errors:
//
//	errors.ErrVHostNotFound    // vhost doesn't exist
//	errors.ErrVHostExists      // vhost already exists
//	errors.ErrInvalidDomain    // domain validation failed
//	errors.ErrPermissionDenied // root access required
//
// # Usage
//
// Creating domain-specific errors:
//
//	// VHost not found
//	return errors.NotFound("example.com")
//
//	// VHost already exists
//	return errors.AlreadyExists("example.com")
//
//	// Validation error
//	return errors.Validation("domain cannot be empty")
//
//	// Wrapping an underlying error
//	return errors.Wrap(errors.ErrCodeConfig, "failed to load config", err)
//
// # Error Checking
//
// Use errors.Is for sentinel error comparison:
//
//	if errors.Is(err, errors.ErrVHostNotFound) {
//	    // Handle not found case
//	}
//
// Use errors.As for type assertion:
//
//	var vhostErr *errors.VHostError
//	if errors.As(err, &vhostErr) {
//	    fmt.Printf("Error code: %s, Domain: %s\n", vhostErr.Code, vhostErr.Domain)
//	}
package errors

import (
	"errors"
	"fmt"
)

// ErrorCode categorizes errors for programmatic handling.
type ErrorCode string

// Error codes for different error categories.
const (
	ErrCodeNotFound      ErrorCode = "NOT_FOUND"      // Resource not found
	ErrCodeAlreadyExists ErrorCode = "ALREADY_EXISTS" // Resource already exists
	ErrCodeValidation    ErrorCode = "VALIDATION"     // Input validation failed
	ErrCodePermission    ErrorCode = "PERMISSION"     // Permission denied
	ErrCodeConfig        ErrorCode = "CONFIG"         // Configuration error
	ErrCodeDriver        ErrorCode = "DRIVER"         // Web server driver error
	ErrCodeSSL           ErrorCode = "SSL"            // SSL/TLS related error
	ErrCodeInternal      ErrorCode = "INTERNAL"       // Internal/unexpected error
)

// VHostError represents a structured error with context about the operation.
type VHostError struct {
	Code    ErrorCode // Error category
	Message string    // Human-readable message
	Domain  string    // Domain name (if applicable)
	Err     error     // Underlying error (if any)
}

// Error implements the error interface.
func (e *VHostError) Error() string {
	if e.Domain != "" && e.Err != nil {
		return fmt.Sprintf("vhost %s: %s: %v", e.Domain, e.Message, e.Err)
	}
	if e.Domain != "" {
		return fmt.Sprintf("vhost %s: %s", e.Domain, e.Message)
	}
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the underlying error for error chain traversal.
func (e *VHostError) Unwrap() error {
	return e.Err
}

// Is reports whether target matches this error.
// Comparison is based on error code.
func (e *VHostError) Is(target error) bool {
	t, ok := target.(*VHostError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// Sentinel errors for common error scenarios.
// Use these with errors.Is() for error checking.
var (
	// ErrVHostNotFound indicates the requested vhost does not exist.
	ErrVHostNotFound = &VHostError{Code: ErrCodeNotFound, Message: "vhost not found"}

	// ErrVHostExists indicates a vhost with the same domain already exists.
	ErrVHostExists = &VHostError{Code: ErrCodeAlreadyExists, Message: "vhost already exists"}

	// ErrInvalidDomain indicates the domain name is not valid.
	ErrInvalidDomain = &VHostError{Code: ErrCodeValidation, Message: "invalid domain"}

	// ErrInvalidType indicates the vhost type is not valid.
	ErrInvalidType = &VHostError{Code: ErrCodeValidation, Message: "invalid vhost type"}

	// ErrInvalidPath indicates a file path is not valid.
	ErrInvalidPath = &VHostError{Code: ErrCodeValidation, Message: "invalid path"}

	// ErrPermissionDenied indicates insufficient privileges for the operation.
	ErrPermissionDenied = &VHostError{Code: ErrCodePermission, Message: "permission denied"}

	// ErrConfigInvalid indicates the configuration is invalid or corrupt.
	ErrConfigInvalid = &VHostError{Code: ErrCodeConfig, Message: "invalid configuration"}

	// ErrDriverNotFound indicates the specified driver is not available.
	ErrDriverNotFound = &VHostError{Code: ErrCodeDriver, Message: "driver not found"}

	// ErrSSLNotInstalled indicates certbot is not installed.
	ErrSSLNotInstalled = &VHostError{Code: ErrCodeSSL, Message: "certbot not installed"}

	// ErrRootRequired indicates root privileges are required.
	ErrRootRequired = &VHostError{Code: ErrCodePermission, Message: "root privileges required"}
)

// NotFound creates an error for a vhost that doesn't exist.
func NotFound(domain string) error {
	return &VHostError{
		Code:    ErrCodeNotFound,
		Message: "vhost not found",
		Domain:  domain,
	}
}

// AlreadyExists creates an error for a vhost that already exists.
func AlreadyExists(domain string) error {
	return &VHostError{
		Code:    ErrCodeAlreadyExists,
		Message: "vhost already exists",
		Domain:  domain,
	}
}

// Validation creates a validation error with a custom message.
func Validation(msg string) error {
	return &VHostError{
		Code:    ErrCodeValidation,
		Message: msg,
	}
}

// Wrap creates an error with the specified code, message, and underlying error.
func Wrap(code ErrorCode, msg string, err error) error {
	return &VHostError{
		Code:    code,
		Message: msg,
		Err:     err,
	}
}

// WrapDomain creates an error with domain context and underlying error.
func WrapDomain(code ErrorCode, domain string, err error) error {
	return &VHostError{
		Code:   code,
		Domain: domain,
		Err:    err,
	}
}

// Is reports whether any error in err's chain matches target.
// This is a re-export of errors.Is for convenience.
var Is = errors.Is

// As finds the first error in err's chain that matches target.
// This is a re-export of errors.As for convenience.
var As = errors.As

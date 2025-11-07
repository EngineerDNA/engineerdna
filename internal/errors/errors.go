package errors

import (
	"fmt"
	"time"
)

// ErrorCode represents a standardized error code
type ErrorCode string

const (
	ErrorAuth        ErrorCode = "auth_failed"
	ErrorRateLimit   ErrorCode = "rate_limit"
	ErrorNetwork     ErrorCode = "network_error"
	ErrorConfig      ErrorCode = "config_invalid"
	ErrorPermission  ErrorCode = "permission_denied"
	ErrorNotFound    ErrorCode = "not_found"
	ErrorInternal    ErrorCode = "internal_error"
	ErrorTimeout     ErrorCode = "timeout"
	ErrorPluginCrash ErrorCode = "plugin_crash"
	ErrorValidation  ErrorCode = "validation_error"
)

// AppError represents a structured application error with user-friendly messaging
type AppError struct {
	Code        ErrorCode              `json:"code"`
	Message     string                 `json:"message"`
	UserMessage string                 `json:"user_message"`
	Suggestions []string               `json:"suggestions,omitempty"`
	DocsURL     string                 `json:"docs_url,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// NewAuthError creates an authentication error
func NewAuthError(provider string) *AppError {
	return &AppError{
		Code:        ErrorAuth,
		Message:     fmt.Sprintf("authentication failed for %s", provider),
		UserMessage: fmt.Sprintf("Failed to authenticate with %s. Please check your credentials.", provider),
		Suggestions: []string{
			fmt.Sprintf("Verify your %s API key or token is correct", provider),
			"Check if your credentials have expired",
			"Ensure you have the necessary permissions",
		},
		DocsURL: "https://github.com/engineerdna/engineerdna#authentication",
		Context: map[string]interface{}{
			"provider": provider,
		},
	}
}

// NewRateLimitError creates a rate limit error
func NewRateLimitError(provider string, resetTime time.Time) *AppError {
	waitMinutes := int(time.Until(resetTime).Minutes())
	if waitMinutes < 1 {
		waitMinutes = 1
	}

	return &AppError{
		Code:        ErrorRateLimit,
		Message:     fmt.Sprintf("rate limit exceeded for %s", provider),
		UserMessage: fmt.Sprintf("You've hit the rate limit for %s. Please wait before trying again.", provider),
		Suggestions: []string{
			fmt.Sprintf("Wait approximately %d minutes before retrying", waitMinutes),
			"Consider upgrading your API plan for higher limits",
			"Reduce the frequency of sync operations",
		},
		DocsURL: "https://github.com/engineerdna/engineerdna#rate-limits",
		Context: map[string]interface{}{
			"provider":   provider,
			"reset_time": resetTime.Format(time.RFC3339),
		},
	}
}

// NewNetworkError creates a network connectivity error
func NewNetworkError(endpoint string) *AppError {
	return &AppError{
		Code:        ErrorNetwork,
		Message:     fmt.Sprintf("network error connecting to %s", endpoint),
		UserMessage: "Unable to connect to the service. Please check your internet connection.",
		Suggestions: []string{
			"Check your internet connection",
			"Verify the service is not experiencing an outage",
			"Check if you're behind a firewall or proxy",
			"Try again in a few moments",
		},
		DocsURL: "https://github.com/engineerdna/engineerdna#troubleshooting",
		Context: map[string]interface{}{
			"endpoint": endpoint,
		},
	}
}

// NewConfigError creates a configuration validation error
func NewConfigError(field string, reason string) *AppError {
	return &AppError{
		Code:        ErrorConfig,
		Message:     fmt.Sprintf("invalid configuration: %s - %s", field, reason),
		UserMessage: fmt.Sprintf("Configuration error in '%s': %s", field, reason),
		Suggestions: []string{
			"Review your configuration settings",
			"Ensure all required fields are filled",
			"Check the format of the field value",
		},
		DocsURL: "https://github.com/engineerdna/engineerdna#configuration",
		Context: map[string]interface{}{
			"field":  field,
			"reason": reason,
		},
	}
}

// NewPermissionError creates a permission denied error
func NewPermissionError(resource string, action string) *AppError {
	return &AppError{
		Code:        ErrorPermission,
		Message:     fmt.Sprintf("permission denied: cannot %s %s", action, resource),
		UserMessage: fmt.Sprintf("You don't have permission to %s %s.", action, resource),
		Suggestions: []string{
			"Verify your API key has the necessary permissions",
			"Check if your account has access to this resource",
			"Contact your administrator for access",
		},
		DocsURL: "https://github.com/engineerdna/engineerdna#permissions",
		Context: map[string]interface{}{
			"resource": resource,
			"action":   action,
		},
	}
}

// NewNotFoundError creates a resource not found error
func NewNotFoundError(resourceType string, identifier string) *AppError {
	return &AppError{
		Code:        ErrorNotFound,
		Message:     fmt.Sprintf("%s not found: %s", resourceType, identifier),
		UserMessage: fmt.Sprintf("The %s '%s' could not be found.", resourceType, identifier),
		Suggestions: []string{
			"Check if the identifier is correct",
			"Verify the resource exists",
			"Try refreshing the page",
		},
		DocsURL: "https://github.com/engineerdna/engineerdna#resources",
		Context: map[string]interface{}{
			"resource_type": resourceType,
			"identifier":    identifier,
		},
	}
}

// NewInternalError creates an internal server error
func NewInternalError(operation string, err error) *AppError {
	return &AppError{
		Code:        ErrorInternal,
		Message:     fmt.Sprintf("internal error during %s: %v", operation, err),
		UserMessage: "An unexpected error occurred. Please try again.",
		Suggestions: []string{
			"Try the operation again",
			"If the problem persists, check the logs",
			"Report this issue if it continues",
		},
		DocsURL: "https://github.com/engineerdna/engineerdna#troubleshooting",
		Context: map[string]interface{}{
			"operation": operation,
		},
	}
}

// NewTimeoutError creates a timeout error
func NewTimeoutError(operation string, timeout time.Duration) *AppError {
	return &AppError{
		Code:        ErrorTimeout,
		Message:     fmt.Sprintf("operation timed out: %s (timeout: %v)", operation, timeout),
		UserMessage: fmt.Sprintf("The operation took too long and timed out after %v.", timeout),
		Suggestions: []string{
			"Try the operation again",
			"The service may be experiencing high load",
			"Consider breaking the operation into smaller chunks",
		},
		DocsURL: "https://github.com/engineerdna/engineerdna#timeouts",
		Context: map[string]interface{}{
			"operation": operation,
			"timeout":   timeout.String(),
		},
	}
}

// NewPluginCrashError creates a plugin crash error
func NewPluginCrashError(pluginName string, stderr string) *AppError {
	suggestions := []string{
		"Check the plugin logs for details",
		"Try reconfiguring the plugin",
		"Verify the plugin binary is not corrupted",
	}

	if stderr != "" {
		suggestions = append([]string{
			"Check the error details below for more information",
		}, suggestions...)
	}

	return &AppError{
		Code:        ErrorPluginCrash,
		Message:     fmt.Sprintf("plugin %s crashed unexpectedly", pluginName),
		UserMessage: fmt.Sprintf("The %s plugin encountered an error and stopped working.", pluginName),
		Suggestions: suggestions,
		DocsURL:     "https://github.com/engineerdna/engineerdna#plugin-errors",
		Context: map[string]interface{}{
			"plugin": pluginName,
			"stderr": stderr,
		},
	}
}

// NewValidationError creates a validation error
func NewValidationError(field string, value string, constraint string) *AppError {
	return &AppError{
		Code:        ErrorValidation,
		Message:     fmt.Sprintf("validation failed for %s: %s", field, constraint),
		UserMessage: fmt.Sprintf("Invalid value for '%s': %s", field, constraint),
		Suggestions: []string{
			"Check the field format and try again",
			"Refer to the documentation for valid values",
		},
		DocsURL: "https://github.com/engineerdna/engineerdna#validation",
		Context: map[string]interface{}{
			"field":      field,
			"value":      value,
			"constraint": constraint,
		},
	}
}

// WrapError wraps an existing error with additional context
func WrapError(err error, code ErrorCode, userMessage string, suggestions []string) *AppError {
	if appErr, ok := err.(*AppError); ok {
		return appErr
	}

	return &AppError{
		Code:        code,
		Message:     err.Error(),
		UserMessage: userMessage,
		Suggestions: suggestions,
		DocsURL:     "https://github.com/engineerdna/engineerdna#troubleshooting",
	}
}

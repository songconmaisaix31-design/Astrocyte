// Package apierrors defines structured error types shared across
// bounded contexts and the HTTP transport layer.
package apierrors

import "fmt"

// SchemaVersion is the error envelope schema version.
const SchemaVersion = 1

// Code enumerates the structured error codes used across the API.
type Code string

const (
	NotFound              Code = "not_found"
	ValidationFailed      Code = "validation_failed"
	UnsupportedCapability Code = "unsupported_capability"
	VersionConflict       Code = "version_conflict"
	ContextStale          Code = "context_stale"
	IndexStale            Code = "index_stale"
	ScopeDenied           Code = "scope_denied"
	ApprovalRequired      Code = "approval_required"
	ApprovalRevoked       Code = "approval_revoked"
	LeaseLost             Code = "lease_lost"
	BudgetExhausted       Code = "budget_exhausted"
	ProviderUnavailable   Code = "provider_unavailable"
	EvidenceMissing       Code = "evidence_missing"
	DeliveryUnknown       Code = "delivery_unknown"
	InternalError         Code = "internal_error"
)

// ServiceError is the domain error type returned by application services.
type ServiceError struct {
	Code           Code   `json:"code"`
	Message        string `json:"message"`
	Retryable      bool   `json:"retryable"`
	RequestID      string `json:"request_id"`
	RequiredAction string `json:"required_action"`
}

func (e *ServiceError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// NewUnsupported creates a 501 error for unimplemented S0 capabilities.
func NewUnsupported(capability string) *ServiceError {
	return &ServiceError{
		Code:           UnsupportedCapability,
		Message:        fmt.Sprintf("capability %q is not available in S0", capability),
		Retryable:      false,
		RequiredAction: "Read the current S0 capabilities; choose an implemented operation",
	}
}

// NewNotFound creates a 404 error for absent resources.
func NewNotFound(resource, id string) *ServiceError {
	return &ServiceError{
		Code:           NotFound,
		Message:        fmt.Sprintf("%s %q not found", resource, id),
		Retryable:      false,
		RequiredAction: "check_id_and_retry",
	}
}

// ErrorEnvelope wraps a ServiceError in the OpenAPI ErrorV1 response shape:
// {"schema_version": 1, "error": {...}}
type ErrorEnvelope struct {
	SchemaVersion int           `json:"schema_version"`
	Error         *ServiceError `json:"error"`
}

// Wrap returns an ErrorEnvelope for the given ServiceError.
func Wrap(err *ServiceError) ErrorEnvelope {
	return ErrorEnvelope{
		SchemaVersion: SchemaVersion,
		Error:         err,
	}
}

// ListResult is the standard paginated list response envelope,
// shared across all bounded contexts.
type ListResult struct {
	SchemaVersion int     `json:"schema_version"`
	Items         []any   `json:"items"`
	NextCursor    *string `json:"next_cursor"`
}

// EmptyList returns an empty list result with the current schema version.
func EmptyList() ListResult {
	return ListResult{
		SchemaVersion: SchemaVersion,
		Items:         []any{},
		NextCursor:    nil,
	}
}

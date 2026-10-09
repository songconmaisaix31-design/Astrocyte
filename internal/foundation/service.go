// Package foundation provides the neutral foundation service for S0.
//
// This service is not part of any bounded context. It provides
// health checks and foundation status reporting.
package foundation

import "context"

// SchemaVersion is the API response schema version for S0.
const SchemaVersion = 1

// HealthResponse is the health check response.
type HealthResponse struct {
	SchemaVersion int    `json:"schema_version"`
	Status        string `json:"status"`
	Service       string `json:"service"`
}

// FoundationResponse describes the current foundation state.
type FoundationResponse struct {
	SchemaVersion int          `json:"schema_version"`
	Stage         string       `json:"stage"`
	Capabilities  Capabilities `json:"capabilities"`
	Storage       StorageInfo  `json:"storage"`
	Fixture       bool         `json:"fixture"`
}

// Capabilities lists which feature flags are active.
type Capabilities struct {
	Imports      bool `json:"imports"`
	Approvals    bool `json:"approvals"`
	Execution    bool `json:"execution"`
	NativeResume bool `json:"native_resume"`
	Handoff      bool `json:"handoff"`
}

// StorageInfo describes the storage backend.
type StorageInfo struct {
	Engine        string `json:"engine"`
	SchemaVersion int    `json:"schema_version"`
}

// Service provides the foundation status endpoints.
type Service interface {
	Health(ctx context.Context) (HealthResponse, error)
	Foundation(ctx context.Context) (FoundationResponse, error)
}

type service struct {
	storageSchemaVersion int
	attention            bool
}

// NewService creates the S0 foundation service.
func NewService(storageSchemaVersion int) Service {
	return &service{storageSchemaVersion: storageSchemaVersion}
}

// NewAttentionService reports the assembled S1 import capability; it does not
// imply approvals, execution, automatic judgment or native agent readiness.
func NewAttentionService(storageSchemaVersion int) Service {
	return &service{storageSchemaVersion: storageSchemaVersion, attention: true}
}

func (s *service) Health(_ context.Context) (HealthResponse, error) {
	return HealthResponse{
		SchemaVersion: SchemaVersion,
		Status:        "ok",
		Service:       "astrocyte",
	}, nil
}

func (s *service) Foundation(_ context.Context) (FoundationResponse, error) {
	stage := "S0"
	if s.attention {
		stage = "S1"
	}
	return FoundationResponse{
		SchemaVersion: SchemaVersion,
		Stage:         stage,
		Capabilities: Capabilities{
			Imports:      s.attention,
			Approvals:    false,
			Execution:    false,
			NativeResume: false,
			Handoff:      false,
		},
		Storage: StorageInfo{
			Engine:        "sqlite",
			SchemaVersion: s.storageSchemaVersion,
		},
		Fixture: false,
	}, nil
}

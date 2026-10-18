package app

import (
	"context"
	"encoding/json"

	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

type OpenVikingStatus struct {
	SchemaVersion int    `json:"schema_version"`
	Configured    bool   `json:"configured"`
	Connected     bool   `json:"connected"`
	BaseURL       string `json:"base_url"`
	RootURI       string `json:"root_uri"`
	Version       string `json:"version"`
	User          string `json:"user"`
	Account       string `json:"account"`
	Message       string `json:"message"`
}

type OpenVikingHit struct {
	URI         string  `json:"uri"`
	Abstract    string  `json:"abstract"`
	Overview    *string `json:"overview"`
	Level       int     `json:"level"`
	Score       float64 `json:"score"`
	ContextType string  `json:"context_type"`
}

type OpenVikingTask struct {
	TaskID          string          `json:"task_id"`
	Status          string          `json:"status"`
	ResourceID      string          `json:"resource_id"`
	Stage           *string         `json:"stage"`
	Error           *string         `json:"error"`
	Result          json.RawMessage `json:"result"`
	ExecutionEvents json.RawMessage `json:"execution_events"`
}

type OpenVikingGateway interface {
	Status(context.Context) (OpenVikingStatus, error)
	Upload(context.Context, string, string) (string, error)
	Submit(context.Context, string, string, string) (string, string, error)
	Task(context.Context, string) (OpenVikingTask, error)
	Find(context.Context, string, string) ([]OpenVikingHit, error)
	Read(context.Context, string, int) (string, error)
}

type OpenVikingRepository interface {
	RecoverOpenViking(context.Context) error
	ReserveOpenViking(context.Context, domain.OpenVikingSubmission) (domain.OpenVikingSubmission, bool, error)
	SaveOpenViking(context.Context, domain.OpenVikingSubmission, int) error
	GetOpenViking(context.Context, string) (domain.OpenVikingSubmission, error)
	ListOpenViking(context.Context, string) ([]domain.OpenVikingSubmission, error)
}

type OpenVikingSubmitCommand struct {
	SchemaVersion        int    `json:"schema_version"`
	RequestID            string `json:"request_id"`
	IdempotencyKey       string `json:"-"`
	ExpectedVersion      *int   `json:"expected_version"`
	ExpectedSpaceVersion int    `json:"expected_space_version"`
	Account              string `json:"account"`
	User                 string `json:"user"`
	MaterialID           string `json:"material_id"`
	Revision             int    `json:"revision"`
	Destination          string `json:"destination"`
	BaseURL              string `json:"base_url"`
}

type OpenVikingSearchCommand struct {
	Query string `json:"query"`
}

type OpenVikingContentCommand struct {
	URI   string `json:"uri"`
	Level *int   `json:"level"`
}

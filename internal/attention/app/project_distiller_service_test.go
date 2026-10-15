package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/distillers"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/importers"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	. "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

func TestGlobalSelectedProcessorStatusDoesNotInventAvailability(t *testing.T) {
	_, db, objects := fixture(t)
	factory := distillers.NewSelectedTextFactory(nil, time.Minute)
	s := NewAttentionService(sqlite.NewAttentionRepository(db), importers.NewReader(nil), objects, ServiceOptions{ProjectDistillers: factory})
	status, err := s.GetDistillerStatus(context.Background(), human)
	if err != nil || status.Available || status.Processor != "selected-project-cli" || status.ConfigurationID != nil || status.Model != nil || status.RequiredAction != "select_permitted_project_cli" {
		t.Fatal("registered production factory invented selected project or native availability")
	}
}

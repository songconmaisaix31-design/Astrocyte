package attention_engine_test

import (
	"context"
	"encoding/json"
	"math"
	"path/filepath"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

func TestM4ExplicitAgeBasisDoesNotUseHumanGovernanceTime(t *testing.T) {
	ctx := context.Background()
	db, _, service := openM4(t, filepath.Join(t.TempDir(), "knowledge.sqlite"))
	defer db.Close()
	source := seedM4Source(t, db, "source", "p1")
	u := app.KnowledgeFluidUpdate{ID: "k", OwnerProject: "p1", Topic: "agents", Fluid: "first", Sources: []domain.EventRef{source}, TS: 2000}
	if _, err := service.RewriteFluid(ctx, u); err != nil {
		t.Fatal(err)
	}
	u.Fluid, u.ExpectedVersion, u.TS = "second", 1, 3000
	if _, err := service.RewriteFluid(ctx, u); err != nil {
		t.Fatal(err)
	}
	human := app.Principal{ID: "human", Kind: "human"}
	if _, err := service.ConfirmKnowledge(ctx, human, "k", app.ConfirmKnowledgeCommand{CommandMeta: m4Meta(2, "confirm"), OwnerProject: "p1"}); err != nil {
		t.Fatal(err)
	}
	entry, err := service.GetKnowledge(ctx, human, "k", "p1")
	if err != nil || entry.ConsolidatedAt == nil || *entry.ConsolidatedAt != 3000 || entry.UpdatedAt != 5000 {
		t.Fatalf("separate content/governance timestamps: %+v %v", entry, err)
	}
	data, _ := json.Marshal(domain.AttentionNode{ID: "agents", Topic: "agents", Attention: 0.6})
	if _, err = db.Conn().Exec(`INSERT INTO attention_engine_nodes(id,topic,attention,momentum,confidence,last_active,data) VALUES('agents','agents',.6,0,0,1000,?)`, string(data)); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Conn().Exec(`PRAGMA query_only=ON`); err != nil {
		t.Fatal(err)
	}
	lambda := 0.1
	for _, fixture := range []struct {
		basis app.KnowledgeAgeBasis
		at    int64
		age   float64
	}{{app.KnowledgeAgeCreated, 2000, 8}, {app.KnowledgeAgeConsolidated, 3000, 7}} {
		matches, err := service.RetrieveKnowledge(ctx, human, "p1", app.KnowledgeQuery{Topic: "agents", Now: 10000, LambdaR: &lambda, AgeBasis: fixture.basis})
		if err != nil || len(matches) != 1 {
			t.Fatalf("age basis %s: %+v %v", fixture.basis, matches, err)
		}
		if matches[0].AgeAt != fixture.at || math.Abs(matches[0].Score-0.6*math.Exp(-lambda*fixture.age)) > 1e-12 {
			t.Fatalf("age basis %s used human action time: %+v", fixture.basis, matches[0])
		}
	}
}

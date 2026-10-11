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

func TestM4PersistentRetrievalRecomputedAndAttentionUnchanged(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "knowledge.sqlite")
	db, _, service := openM4(t, path)
	defer func() { db.Close() }()
	human := app.Principal{ID: "human", Kind: "human"}
	for _, fixture := range []struct {
		id, topic, project string
		at                 int64
		confirm            bool
	}{
		{"a", "agents", "p1", 2000, true},
		{"b", "systems", "p1", 4000, true},
		{"hidden", "agents", "p2", 2000, true},
		{"inbox", "agents", "p1", 2000, false},
	} {
		source := seedM4Source(t, db, "source-"+fixture.id, fixture.project)
		if _, err := service.RewriteFluid(ctx, app.KnowledgeFluidUpdate{ID: fixture.id, OwnerProject: fixture.project, Topic: fixture.topic, Fluid: "research", Sources: []domain.EventRef{source}, TS: fixture.at}); err != nil {
			t.Fatal(err)
		}
		if fixture.confirm {
			if _, err := service.ConfirmKnowledge(ctx, human, fixture.id, app.ConfirmKnowledgeCommand{CommandMeta: m4Meta(1, "confirm-"+fixture.id), OwnerProject: fixture.project}); err != nil {
				t.Fatal(err)
			}
		}
	}
	if _, err := service.ReferenceKnowledge(ctx, human, "a", app.ReferenceKnowledgeCommand{CommandMeta: m4Meta(2, "reference-a"), ProjectID: "p1"}); err != nil {
		t.Fatal(err)
	}
	for _, node := range []domain.AttentionNode{
		{ID: "agents", Topic: "agents", Attention: 0.6, Momentum: 0.1, Confidence: 0.8, LastActive: 1000, RelatedTopics: map[string]float64{}},
		{ID: "systems", Topic: "systems", Attention: 0.9, Momentum: 0.05, Confidence: 0.7, LastActive: 1000, RelatedTopics: map[string]float64{"agents": 0.25}},
	} {
		data, _ := json.Marshal(node)
		if _, err := db.Conn().Exec(`INSERT INTO attention_engine_nodes(id,topic,attention,momentum,confidence,last_active,data,updated_at,version) VALUES(?,?,?,?,?,?,?,?,?)`, node.ID, node.Topic, node.Attention, node.Momentum, node.Confidence, node.LastActive, string(data), 5000, 1); err != nil {
			t.Fatal(err)
		}
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db, _, service = openM4(t, path)
	var before string
	if err := db.Conn().QueryRow(`SELECT group_concat(data || ':' || attention || ':' || updated_at || ':' || version,'|') FROM (SELECT * FROM attention_engine_nodes ORDER BY id)`).Scan(&before); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Conn().Exec(`PRAGMA query_only=ON`); err != nil {
		t.Fatal(err)
	}
	if _, err := service.RetrieveKnowledge(ctx, human, "p1", app.KnowledgeQuery{Now: 10000, Topic: "agents"}); err == nil {
		t.Fatal("retrieval froze missing lambda/age configuration")
	}
	for _, fixture := range []struct {
		lambda float64
		winner string
	}{{0, "a"}, {1.5, "b"}} {
		lambda := fixture.lambda
		q := app.KnowledgeQuery{Topic: "agents", Now: 10000, LambdaR: &lambda, AgeBasis: app.KnowledgeAgeCreated, ReferenceBoost: 2}
		matches, err := service.RetrieveKnowledge(ctx, human, "p1", q)
		if err != nil {
			t.Fatal(err)
		}
		if len(matches) != 2 || matches[0].Knowledge.Entry.ID != fixture.winner {
			t.Fatalf("lambda=%g scoped rank: %+v", lambda, matches)
		}
		manual := map[string]float64{"a": 1 * math.Exp(-lambda*8) * 0.6 * 3, "b": 0.25 * math.Exp(-lambda*6) * 0.9 * 1}
		for _, match := range matches {
			if math.Abs(match.Score-manual[match.Knowledge.Entry.ID]) > 1e-12 {
				t.Fatalf("manual score: %+v want=%g", match, manual[match.Knowledge.Entry.ID])
			}
			if match.RelevanceMode != "sqlite_vec_cut:topic_keyword" {
				t.Fatalf("CUT hidden: %s", match.RelevanceMode)
			}
		}
	}
	var after string
	if err := db.Conn().QueryRow(`SELECT group_concat(data || ':' || attention || ':' || updated_at || ':' || version,'|') FROM (SELECT * FROM attention_engine_nodes ORDER BY id)`).Scan(&after); err != nil {
		t.Fatal(err)
	}
	if before != after {
		t.Fatalf("retrieval changed attention SQLite state:\nbefore=%s\nafter=%s", before, after)
	}
}

package attention_engine_test

import (
	"context"
	"encoding/json"
	"math"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

// Both producers persist StoredAttentionNode, including dynamics history. Keep
// the older flat-node regression too: it covers the legacy representation.
func TestM4ProductionNodeRetrievalAfterRestart(t *testing.T) {
	for _, producer := range []string{"W2Dynamics", "W3Sleep"} {
		t.Run(producer, func(t *testing.T) {
			ctx := context.Background()
			path := m4Path(t)
			db, _, service := openM4(t, path)
			defer func() { db.Close() }()
			human := app.Principal{ID: "human", Kind: "human"}
			for _, fixture := range []struct {
				id, topic string
				at        int64
			}{{"z-exact", "agents", 2000}, {"a-neighbor", "systems", 4000}} {
				source := seedM4Source(t, db, "source-"+fixture.id, "p1")
				if _, err := service.RewriteFluid(ctx, app.KnowledgeFluidUpdate{ID: fixture.id, OwnerProject: "p1", Topic: fixture.topic, Fluid: "research", Sources: []domain.EventRef{source}, TS: fixture.at}); err != nil {
					t.Fatal(err)
				}
				if _, err := service.ConfirmKnowledge(ctx, human, fixture.id, app.ConfirmKnowledgeCommand{CommandMeta: m4Meta(1, "confirm-"+fixture.id), OwnerProject: "p1"}); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := service.ReferenceKnowledge(ctx, human, "z-exact", app.ReferenceKnowledgeCommand{CommandMeta: m4Meta(2, "reference"), ProjectID: "p1"}); err != nil {
				t.Fatal(err)
			}
			save := func(tx app.DynamicsTx) error {
				for _, node := range []domain.AttentionNode{
					{ID: "agents", Topic: "agents", Attention: .6, Momentum: .1, Confidence: .8, LastActive: 1000, RelatedTopics: map[string]float64{}},
					{ID: "systems", Topic: "systems", Attention: .9, Momentum: .05, Confidence: .7, LastActive: 1000, RelatedTopics: map[string]float64{"agents": .25}},
				} {
					if err := tx.SaveNode(app.StoredAttentionNode{Node: node, UpdatedAt: 5000, Lifecycle: "active", History: json.RawMessage(`{"max_attention":0.95,"related_at":{"agents":1000}}`), Version: 7}); err != nil {
						return err
					}
				}
				return nil
			}
			var err error
			if producer == "W2Dynamics" {
				err = sqlite.NewDynamicsRepository(db).WithDynamicsTx(ctx, save)
			} else {
				err = sqlite.NewSleepRepository(db).WithSleepTx(ctx, func(tx app.SleepTx) error { return save(tx) })
			}
			if err != nil {
				t.Fatal(err)
			}
			if err = db.Close(); err != nil {
				t.Fatal(err)
			}
			db, _, service = openM4(t, path)
			snapshot := func() string {
				t.Helper()
				var data string
				if err := db.Conn().QueryRow(`SELECT group_concat(data || ':' || attention || ':' || momentum || ':' || confidence || ':' || last_active || ':' || updated_at || ':' || lifecycle || ':' || version,'|') FROM (SELECT * FROM attention_engine_nodes ORDER BY id)`).Scan(&data); err != nil {
					t.Fatal(err)
				}
				return data
			}
			before := snapshot()
			if _, err = db.Conn().Exec(`PRAGMA query_only=ON`); err != nil {
				t.Fatal(err)
			}
			runs := []struct {
				Query   app.KnowledgeQuery
				Matches []app.KnowledgeMatch
			}{}
			for _, query := range []struct {
				topic  string
				lambda float64
				winner string
			}{{"agents", 0, "z-exact"}, {"agents", 1.5, "a-neighbor"}, {"", 0, "z-exact"}} {
				q := app.KnowledgeQuery{Topic: query.topic, Now: 10000, LambdaR: &query.lambda, AgeBasis: app.KnowledgeAgeCreated, ReferenceBoost: 2}
				matches, err := service.RetrieveKnowledge(ctx, human, "p1", q)
				if err != nil {
					t.Fatal(err)
				}
				for _, match := range matches {
					want := domain.RetrievalCandidate{ID: "z-exact", Relevance: 1, Attention: .6, AccessBoost: 3, AgeSeconds: 8}
					if match.Knowledge.Entry.ID == "a-neighbor" {
						want = domain.RetrievalCandidate{ID: "a-neighbor", Relevance: .25, Attention: .9, AccessBoost: 1, AgeSeconds: 6}
						if query.topic == "" {
							want.Relevance = 1
						}
					}
					if match.Inputs != want {
						t.Fatalf("production node inputs: got=%+v want=%+v", match.Inputs, want)
					}
					manual := want.Relevance * math.Exp(-query.lambda*want.AgeSeconds) * want.Attention * want.AccessBoost
					if match.Score <= 0 || math.Abs(match.Score-manual) > 1e-12 {
						t.Fatalf("score=%g want=%g for %s", match.Score, manual, want.ID)
					}
				}
				// The zero-decay winner sorts after its neighbor by ID: an all-zero
				// alphabetical order cannot pass this oracle.
				if len(matches) != 2 || matches[0].Knowledge.Entry.ID != query.winner {
					t.Fatalf("topic=%q lambda=%g rank=%+v want first=%s", query.topic, query.lambda, matches, query.winner)
				}
				runs = append(runs, struct {
					Query   app.KnowledgeQuery
					Matches []app.KnowledgeMatch
				}{q, matches})
			}
			after := snapshot()
			if before != after {
				t.Fatalf("retrieval mutated attention/history: before=%s after=%s", before, after)
			}
			m4Evidence(t, path, struct {
				Producer, AttentionBefore, AttentionAfter string
				Runs                                      any
			}{producer, before, after, runs})
		})
	}
}

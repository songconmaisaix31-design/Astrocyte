package attention_engine_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

// These source receipts are contract_local fixtures; storage, restart and
// governance use the real SQLite adapter. W1 supplies the production EventReader.
type experienceSQLReader struct{ db *sqlite.DB }

func (r experienceSQLReader) GetEvent(ctx context.Context, id string) (app.StoredEvent, error) {
	var record app.StoredEvent
	var data, observation, mapping string
	err := r.db.Conn().QueryRowContext(ctx, `SELECT id,origin_key,received_at,data,observation,mapping_status,raw_bytes
		FROM attention_engine_events WHERE id=?`, id).
		Scan(&record.ID, &record.OriginKey, &record.ReceivedAt, &data, &observation, &mapping, &record.RawBytes)
	if err != nil {
		return record, err
	}
	if err = json.Unmarshal([]byte(observation), &record.Observation); err != nil {
		return record, err
	}
	if mapping == "mapped" {
		record.Event = &domain.Event{}
		err = json.Unmarshal([]byte(data), record.Event)
	}
	return record, err
}

func (r experienceSQLReader) ListEvents(ctx context.Context, from, to int64) ([]app.StoredEvent, error) {
	rows, err := r.db.Conn().QueryContext(ctx, `SELECT id FROM attention_engine_events WHERE ts>=? AND ts<=? ORDER BY id`, from, to)
	if err != nil {
		return nil, err
	}
	ids := []string{}
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			rows.Close()
			return nil, err
		}
		ids = append(ids, id)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	result := []app.StoredEvent{}
	for _, id := range ids {
		event, err := r.GetEvent(ctx, id)
		if err != nil {
			return nil, err
		}
		result = append(result, event)
	}
	return result, nil
}

func experienceDB(t *testing.T, path string) *sqlite.DB {
	t.Helper()
	db, err := sqlite.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = db.RunMigrations(context.Background(), migrations.FS); err != nil {
		db.Close()
		t.Fatal(err)
	}
	return db
}

func experienceService(db *sqlite.DB, project string) *app.ExperienceService {
	return app.NewExperienceService(sqlite.NewExperienceRepository(db), experienceSQLReader{db}, app.ExperienceOptions{
		ProjectID: project, Clock: func() time.Time { return time.UnixMilli(1800000001000) },
	})
}

func experienceReceipt(t *testing.T, db *sqlite.DB, id, project, kind, sourceKind, outcome, mapping string) []byte {
	t.Helper()
	raw := []byte(" \n{\"outcome\":\"" + outcome + "\",\"private_text\":\"centralized scheduler secret\",\"unknown\":1}\r\n")
	event := domain.Event{Source: "codex", Session: "contract-local", TS: 1800000000000, Kind: kind, Scope: project, Topic: "task", Raw: raw}
	data, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}
	observation, err := json.Marshal(app.Observation{Source: "codex", SourceKind: sourceKind, Session: event.Session, Scope: project,
		TS: event.TS, Outcome: outcome, Provenance: "contract_local_fixture"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Conn().ExecContext(context.Background(), `INSERT INTO attention_engine_events
		(id,source,ts,kind,topic,data,origin_key,received_at,raw_bytes,observation,mapping_status)
		VALUES(?,?,?,?,?,?,?,?,?,?,?)`, id, event.Source, event.TS, event.Kind, event.Topic, string(data), "fixture:"+id,
		1800000000001, raw, string(observation), mapping)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func experienceError(t *testing.T, err error, code apierrors.Code) {
	t.Helper()
	var serviceErr *apierrors.ServiceError
	if !errors.As(err, &serviceErr) || serviceErr.Code != code {
		t.Fatalf("wanted %s, got %v", code, err)
	}
}

func experienceMeta(request string, version int) app.CommandMeta {
	return app.CommandMeta{SchemaVersion: 1, RequestID: request, ExpectedVersion: version, IdempotencyKey: request}
}

var experienceHuman = app.Principal{ID: "local-human", Kind: "human"}
var experienceAgent = app.Principal{ID: "bound-agent", Kind: "agent"}

func TestM5ExperienceCaptureSQLiteRestartAndConcurrentIdempotence(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "experience.sqlite")
	db := experienceDB(t, path)
	defer func() { db.Close() }()
	raw := experienceReceipt(t, db, "failed-task", "A", "agent_task", "agent_task", "failed", "mapped")
	svc := experienceService(db, "A")
	first, err := svc.CaptureFailure(ctx, "failed-task")
	if err != nil || first == nil {
		t.Fatalf("capture: %v %v", first, err)
	}
	if first.Experience.Status != "quarantine" || first.Experience.Pattern != "agent_task" || first.Experience.Verdict != "condition" || first.Experience.Condition != "outcome=failed" {
		t.Fatalf("observed candidate: %+v", first)
	}
	if first.OwnerProject != "A" || first.OriginEventID != "failed-task" || len(first.Experience.Evidence) != 1 || first.Experience.Evidence[0].TS != 1800000000000 {
		t.Fatalf("trace/scope: %+v", first)
	}
	other, err := sqlite.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer other.Close()
	services := []*app.ExperienceService{svc, experienceService(other, "A")}
	var wg sync.WaitGroup
	for _, service := range services {
		wg.Add(1)
		go func(service *app.ExperienceService) {
			defer wg.Done()
			got, err := service.CaptureFailure(ctx, "failed-task")
			if err != nil || got == nil || got.Experience.ID != first.Experience.ID || got.Version != 1 {
				t.Errorf("duplicate: %v %v", got, err)
			}
		}(service)
	}
	wg.Wait()
	if _, err = svc.ApproveExperience(ctx, experienceHuman, first.Experience.ID, app.ApproveExperienceCommand{CommandMeta: experienceMeta("approve", 1), OwnerProject: "A"}); err != nil {
		t.Fatal(err)
	}
	if err = other.Close(); err != nil {
		t.Fatal(err)
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	db = experienceDB(t, path)
	svc = experienceService(db, "A")
	again, err := svc.CaptureFailure(ctx, "failed-task")
	if err != nil || again == nil || again.Experience.ID != first.Experience.ID || again.Experience.Status != "approved" || again.Version != 2 || again.CreatedAt != first.CreatedAt {
		t.Fatalf("restart replay reset approved: %v %v", again, err)
	}
	evidence, err := svc.ResolveExperienceEvidence(ctx, experienceAgent, first.Experience.ID, "A")
	if err != nil || len(evidence) != 1 || evidence[0].ID != "failed-task" || !bytes.Equal(evidence[0].RawBytes, raw) {
		t.Fatalf("original persisted evidence: %v %v", evidence, err)
	}
	var count int
	var data string
	if err = db.Conn().QueryRowContext(ctx, `SELECT count(*),data FROM attention_engine_experiences`).Scan(&count, &data); err != nil || count != 1 || bytes.Contains([]byte(data), []byte("secret")) {
		t.Fatalf("duplicate or private conversation copied: %d %s %v", count, data, err)
	}
}

func TestM5ExperienceRejectsImportDistillationPendingAndUnknown(t *testing.T) {
	ctx := context.Background()
	db := experienceDB(t, filepath.Join(t.TempDir(), "experience.sqlite"))
	defer db.Close()
	svc := experienceService(db, "A")
	for _, fixture := range []struct{ id, kind, sourceKind, outcome, mapping string }{
		{"import", "agent_task", "import", "failed", "mapped"},
		{"distillation", "agent_task", "distillation", "failed", "mapped"},
		{"job", "agent_task", "attention.job_failed", "failed", "mapped"},
		{"success", "agent_task", "agent_task", "succeeded", "mapped"},
		{"unknown", "agent_task", "agent_task", "unknown", "mapped"},
		{"pending", "agent_task", "agent_task", "failed", "pending"},
		{"search", "search", "search", "failed", "mapped"},
	} {
		experienceReceipt(t, db, fixture.id, "A", fixture.kind, fixture.sourceKind, fixture.outcome, fixture.mapping)
		candidate, err := svc.CaptureFailure(ctx, fixture.id)
		if err != nil || candidate != nil {
			t.Fatalf("%s captured: %v %v", fixture.id, candidate, err)
		}
	}
	experienceReceipt(t, db, "other-project", "B", "agent_task", "agent_task", "failed", "mapped")
	_, err := svc.CaptureFailure(ctx, "other-project")
	experienceError(t, err, apierrors.ScopeDenied)
	_, err = experienceService(db, "").CaptureFailure(ctx, "other-project")
	experienceError(t, err, apierrors.ScopeDenied)
	experienceReceipt(t, db, "eligible", "A", "agent_task", "agent_task", "failed", "mapped")
	candidates, err := svc.CaptureFailures(ctx, 0, 1900000000000)
	if err != nil || len(candidates) != 1 || candidates[0].OriginEventID != "eligible" {
		t.Fatalf("catch-up: %v %v", candidates, err)
	}
}

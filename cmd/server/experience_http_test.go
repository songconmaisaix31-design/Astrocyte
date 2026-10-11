package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/httpapi"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

func TestExperienceHTTPActualFactoryReplayAndRestart(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "state.sqlite")
	db, err := sqlite.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { db.Close() }()
	if err = db.RunMigrations(ctx, migrations.FS); err != nil {
		t.Fatal(err)
	}
	events := sqlite.NewEventRepository(db)
	service := app.NewExperienceService(sqlite.NewExperienceRepository(db), events, app.ExperienceOptions{ProjectID: "project-a"})
	rows, err := app.NewEventService(events, .1, service.CaptureCommittedFailure).Record(ctx, []app.StoredEvent{{OriginKey: "contract-local:http-failure", RawBytes: []byte(` { "outcome": "failed" } `), Observation: app.Observation{Source: "codex", SourceKind: "agent_task", Scope: "project-a", TS: 100, Outcome: "failed", Depth: "unknown", Provenance: "contract_local"}, Event: &domain.Event{Source: "codex", Kind: "agent_task", Scope: "project-a", TS: 100, Raw: json.RawMessage(`{"outcome":"failed"}`)}}})
	if err != nil {
		t.Fatal(err)
	}
	entry, err := service.CaptureFailure(ctx, rows[0].ID)
	if err != nil || entry == nil {
		t.Fatalf("capture: %+v %v", entry, err)
	}
	var server *httptest.Server
	var client *http.Client
	var csrf string
	start := func() {
		bridge := &experienceHTTPBridge{db: db, events: sqlite.NewEventRepository(db), projects: &experienceProjects{}}
		api := httpapi.NewServer(httpapi.Config{AgentToken: "fixture-agent", Services: httpapi.Services{ExperienceGovernance: bridge}, Logger: slog.New(slog.NewTextHandler(io.Discard, nil))})
		server = httptest.NewServer(api.Handler())
		jar, _ := cookiejar.New(nil)
		client = &http.Client{Jar: jar}
		response, err := client.Get(server.URL + "/api/v1/auth/session")
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		var session struct {
			CSRF string `json:"csrf_token"`
		}
		if err = json.NewDecoder(response.Body).Decode(&session); err != nil || response.StatusCode != 200 || session.CSRF == "" {
			t.Fatalf("session: %d %v", response.StatusCode, err)
		}
		csrf = session.CSRF
	}
	start()
	defer func() { server.Close() }()
	request := func(endpoint string, body any, key, trace, token, bearer string, want int) []byte {
		t.Helper()
		data, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", server.URL+endpoint, bytes.NewReader(data))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", key)
		req.Header.Set("X-Request-ID", trace)
		req.Header.Set("X-CSRF-Token", token)
		if bearer != "" {
			req.Header.Set("Authorization", "Bearer "+bearer)
		}
		response, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		payload, _ := io.ReadAll(response.Body)
		if response.StatusCode != want {
			t.Fatalf("%s = %d want %d: %s", endpoint, response.StatusCode, want, payload)
		}
		return payload
	}
	meta := func(id string, version int) app.CommandMeta {
		return app.CommandMeta{SchemaVersion: 1, RequestID: id, ExpectedVersion: version}
	}
	approve := app.ApproveExperienceCommand{CommandMeta: meta("business-approve", 1), OwnerProject: "project-a"}
	approvePath := "/api/v1/experiences/" + entry.Experience.ID + "/approve"
	request(approvePath, approve, "approval-key", "missing-csrf", "", "", 403)
	request(approvePath, approve, "approval-key", "agent", csrf, "fixture-agent", 403)
	first := request(approvePath, approve, "approval-key", "trace-one", csrf, "", 200)
	var result app.ExperienceResult
	if err = json.Unmarshal(first, &result); err != nil || result.Version != 2 || result.Experience.Status != domain.ExperienceStatusApproved {
		t.Fatalf("approval: %+v %v", result, err)
	}
	// Request trace is independent of business idempotency and CAS version.
	approve.RequestID = "business-retry"
	if replay := request(approvePath, approve, "approval-key", "trace-two", csrf, "", 200); !bytes.Equal(replay, first) {
		t.Fatalf("changed replay: %s / %s", first, replay)
	}
	changed := approve
	changed.ExpectedVersion = 2
	request(approvePath, changed, "approval-key", "trace-changed", csrf, "", 409)
	foreign := approve
	foreign.OwnerProject = "project-b"
	request(approvePath, foreign, "approval-key", "trace-foreign", csrf, "", 403)
	share := app.ShareMemoryCommand{CommandMeta: meta("share-one", 2), OwnerProject: "project-a", TargetProject: "project-b"}
	sharePath := "/api/v1/experiences/" + entry.Experience.ID + "/shares"
	shared := request(sharePath, share, "share-key", "trace-share", csrf, "", 200)
	share.RequestID = "share-retry"
	if replay := request(sharePath, share, "share-key", "trace-share-retry", csrf, "", 200); !bytes.Equal(shared, replay) {
		t.Fatal("share replay changed")
	}
	server.Close()
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = sqlite.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	start()
	if replay := request(approvePath, approve, "approval-key", "trace-restart", csrf, "", 200); !bytes.Equal(first, replay) {
		t.Fatal("restart approval replay changed")
	}
	if replay := request(sharePath, share, "share-key", "trace-share-restart", csrf, "", 200); !bytes.Equal(shared, replay) {
		t.Fatal("restart share replay changed")
	}
	visible, err := (&experienceHTTPBridge{db: db, events: sqlite.NewEventRepository(db), projects: &experienceProjects{}}).GetExperience(ctx, app.Principal{ID: "local-human", Kind: "human"}, entry.Experience.ID, "project-b")
	if err != nil || visible.Version != 3 || len(visible.Shares) != 1 || len(visible.Actions) != 2 || visible.Actions[0].Action != "approve" || visible.Actions[1].Action != "share" {
		t.Fatalf("audited restart: %+v %v", visible, err)
	}
}

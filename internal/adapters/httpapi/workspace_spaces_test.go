package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/projectspacebridge"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

func httpAttentionSpaces(db *sqlite.DB) projectspacebridge.Attention {
	return projectspacebridge.Attention{Spaces: workspace.NewProjectSpaceService(sqlite.NewProjectSpaceRepository(db), workspace.ProjectSpaceOptions{})}
}

func TestWorkspaceSpacesHTTPWithoutAttention(t *testing.T) {
	db, err := sqlite.Open(filepath.Join(t.TempDir(), "state.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
	if err = db.RunMigrations(context.Background(), migrations.FS); err != nil {
		t.Fatal(err)
	}
	server := NewServer(Config{Services: Services{ProjectSpaces: httpAttentionSpaces(db).Spaces}})
	cookie, csrf := bootstrapSession(t, server)
	body := `{"schema_version":1,"request_id":"workspace-space-http","expected_version":1,"title":"Independent SWA development space"}`
	unauthorized := attentionRequest(server, http.MethodPost, "/api/v1/project-spaces", body, cookie, "", "", "")
	if unauthorized.Code != http.StatusForbidden {
		t.Fatal("space creation bypassed CSRF", unauthorized.Code)
	}
	created := attentionRequest(server, http.MethodPost, "/api/v1/project-spaces", body, cookie, csrf, "", "")
	var result workspace.ProjectSpaceResult
	if created.Code != http.StatusCreated || json.Unmarshal(created.Body.Bytes(), &result) != nil || result.Space.ID == "" {
		t.Fatal("independent Workspace HTTP create failed", created.Code, created.Body.String())
	}
	for _, route := range []string{"/api/v1/project-spaces", "/api/v1/project-spaces/" + result.Space.ID} {
		response := attentionRequest(server, http.MethodGet, route, "", cookie, "", "", "")
		if response.Code != http.StatusOK {
			t.Fatal("independent Workspace HTTP read failed", route, response.Code)
		}
	}
	for _, suffix := range []string{"/references", "/references/remove"} {
		response := attentionRequest(server, http.MethodPost, "/api/v1/project-spaces/"+result.Space.ID+suffix, body, cookie, csrf, "", "")
		if response.Code != http.StatusNotImplemented {
			t.Fatal("missing material integration was not explicit", suffix, response.Code)
		}
	}
}

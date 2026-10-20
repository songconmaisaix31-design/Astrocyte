package app_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestOrcaCLIRealCurrentProjectAuthorization(t *testing.T) {
	if os.Getenv("SWA_REAL_ENABLED") != "1" {
		t.Skip("真实链路须在已授窗口和同安装候选运行，当前不计验收")
	}
	endpoint, project, token := os.Getenv("SWA_ENDPOINT"), os.Getenv("SWA_PROJECT_ID"), os.Getenv("SWA_PROJECT_TOKEN")
	data := os.Getenv("SWA_REAL_DATA_ROOT")
	if endpoint == "" || project == "" || token == "" || !filepath.IsAbs(data) {
		t.Fatal("实际HTTP、项目批准或私有根缺失")
	}
	client := &http.Client{Timeout: 30 * time.Second, CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }}
	request, err := http.NewRequest(http.MethodPost, endpoint+"/api/v1/local-projects/"+url.PathEscape(project)+"/orca/commands", bytes.NewBufferString(`{"action":"status"}`))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")
	response, err := client.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	var reply struct {
		OK        bool   `json:"ok"`
		RuntimeID string `json:"runtime_id"`
		Result    struct {
			Databases []struct {
				Name string `json:"name"`
				File string `json:"file"`
			} `json:"databases"`
		} `json:"result"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 2*1024*1024)).Decode(&reply); err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusOK || !reply.OK || reply.RuntimeID != os.Getenv("SWA_REAL_RUNTIME_ID") {
		t.Fatal("真实当前项目批准没有取得同一私有 runtime 结果")
	}
	found := false
	for _, database := range reply.Result.Databases {
		if database.Name != "main" {
			continue
		}
		rel, err := filepath.Rel(data, database.File)
		if err != nil || !filepath.IsAbs(database.File) || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			t.Fatal("实际 SQLite 不属于本项私有数据根")
		}
		if _, err := os.Stat(database.File); err != nil {
			t.Fatal(err)
		}
		found = true
	}
	if !found {
		t.Fatal("实际 PRAGMA database_list 缺少主库")
	}
}

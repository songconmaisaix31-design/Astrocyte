package s5

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	adapter "github.com/songconmaisaix31-design/Astrocyte/internal/adapters/judgment"
	"github.com/songconmaisaix31-design/Astrocyte/internal/judgment"
)

func s5Root(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func liveEnabled(t *testing.T) {
	t.Helper()
	if os.Getenv("ASTROCYTE_S5_LIVE") != "1" {
		t.Skip("actual public S5 model assets are required; enable ASTROCYTE_S5_LIVE=1")
	}
}

func actualLayaConfig(t *testing.T, checkpoint, language string, round int) adapter.LayaConfig {
	t.Helper()
	root := s5Root(t)
	assetsPath := filepath.Join(root, ".verification", "s5", "laya-"+checkpoint+"-assets.json")
	data, err := os.ReadFile(assetsPath)
	if err != nil {
		t.Fatal(err)
	}
	var assets struct {
		ModelRevision  string            `json:"model_revision"`
		SourceRevision string            `json:"source_revision"`
		Files          map[string]string `json:"files"`
	}
	if err := json.Unmarshal(data, &assets); err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(assets.Files)
	if err != nil {
		t.Fatal(err)
	}
	revision := assets.ModelRevision + "/" + checkpoint + "/" + assets.SourceRevision + "/" + judgment.DigestText(string(encoded))
	assetRoot := os.Getenv("ASTROCYTE_S5_ASSET_ROOT")
	if !filepath.IsAbs(assetRoot) {
		t.Fatal("ASTROCYTE_S5_ASSET_ROOT must be an explicit absolute owned asset directory")
	}
	modelDir := filepath.Join(assetRoot, checkpoint)
	node, err := exec.LookPath("node")
	if err != nil {
		t.Fatal(err)
	}
	stderr, err := os.CreateTemp(filepath.Join(root, ".verification", "s5"), "go-laya-"+checkpoint+"-"+language+"-*.log")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := stderr.Close(); err != nil {
			t.Error(err)
		}
	})
	return adapter.LayaConfig{
		Node: node, Worker: filepath.Join(root, "internal", "adapters", "judgment", "laya-worker.mjs"),
		ModelDir: modelDir, AssetsPath: assetsPath,
		SDKEntry: filepath.Join(root, ".verification", "s5", "upstream", "laya", "laya-ts", "dist", "index.js"),
		RPCEntry: filepath.Join(root, ".verification", "s5", "ipc", "node_modules", "vscode-jsonrpc", "lib", "node", "main.js"),
		Language: language, ModelRevision: revision, Stderr: stderr, StartupTimeoutMS: 60000, ShutdownTimeoutMS: 10000,
	}
}

func TestLayaActualCPUThreeTypesAndRestart(t *testing.T) {
	liveEnabled(t)
	root := s5Root(t)
	english, err := os.ReadFile(filepath.Join(root, ".verification", "s5", "upstream", "laya", "laya-ts", "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	chinese, err := os.ReadFile(filepath.Join(root, ".verification", "s5", "upstream", "astrocyte-readme.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, route := range []struct{ checkpoint, language string }{{"english", "en"}, {"multilingual", "en"}, {"multilingual", "zh"}} {
		t.Run(route.checkpoint+"/"+route.language, func(t *testing.T) {
			var oldPID int
			for round := 1; round <= 2; round++ {
				ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
				config := actualLayaConfig(t, route.checkpoint, route.language, round)
				model, err := adapter.NewLaya(ctx, config)
				if err != nil {
					cancel()
					t.Fatal(err)
				}
				t.Cleanup(func() {
					if err := model.Close(); err != nil {
						t.Error(err)
					}
					cancel()
				})
				health, err := model.Health(ctx)
				if err != nil {
					t.Fatal(err)
				}
				if health.PID == oldPID {
					t.Fatal("explicit restart reused the old process")
				}
				oldPID = health.PID
				state := strings.Split(string(english), "## Node")[0]
				if route.language == "zh" {
					state = string([]rune(string(chinese))[:200])
				}
				questions := []judgment.Question{
					{ID: "runtime", Type: judgment.Choice, Instructions: "Which runtime does the document describe?", Options: []judgment.Option{{ID: "agent", Description: "A typed Agent decision engine"}, {ID: "music", Description: "A music player"}}},
					{ID: "truth", Type: judgment.Truth, Instructions: "Does the document describe an Agent or inference runtime?"},
					{ID: "detail", Type: judgment.Score, Instructions: "How much detail is given about inference or Agent workflow?", Levels: []string{"No description", "Name only", "A concrete workflow is described"}},
				}
				for _, question := range questions {
					input := judgment.BackendInput{Question: question, State: state, Language: route.language, ModelRevision: config.ModelRevision, Budget: judgment.Budget{TimeoutMS: 60000, MaxInputBytes: 131072, MaxInputTokens: 2048, MaxOutputTokens: 0}}
					out, err := model.Evaluate(ctx, input)
					if err != nil {
						t.Fatal(err)
					}
					if out.Truncated || out.ModelRevision != config.ModelRevision || out.ProviderModel != "laya-rl-agent" || out.Usage.InputTokens <= 0 {
						t.Fatalf("actual model/usage or short-input truncation differs: %+v", out)
					}
					input.State = string(english)
					long, err := model.Evaluate(ctx, input)
					if err != nil || !long.Truncated {
						t.Fatal("actual full official source did not report truncation", err)
					}
					input.Language = "unsupported"
					if _, err := model.Evaluate(ctx, input); err == nil {
						t.Fatal("unsupported language silently selected a backend")
					}
				}
				after, err := model.Health(ctx)
				if err != nil || after.PID != health.PID {
					t.Fatal("single resident model was not retained", err)
				}
				if err := model.Close(); err != nil {
					t.Fatal(err)
				}
				cancel()
			}
		})
	}
}

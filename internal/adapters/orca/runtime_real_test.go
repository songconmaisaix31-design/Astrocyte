package orca

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
)

func TestOrcaCLIRealPrivateRuntimeScope(t *testing.T) {
	if os.Getenv("SWA_REAL_ENABLED") != "1" {
		t.Skip("真实链路须在已授窗口和同安装候选运行，当前不计验收")
	}
	root, data := os.Getenv("SWA_REAL_INSTALL_ROOT"), os.Getenv("SWA_REAL_DATA_ROOT")
	controller, project := os.Getenv("SWA_REAL_CONTROLLER"), os.Getenv("SWA_PROJECT_ID")
	if !filepath.IsAbs(root) || !filepath.IsAbs(data) || controller == "" || project == "" {
		t.Fatal("实际私有安装、项目或控制端缺失")
	}
	config := Config{InstallRoot: root, DataRoot: data, NodeExecutable: filepath.Join(root, "node", "node.exe"), ClientEntry: filepath.Join(root, "runtime", "client.cjs"), PiEntry: filepath.Join(root, "runtime", "pi-entry.mjs"), CodexAuthFile: os.Getenv("SWA_REAL_CODEX_AUTH_FILE"), ProtectedRoots: filepath.SplitList(os.Getenv("SWA_REAL_PROTECTED_ROOTS"))}
	runtime, err := New(config)
	if err != nil {
		t.Fatal(err)
	}
	if runtime.Target() != "pi" {
		t.Fatal("执行授权目标必须是实际已注册 pi")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	scope, err := runtime.ResolveScope(ctx, project, app.OrcaCommand{Action: "controller.show", Controller: controller})
	if err != nil {
		t.Fatal(err)
	}
	if scope.ProjectID != project || scope.Controller != controller || scope.AuthorityDomain == "" {
		t.Fatal("真实原生记录未证明当前项目和控制端")
	}
	config.ProtectedRoots = append(config.ProtectedRoots, data)
	if _, err := New(config); err == nil {
		t.Fatal("实际私有数据根与保护目录重合时必须拒绝")
	}
}

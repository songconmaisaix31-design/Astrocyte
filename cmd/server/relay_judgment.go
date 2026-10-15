package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/admissionbridge"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/httpapi"
	judgmentadapter "github.com/songconmaisaix31-design/Astrocyte/internal/adapters/judgment"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/judgment"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
)

type relayJudgmentConfig struct {
	jevKey    string
	laya      *judgmentadapter.LayaConfig
	stderrDir string
}

// 这里只解析可信启动配置；每次请求仍须明确选择已授权后端、语言和预算。
func resolveRelayJudgmentConfig() (relayJudgmentConfig, error) {
	config := relayJudgmentConfig{jevKey: os.Getenv("TYPESAFE_API_KEY")}
	laya := judgmentadapter.LayaConfig{
		Node: os.Getenv("ASTROCYTE_LAYA_NODE"), Worker: os.Getenv("ASTROCYTE_LAYA_WORKER"),
		ModelDir: os.Getenv("ASTROCYTE_LAYA_MODEL_DIR"), AssetsPath: os.Getenv("ASTROCYTE_LAYA_ASSETS"),
		SDKEntry: os.Getenv("ASTROCYTE_LAYA_SDK_ENTRY"), RPCEntry: os.Getenv("ASTROCYTE_LAYA_RPC_ENTRY"),
		Language:    os.Getenv("ASTROCYTE_LAYA_LANGUAGE"),
		Environment: []string{"NODE_OPTIONS=--max-old-space-size=512"},
	}
	config.stderrDir = os.Getenv("ASTROCYTE_LAYA_STDERR_DIR")
	values := []string{laya.Node, laya.Worker, laya.ModelDir, laya.AssetsPath, laya.SDKEntry, laya.RPCEntry, laya.Language, config.stderrDir, os.Getenv("ASTROCYTE_LAYA_STARTUP_TIMEOUT_MS"), os.Getenv("ASTROCYTE_LAYA_SHUTDOWN_TIMEOUT_MS")}
	configured := false
	for _, value := range values {
		configured = configured || value != ""
	}
	if !configured {
		return config, nil
	}
	for _, value := range values {
		if value == "" {
			return config, errors.New("Laya requires all explicit runtime paths, language, private stderr directory and lifecycle budgets")
		}
	}
	for _, path := range []string{laya.Node, laya.Worker, laya.ModelDir, laya.AssetsPath, laya.SDKEntry, laya.RPCEntry, config.stderrDir} {
		if !filepath.IsAbs(path) {
			return config, errors.New("Laya runtime paths must be explicit absolute paths")
		}
	}
	var err error
	laya.StartupTimeoutMS, err = strconv.ParseInt(values[8], 10, 64)
	if err != nil || laya.StartupTimeoutMS < 1 || laya.StartupTimeoutMS > int64((1<<63-1)/time.Millisecond) {
		return config, errors.New("Laya startup timeout must be an explicit positive millisecond budget")
	}
	laya.ShutdownTimeoutMS, err = strconv.ParseInt(values[9], 10, 64)
	if err != nil || laya.ShutdownTimeoutMS < 1 || laya.ShutdownTimeoutMS > int64((1<<63-1)/time.Millisecond) {
		return config, errors.New("Laya shutdown timeout must be an explicit positive millisecond budget")
	}
	laya.ModelRevision, err = judgmentadapter.LayaRevision(laya.AssetsPath)
	if err != nil {
		return config, errors.New("Laya fixed asset descriptor could not be verified")
	}
	config.laya = &laya
	return config, nil
}

// 判断、委托、品味和准入复用原主数据库、S2、资料服务和同一个快照 provider。
func assembleRelayJudgment(ctx context.Context, services *httpapi.Services, db *sqlite.DB, references workspace.ProjectReferences, snapshots workspace.ProjectSnapshotProvider, core *attention.Service, config relayJudgmentConfig) (func() error, error) {
	if services == nil || db == nil || core == nil || services.WorkspaceS2 == nil || references == nil {
		return nil, errors.New("judgment assembly requires the original database, Workspace and Attention services")
	}
	services.Judgment = &httpapi.JudgmentServices{Taste: core}
	if snapshots == nil {
		if config.laya != nil || config.jevKey != "" {
			return nil, errors.New("configured judgment models require the actual repository snapshot provider")
		}
		return func() error { return nil }, nil
	}
	models := []judgment.Model{}
	var laya *judgmentadapter.Laya
	var stderr *os.File
	var closeOnce sync.Once
	var closeErr error
	closeModels := func() error {
		closeOnce.Do(func() {
			if laya != nil {
				closeErr = laya.Close()
			}
			if stderr != nil {
				closeErr = errors.Join(closeErr, stderr.Close())
			}
		})
		return closeErr
	}
	if config.jevKey != "" {
		model, err := judgmentadapter.NewJev(config.jevKey)
		if err != nil {
			return nil, err
		}
		models = append(models, model)
	}
	if config.laya != nil {
		if err := os.MkdirAll(config.stderrDir, 0700); err != nil {
			return nil, errors.New("Laya private stderr directory is unavailable")
		}
		var err error
		stderr, err = os.CreateTemp(config.stderrDir, "laya-*.stderr.log")
		if err != nil {
			return nil, errors.New("Laya private stderr file is unavailable")
		}
		configured := *config.laya
		configured.Stderr = stderr
		laya, err = judgmentadapter.NewLaya(ctx, configured)
		if err != nil {
			_, writeErr := fmt.Fprintln(stderr, "startup_error:", err)
			return nil, errors.Join(errors.New("configured Laya startup failed; inspect its private stderr log"), writeErr, closeModels())
		}
		models = append(models, laya)
	}
	access := &judgmentadapter.ProjectAuthorization{Projects: db, References: references, Policies: db, Snapshots: snapshots, Goals: services.WorkspaceS2, Ranking: core, Rules: db, Taste: core}
	judgments, err := judgment.NewService(db, access, models...)
	if err != nil {
		return nil, errors.Join(err, closeModels())
	}
	delegation, err := judgment.NewDelegationService(db, judgments, access, judgmentadapter.AttentionOpportunities{Attention: core})
	if err != nil {
		return nil, errors.Join(err, closeModels())
	}
	admissions, err := admissionbridge.Connect(core, sqlite.NewAdmissionRepository(db), services.WorkspaceS2, attention.AdmissionOptions{Projects: judgmentadapter.AdmissionProjects{Authorization: access, Rules: db}, Judgment: judgmentadapter.AdmissionAuthorizer{Delegation: delegation}})
	if err != nil {
		return nil, errors.Join(err, closeModels())
	}
	services.Judgment = &httpapi.JudgmentServices{Judgments: judgments, Policies: access, Delegation: delegation, Admissions: judgmentadapter.AdmissionWorkflow{Delegation: delegation, Admissions: admissions}, Taste: core}
	return closeModels, nil
}

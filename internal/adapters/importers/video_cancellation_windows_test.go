//go:build video_live

package importers_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/httpapi"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/importers"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/objects"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
	"golang.org/x/sys/windows"
)

type videoCancellationEvent struct {
	importers.MediaProcessEvent
	URL string `json:"url,omitempty"`
}

// 子进程只组装真实应用、HTTP、SQLite 与 Reader，不替代任一外部能力。
func videoCancellationServiceChild(t *testing.T) {
	root := videoRequiredEnv(t, "ASTROCYTE_TEST_EXPORT_DIR")
	toolRoot := filepath.Join(root, "tool-temp")
	if err := os.MkdirAll(toolRoot, 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("TEMP", toolRoot)
	t.Setenv("TMP", toolRoot)
	var mutex sync.Mutex
	encoder := json.NewEncoder(os.Stderr)
	emit := func(event videoCancellationEvent) {
		mutex.Lock()
		defer mutex.Unlock()
		if err := encoder.Encode(event); err != nil {
			panic(err)
		}
	}
	extractor, err := importers.NewSummarizeExtractorWithOptions(
		videoRequiredEnv(t, "ASTROCYTE_TEST_SUMMARIZE_NODE"), videoRequiredEnv(t, "ASTROCYTE_TEST_SUMMARIZE_CLI"),
		importers.SummarizeOptions{
			YtDlpPath: videoRequiredEnv(t, "ASTROCYTE_TEST_YT_DLP"), FFmpegPath: videoRequiredEnv(t, "ASTROCYTE_TEST_FFMPEG"),
			WhisperBinary: videoRequiredEnv(t, "ASTROCYTE_TEST_WHISPER"), WhisperModel: videoRequiredEnv(t, "ASTROCYTE_TEST_WHISPER_MODEL"),
			ProxyURL: os.Getenv("ASTROCYTE_TEST_SUMMARIZE_PROXY_URL"), Timeout: 5 * time.Minute,
			ObserveMediaProcess: func(event importers.MediaProcessEvent) { emit(videoCancellationEvent{MediaProcessEvent: event}) },
		})
	if err != nil {
		t.Fatal(err)
	}
	db, err := sqlite.Open(filepath.Join(root, "state.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	}()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := db.RunMigrations(ctx, migrations.FS); err != nil {
		t.Fatal(err)
	}
	store, err := objects.New(filepath.Join(root, "objects"))
	if err != nil {
		t.Fatal(err)
	}
	reader := importers.NewReader(nil)
	reader.Summarize = extractor
	service := app.NewAttentionService(sqlite.NewAttentionRepository(db), reader, store, app.ServiceOptions{WorkerConcurrency: 1, MaxAttempts: 1, JobTimeout: 5 * time.Minute})
	server := httptest.NewServer(httpapi.NewServer(httpapi.Config{Services: httpapi.Services{Attention: service}, Logger: slog.New(slog.NewTextHandler(io.Discard, nil))}).Handler())
	defer server.Close()
	finished := make(chan error, 1)
	go func() { finished <- service.Run(ctx) }()
	emit(videoCancellationEvent{MediaProcessEvent: importers.MediaProcessEvent{Kind: "go-service", Phase: "ready", PID: os.Getpid()}, URL: server.URL})
	if _, err := io.Copy(io.Discard, os.Stdin); err != nil {
		t.Fatal(err)
	}
	cancel()
	if err := <-finished; err != nil {
		t.Fatal("actual cancellation service stopped with error", err)
	}
}

func TestVideoNativeStartupCancellationHTTPProcessRestart(t *testing.T) {
	if os.Getenv("ASTROCYTE_VIDEO_CANCEL_CHILD") == "1" {
		videoCancellationServiceChild(t)
		return
	}
	root := videoRequiredEnv(t, "ASTROCYTE_TEST_EXPORT_DIR")
	if _, err := os.Stat(filepath.Join(root, "state.sqlite")); !os.IsNotExist(err) {
		t.Fatal("preserve earlier cancellation evidence; provide a new directory")
	}
	if err := os.MkdirAll(root, 0700); err != nil {
		t.Fatal(err)
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	trace, err := os.OpenFile(filepath.Join(root, "owned-process-events.jsonl"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer trace.Close()
	traceEncoder := json.NewEncoder(trace)
	var receipt app.ImportJobResult
	var cancelled app.Job
	startedProcesses := 0
	for iteration := 0; iteration < 2; iteration++ {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cancel()
		command := exec.CommandContext(ctx, executable, "-test.run", "^TestVideoNativeStartupCancellationHTTPProcessRestart$")
		command.Env = append(os.Environ(), "ASTROCYTE_VIDEO_CANCEL_CHILD=1")
		stdin, err := command.StdinPipe()
		if err != nil {
			t.Fatal(err)
		}
		stderr, err := command.StderrPipe()
		if err != nil {
			t.Fatal(err)
		}
		stdout, err := os.OpenFile(filepath.Join(root, "go-service-"+string(rune('1'+iteration))+".log"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			t.Fatal(err)
		}
		command.Stdout = stdout
		if err := command.Start(); err != nil {
			t.Fatal(err)
		}
		goProcess, err := windows.OpenProcess(windows.SYNCHRONIZE|windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(command.Process.Pid))
		if err != nil {
			t.Fatal(err)
		}
		defer windows.CloseHandle(goProcess)
		t.Cleanup(func() { stdin.Close(); cancel() })
		events := make(chan videoCancellationEvent, 256)
		decodeError := make(chan error, 1)
		go func() {
			defer close(events)
			decoder := json.NewDecoder(stderr)
			for {
				var event videoCancellationEvent
				err := decoder.Decode(&event)
				if errors.Is(err, io.EOF) {
					decodeError <- nil
					return
				}
				if err != nil {
					decodeError <- err
					return
				}
				events <- event
			}
		}()
		handles := map[int]windows.Handle{}
		observe := func(event videoCancellationEvent) {
			if err := traceEncoder.Encode(event); err != nil {
				t.Fatal(err)
			}
			if event.Phase == "started" {
				startedProcesses++
				if iteration != 0 {
					t.Fatal("restart or replay started an extraction process", event.Kind)
				}
				handle, err := windows.OpenProcess(windows.SYNCHRONIZE|windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(event.PID))
				if err != nil {
					t.Fatal("actual tracked process handle unavailable", event.Kind, event.PID, err)
				}
				handles[event.PID] = handle
				deferCloser := handle
				t.Cleanup(func() { windows.CloseHandle(deferCloser) })
			}
		}
		var ready videoCancellationEvent
		select {
		case ready = <-events:
			observe(ready)
		case <-ctx.Done():
			t.Fatal("actual Go HTTP service did not start", ctx.Err())
		}
		if ready.Kind != "go-service" || ready.Phase != "ready" || ready.PID != command.Process.Pid || ready.URL == "" {
			t.Fatal("actual Go service observation mismatch")
		}
		jar, err := cookiejar.New(nil)
		if err != nil {
			t.Fatal(err)
		}
		client := &http.Client{Jar: jar, Timeout: 30 * time.Second}
		var csrf string
		request := func(method, path string, body any, result any, expected int) {
			data, err := json.Marshal(body)
			if err != nil {
				t.Fatal(err)
			}
			r, err := http.NewRequestWithContext(ctx, method, ready.URL+path, bytes.NewReader(data))
			if err != nil {
				t.Fatal(err)
			}
			r.Header.Set("Content-Type", "application/json")
			r.Header.Set("Idempotency-Key", rand.Text())
			r.Header.Set("X-CSRF-Token", csrf)
			response, err := client.Do(r)
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()
			if response.StatusCode != expected {
				t.Fatal("actual cancellation HTTP status", method, path, response.StatusCode)
			}
			if err := json.NewDecoder(response.Body).Decode(result); err != nil {
				t.Fatal(err)
			}
		}
		var session struct {
			CSRF string `json:"csrf_token"`
		}
		request("GET", "/api/v1/auth/session", nil, &session, 200)
		csrf = session.CSRF
		importCommand := app.ImportMaterialCommand{CommandMeta: app.CommandMeta{SchemaVersion: 1, RequestID: rand.Text(), ExpectedVersion: 1}, Adapter: "summarize_url", Kind: "video", SourceLocator: "https://www.bilibili.com/video/BV1PReT6EEqR/"}
		if iteration == 0 {
			request("POST", "/api/v1/materials/imports", importCommand, &receipt, 202)
			whisperStarted := false
			for !whisperStarted {
				select {
				case event, ok := <-events:
					if !ok {
						t.Fatal("actual process observer ended before native startup", <-decodeError)
					}
					observe(event)
					whisperStarted = event.Kind == "whisper.cpp" && event.Phase == "started"
				case <-ctx.Done():
					t.Fatal("actual whisper process did not start", ctx.Err())
				}
			}
			var running app.Job
			request("GET", "/api/v1/jobs/"+receipt.JobID, nil, &running, 200)
			if running.Status != "running" || running.MaterialID != nil || running.Attempts != 1 {
				t.Fatal("native startup did not match actual running import job")
			}
			meta := app.CommandMeta{SchemaVersion: 1, RequestID: rand.Text(), ExpectedVersion: running.Version}
			request("POST", "/api/v1/jobs/"+receipt.JobID+"/cancel", meta, &cancelled, 200)
			if cancelled.Status != "cancelled" || !cancelled.CancelRequested || cancelled.DeliveryUnknown || cancelled.MaterialID != nil || cancelled.MaterialRevision != nil {
				t.Fatal("actual source cancellation changed terminal semantics")
			}
			nodeExited := false
			for !nodeExited {
				select {
				case event, ok := <-events:
					if !ok {
						t.Fatal("actual process observer ended before bridge exit", <-decodeError)
					}
					observe(event)
					nodeExited = event.Kind == "node" && event.Phase == "exited"
				case <-ctx.Done():
					t.Fatal("cancelled owned bridge did not exit", ctx.Err())
				}
			}
		}
		var stored app.Job
		request("GET", "/api/v1/jobs/"+receipt.JobID, nil, &stored, 200)
		if stored.Status != "cancelled" || stored.Attempts != 1 || stored.OperationID != cancelled.OperationID || stored.MaterialID != nil || stored.DeliveryUnknown {
			t.Fatal("cancelled job changed after stop or restart")
		}
		var duplicate app.ImportJobResult
		request("POST", "/api/v1/materials/imports", importCommand, &duplicate, 202)
		if duplicate.JobID != receipt.JobID || duplicate.Status != "cancelled" {
			t.Fatal("cancelled source replay scheduled extraction")
		}
		if err := stdin.Close(); err != nil {
			t.Fatal(err)
		}
		for event := range events {
			observe(event)
		}
		if err := <-decodeError; err != nil {
			t.Fatal("actual child observations invalid", err)
		}
		if err := command.Wait(); err != nil {
			t.Fatal("actual Go service child did not exit cleanly", err)
		}
		if err := stdout.Close(); err != nil {
			t.Fatal(err)
		}
		for pid, handle := range handles {
			status, err := windows.WaitForSingleObject(handle, 5000)
			if err != nil || status != windows.WAIT_OBJECT_0 {
				t.Fatal("owned native process exit unverified", pid, status, err)
			}
		}
		status, err := windows.WaitForSingleObject(goProcess, 5000)
		if err != nil || status != windows.WAIT_OBJECT_0 {
			t.Fatal("owned Go HTTP process exit unverified", status, err)
		}
		cancel()
	}
	db, err := sqlite.Open(filepath.Join(root, "state.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for table, expected := range map[string]int{"attention_jobs": 1, "attention_materials": 0, "attention_material_revisions": 0, "attention_distillations": 0, "attention_opportunities": 0} {
		var count int
		if err := db.Conn().QueryRow("SELECT count(*) FROM " + table).Scan(&count); err != nil || count != expected {
			t.Fatal("cancel/restart produced unexpected rows", table, count, err)
		}
	}
	t.Logf("actual native startup cancellation: tracked_starts=%d owned_go_exits=2 native_exits_verified=true cancelled_jobs=1 revisions=0 replay_reused=true restart_no_extraction=true mid_asr_progress=NOT_RUN no_agent_model_calls=true evidence=%s", startedProcesses, root)
}

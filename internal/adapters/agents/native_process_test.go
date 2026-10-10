package agents

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type nativeShortWriter struct{}

func (nativeShortWriter) Write(p []byte) (int, error) { return len(p) - 1, nil }
func (nativeShortWriter) Close() error                { return nil }

func TestNativeShortWriteIsUnknownAndNeverRetried(t *testing.T) {
	p := &nativeProcess{stdin: nativeShortWriter{}, done: make(chan struct{}), pending: map[string]chan map[string]json.RawMessage{}}
	_, err := p.call(context.Background(), "turn/start", map[string]any{}, false)
	var failure *apierrors.ServiceError
	if !errors.As(err, &failure) || failure.Code != apierrors.DeliveryUnknown || p.next != 1 || len(p.pending) != 0 {
		t.Fatalf("partial native write was retried or marked delivered: %v", err)
	}
}

func TestNativeRepliesAreCorrelatedAndBoundedOutputCancels(t *testing.T) {
	one, two := make(chan map[string]json.RawMessage, 1), make(chan map[string]json.RawMessage, 1)
	cancelled := make(chan struct{}, 1)
	p := &nativeProcess{done: make(chan struct{}), pending: map[string]chan map[string]json.RawMessage{"1": one, "2": two}, cmd: &exec.Cmd{Cancel: func() error { cancelled <- struct{}{}; return nil }}}
	p.read(strings.NewReader("{\"id\":\"2\",\"result\":\"second\"}\n{\"id\":\"1\",\"result\":\"first\"}\n"))
	if string((<-one)["result"]) != `"first"` || string((<-two)["result"]) != `"second"` {
		t.Fatal("out-of-order RPC reply crossed requests")
	}
	p.mu.Lock()
	p.addEvent("text", strings.Repeat("x", 128*1024))
	p.addEvent("text", "overflow")
	p.mu.Unlock()
	select {
	case <-cancelled:
	case <-time.After(time.Second):
		t.Fatal("output overflow did not cancel the owned process")
	}
	obs := p.observe()
	if !obs.OutputTruncated || obs.Status != "blocked" || len(obs.Events) != 1 || len(obs.Events[0].Text) != 128*1024 {
		t.Fatal("output overflow escaped retention bounds")
	}
	if obs.StopConfirmed {
		t.Fatal("kill request was fabricated as confirmed exit")
	}
}

func TestNativeBudgetIsSharedAcrossCLIsAndFailedLaunchReleasesSlot(t *testing.T) {
	t.Setenv("PATH", "")
	r := NewRegistry(t.TempDir())
	n := r.adapters["codex"]
	for i := 0; i < cap(n.slots); i++ {
		n.slots <- struct{}{}
	}
	other := r.adapters["claude"]
	s, err := other.Start(context.Background(), domain.NativeRequest{SessionID: "budget", Project: domain.LocalProject{ID: "p", Root: t.TempDir()}})
	var failure *apierrors.ServiceError
	if !errors.As(err, &failure) || failure.Code != apierrors.BudgetExhausted || !s.StopConfirmed || s.Ownership != "unstarted" {
		t.Fatal("global native budget did not block a different CLI")
	}
	<-n.slots
	_, err = other.Start(context.Background(), domain.NativeRequest{SessionID: "missing", Project: domain.LocalProject{ID: "p", Root: t.TempDir()}})
	if !errors.As(err, &failure) || failure.Code != apierrors.ProviderUnavailable || len(n.slots) != cap(n.slots)-1 {
		t.Fatal("unstarted failure leaked a global process slot")
	}
}

func TestCachedConfigurationCannotBecomeFreshByObservingOldResult(t *testing.T) {
	r := NewRegistry(t.TempDir())
	n := r.adapters["claude"]
	old := time.Now().Add(-10 * time.Minute)
	p := &nativeProcess{model: "native-observed", provider: "configured_cli_transport", modelObservedAt: old}
	n.refreshConfiguration(p)
	n.refreshConfiguration(p)
	if !n.configAt.Equal(old) {
		t.Fatal("cached read fabricated fresh model observation")
	}
	if _, err := r.ConfigurationID(context.Background(), "claude"); err == nil {
		t.Fatal("stale configuration became processing authority")
	}
}

var _ io.WriteCloser = nativeShortWriter{}

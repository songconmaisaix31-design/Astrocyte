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

func TestNativeStreamingSmallDeltasPreserveFullMessageAndComplete(t *testing.T) {
	cancelled := make(chan struct{}, 1)
	p := &nativeProcess{done: make(chan struct{}), cmd: &exec.Cmd{Cancel: func() error {
		select {
		case cancelled <- struct{}{}:
		default:
		}
		return nil
	}}}
	event := func(method string, params any) {
		raw, _ := json.Marshal(params)
		p.mu.Lock()
		p.handleEvent(method, "", map[string]json.RawMessage{"params": raw})
		p.mu.Unlock()
	}
	event("turn/started", map[string]any{"turn": map[string]any{"id": "turn-1"}})
	for i := 0; i < 1000; i++ {
		event("item/agentMessage/delta", map[string]any{"turnId": "turn-1", "itemId": "message-1", "delta": "摘"})
	}
	event("turn/completed", map[string]any{"turn": map[string]any{"id": "turn-1", "status": "completed"}})
	obs := p.observe()
	if obs.OutputTruncated || obs.Status != "completed" || len(obs.Events) != 3 || obs.Events[1].Text != strings.Repeat("摘", 1000) {
		t.Fatalf("small deltas consumed the event budget: status=%s truncated=%t events=%d bytes=%d", obs.Status, obs.OutputTruncated, len(obs.Events), p.bytes)
	}
	select {
	case <-cancelled:
		t.Fatal("reasonable streamed summary cancelled owned process")
	default:
	}
}

func TestNativeStreamingDoesNotMergeMessagesOrControlsAndSnapshotsStayStable(t *testing.T) {
	p := &nativeProcess{done: make(chan struct{})}
	p.addTextDelta("first-message", "a")
	first := p.observe()
	p.addTextDelta("first-message", "b")
	p.addTextDelta("second-message", "c")
	p.addEvent("turn_completed", "")
	p.addTextDelta("second-message", "d")
	obs := p.observe()
	if len(obs.Events) != 4 || obs.Events[0].Text != "ab" || obs.Events[1].Text != "c" || obs.Events[2].Kind != "turn_completed" || obs.Events[3].Text != "d" {
		t.Fatal("streamed deltas crossed a message or control boundary")
	}
	if first.Events[0].Text != "a" || obs.Events[0].Sequence != 1 || obs.Events[3].Sequence != 4 || p.bytes != 4 {
		t.Fatal("observe lost cumulative text, snapshot stability or byte accounting")
	}
}

func TestNativePiStreamingUsesMessageBoundaries(t *testing.T) {
	p := &nativeProcess{done: make(chan struct{})}
	p.handleEvent("", "message_start", nil)
	for i := 0; i < 1000; i++ {
		p.handleEvent("", "message_update", map[string]json.RawMessage{"assistantMessageEvent": json.RawMessage(`{"type":"text_delta","delta":"x"}`)})
	}
	p.handleEvent("", "message_end", nil)
	p.handleEvent("", "message_start", nil)
	p.handleEvent("", "message_update", map[string]json.RawMessage{"assistantMessageEvent": json.RawMessage(`{"type":"text_delta","delta":"y"}`)})
	p.handleEvent("", "agent_settled", nil)
	obs := p.observe()
	if obs.Status != "completed" || obs.OutputTruncated || len(obs.Events) != 3 || obs.Events[0].Text != strings.Repeat("x", 1000) || obs.Events[1].Text != "y" {
		t.Fatal("Pi delta retention lost message boundaries or full output")
	}
}

func TestNativeStreamingStillCancelsByteAndControlOverflow(t *testing.T) {
	for _, kind := range []string{"bytes", "controls"} {
		t.Run(kind, func(t *testing.T) {
			cancelled := make(chan struct{}, 1)
			p := &nativeProcess{done: make(chan struct{}), cmd: &exec.Cmd{Cancel: func() error { cancelled <- struct{}{}; return nil }}}
			if kind == "bytes" {
				p.addTextDelta("one-message", strings.Repeat("x", 128*1024))
				p.addTextDelta("one-message", "overflow")
			} else {
				for i := 0; i < 257; i++ {
					p.addEvent("turn_started", "")
				}
			}
			p.handleEvent("turn/completed", "", map[string]json.RawMessage{"params": json.RawMessage(`{"turn":{"status":"completed"}}`)})
			obs := p.observe()
			if !obs.OutputTruncated || obs.Status != "blocked" || obs.StopConfirmed || p.bytes > 128*1024 || len(obs.Events) > 256 {
				t.Fatal("real overflow escaped original bounds or became completed/stopped")
			}
			select {
			case <-cancelled:
			case <-time.After(time.Second):
				t.Fatal("overflow did not cancel owned process")
			}
		})
	}
}

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

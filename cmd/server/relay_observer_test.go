package main

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"
)

type relayBlockingObserver struct {
	entered, exited chan struct{}
	calls           int
}

func (o *relayBlockingObserver) ObserveOwned(ctx context.Context) error {
	o.calls++
	close(o.entered)
	<-ctx.Done()
	close(o.exited)
	return ctx.Err()
}
func TestRelayObserverCancellationJoinsBeforeStateCloses(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	o := &relayBlockingObserver{entered: make(chan struct{}), exited: make(chan struct{})}
	done := make(chan struct{})
	go func() {
		defer close(done)
		runRelayObserver(ctx, o, time.Millisecond, slog.New(slog.NewTextHandler(io.Discard, nil)))
	}()
	select {
	case <-o.entered:
	case <-time.After(time.Second):
		t.Fatal("owned observer did not start")
	}
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("observer did not join on shutdown")
	}
	select {
	case <-o.exited:
	default:
		t.Fatal("observer still writing after join")
	}
	if o.calls != 1 {
		t.Fatal("cancelled observer retried")
	}
}

package misc

import (
	"errors"
	"testing"
	"time"
)

func TestStopWithError(t *testing.T) {
	// The error is kept for as long as the process lives
	stopErr = nil
	t.Cleanup(func() { stopErr = nil })

	if err := StopError(); err != nil {
		t.Fatalf("unexpected error before the TUI was stopped: %v", err)
	}

	first := errors.New("first")
	StopWithError(first)
	StopWithError(errors.New("second"))

	// The error that stopped the TUI is the one to report
	if err := StopError(); err != first {
		t.Fatalf("got error %v, want %v", err, first)
	}
}

func TestDeferReload(t *testing.T) {
	reloads := make(chan struct{}, 4)
	ReloadConfig = func() { reloads <- struct{}{} }
	t.Cleanup(func() { ReloadConfig = nil })

	reloaded := func() bool {
		select {
		case <-reloads:
			return true
		case <-time.After(50 * time.Millisecond):
			return false
		}
	}

	if DeferReload() {
		t.Fatal("a reload should not wait when nothing is running")
	}

	// Two runs: the reload waits for the last of them
	BeginRun()
	BeginRun()
	// Asked twice, as when the config is saved twice during a run
	for range 2 {
		if !DeferReload() {
			t.Fatal("a reload should wait for the runs to finish")
		}
	}
	EndRun()
	if reloaded() {
		t.Fatal("reloaded while a run was still active")
	}
	EndRun()
	if !reloaded() {
		t.Fatal("the reload that was put off was not done when the runs had finished")
	}
	if reloaded() {
		t.Fatal("reloaded more than once")
	}

	// Nothing was put off: no reload when a run ends
	BeginRun()
	EndRun()
	if reloaded() {
		t.Fatal("reloaded although no reload was waiting")
	}
}

package misc

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
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

// A reload that is asked for while the last run ends is done exactly once: it
// is either seen by the run that ends, or it doesn't have to wait.
func TestDeferReload_RaceWithEndOfRun(t *testing.T) {
	var reloads atomic.Int32
	ReloadConfig = func() { reloads.Add(1) }
	t.Cleanup(func() { ReloadConfig = nil })

	for range 2000 {
		reloads.Store(0)
		BeginRun()

		var wg sync.WaitGroup
		var deferred bool
		wg.Add(2)
		go func() { defer wg.Done(); EndRun() }()
		go func() { defer wg.Done(); deferred = DeferReload() }()
		wg.Wait()

		// Put off: the run that ended reloads. Not put off: the caller does.
		want := int32(0)
		if deferred {
			want = 1
		}
		deadline := time.Now().Add(time.Second)
		for reloads.Load() != want && time.Now().Before(deadline) {
			time.Sleep(time.Millisecond)
		}
		if got := reloads.Load(); got != want {
			t.Fatalf("deferred = %t, reloaded %d times; want %d", deferred, got, want)
		}
		if RunActive() {
			t.Fatal("a run is still counted as active")
		}

		// Nothing may be left waiting for a later run
		BeginRun()
		EndRun()
		time.Sleep(time.Millisecond)
		if got := reloads.Load(); got != want {
			t.Fatalf("a reload was left over and done after a later run (%d reloads, want %d)", got, want)
		}
	}
}

func TestFilesState(t *testing.T) {
	file := filepath.Join(t.TempDir(), "mani.yaml")
	if err := os.WriteFile(file, []byte("a: 1\n"), 0644); err != nil {
		t.Fatal(err)
	}
	before := FilesState(file)

	// Reading the file, or changing its attributes, is not a write
	if err := os.Chmod(file, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := os.ReadFile(file); err != nil {
		t.Fatal(err)
	}
	if got := FilesState(file); got != before {
		t.Fatalf("the state changed without a write: %q -> %q", before, got)
	}

	if err := os.WriteFile(file, []byte("a: 1\nb: 2\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if got := FilesState(file); got == before {
		t.Fatal("the state is the same after the file was written")
	}

	if got := FilesState(filepath.Join(t.TempDir(), "missing.yaml")); got == "" {
		t.Fatal("a missing file should have a state too")
	}
}

package tui

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
)

func TestWatchLoop(t *testing.T) {
	events := make(chan fsnotify.Event)
	errs := make(chan error)
	var reloads atomic.Int32

	const delay = 40 * time.Millisecond
	done := make(chan struct{})
	go func() {
		defer close(done)
		watchLoop(events, errs, []string{"/ws/mani.yaml"}, delay, func() { reloads.Add(1) })
	}()

	// A save is several events: reloaded once, after they have stopped
	for range 5 {
		events <- fsnotify.Event{Name: "/ws/mani.yaml", Op: fsnotify.Write}
		time.Sleep(delay / 8)
	}
	if got := reloads.Load(); got != 0 {
		t.Fatalf("reloaded %d times while the file was still being written", got)
	}
	time.Sleep(3 * delay)
	if got := reloads.Load(); got != 1 {
		t.Fatalf("reloaded %d times after one save; want 1", got)
	}

	// Other files in the directory are not the config
	events <- fsnotify.Event{Name: "/ws/notes.txt", Op: fsnotify.Write}
	time.Sleep(3 * delay)
	if got := reloads.Load(); got != 1 {
		t.Fatalf("reloaded on a change to another file (%d reloads)", got)
	}

	// Each save is reloaded, also when the file has the same modification time
	events <- fsnotify.Event{Name: "/ws/mani.yaml", Op: fsnotify.Write}
	time.Sleep(3 * delay)
	if got := reloads.Load(); got != 2 {
		t.Fatalf("reloaded %d times after two saves; want 2", got)
	}

	close(events)
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("watchLoop did not return when the events were closed")
	}
}

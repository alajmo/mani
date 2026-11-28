package tui

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
)

func TestWatchLoop(t *testing.T) {
	events := make(chan fsnotify.Event)
	errs := make(chan error)
	var reloads atomic.Int32

	// What the file holds: changed by write, as a save would
	var content atomic.Int32
	write := func() {
		content.Add(1)
		events <- fsnotify.Event{Name: "/ws/mani.yaml", Op: fsnotify.Write}
	}
	state := func() string { return fmt.Sprint(content.Load()) }

	const delay = 40 * time.Millisecond
	done := make(chan struct{})
	go func() {
		defer close(done)
		watchLoop(events, errs, []string{"/ws/mani.yaml"}, delay, state, func() { reloads.Add(1) })
	}()

	// A save is several events: reloaded once, after they have stopped
	for range 5 {
		write()
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

	// An event that isn't a write (the attributes of the file were changed by
	// an indexer or a backup) leaves the file as it was: nothing to reload
	events <- fsnotify.Event{Name: "/ws/mani.yaml", Op: fsnotify.Chmod}
	time.Sleep(3 * delay)
	if got := reloads.Load(); got != 1 {
		t.Fatalf("reloaded although the file was not written (%d reloads)", got)
	}

	// The next save is reloaded
	write()
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

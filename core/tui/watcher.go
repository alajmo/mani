package tui

import (
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/fsnotify/fsnotify"
)

func WatchFiles(app *App, files ...string) {
	if len(files) < 1 {
		return
	}

	w, err := fsnotify.NewWatcher()
	if err != nil {
		os.Exit(1)
	}

	go func() {
		defer func() {
			_ = w.Close()
		}()

		for _, p := range files {
			st, err := os.Lstat(p)
			if err != nil {
				os.Exit(1)
			}
			if st.IsDir() {
				os.Exit(1)
			}
			err = w.Add(filepath.Dir(p))
			if err != nil {
				os.Exit(1)
			}
		}

		watchLoop(w.Events, w.Errors, files, reloadDelay, app.Reload)
	}()
}

// reloadDelay is how long the config has to be left alone before it's
// reloaded.
const reloadDelay = 150 * time.Millisecond

// watchLoop calls reload when one of files has changed.
//
// A save is often several events (truncate, write, rename), and the file isn't
// complete at the first of them. The reload is therefore done once the events
// have stopped for delay, which also makes it one reload per save.
func watchLoop(events <-chan fsnotify.Event, errs <-chan error, files []string, delay time.Duration, reload func()) {
	timer := time.NewTimer(delay)
	timer.Stop()
	defer timer.Stop()

	for {
		select {
		case _, ok := <-errs:
			if !ok {
				return
			}
			os.Exit(1)
		case e, ok := <-events:
			if !ok {
				return
			}
			if slices.Contains(files, e.Name) {
				timer.Reset(delay)
			}
		case <-timer.C:
			reload()
		}
	}
}

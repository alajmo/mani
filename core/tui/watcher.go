package tui

import (
	"os"
	"path/filepath"
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

		lastMod := make(map[string]time.Time)
		for {
			select {
			case _, ok := <-w.Errors:
				if !ok {
					return
				}
				os.Exit(1)
			case e, ok := <-w.Events:
				if !ok {
					return
				}

				for _, f := range files {
					if f == e.Name {
						stat, err := os.Stat(f)
						if err != nil {
							continue
						}

						currentMod := stat.ModTime()
						if lastMod[f] != currentMod {
							// Reload marshals widget mutation onto the event loop
							// via QueueUpdateDraw, so the old sleep hack that
							// worked around the off-loop race is no longer needed.
							app.Reload()
							lastMod[f] = currentMod
						}
						break
					}
				}
			}
		}
	}()
}

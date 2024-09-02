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
		defer w.Close()
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
							// TODO: For some reason, the reload is not working correctly, must be due to it being called in a goroutine
							// Sleeping resolves it somehow.
							time.Sleep(500 * time.Millisecond)
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

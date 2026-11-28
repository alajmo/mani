package misc

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/alajmo/mani/core/dao"
	"github.com/rivo/tview"
)

var Config *dao.Config
var ThemeName *string
var TUITheme *dao.TUI
var BlockTheme *dao.Block

var App *tview.Application
var Pages *tview.Pages
var MainPage *tview.Pages
var PreviousPane tview.Primitive

// ReloadConfig re-reads the config from disk and rebuilds the UI on the event
// loop. The top-level tui package wires this up at startup so lower-level
// packages (e.g. views) can trigger the reload path without importing tui,
// which would create an import cycle.
var ReloadConfig func()

var PreviousModel interface{}

// Nav
var ProjectBtn *tview.Button
var TaskBtn *tview.Button
var RunBtn *tview.Button
var ExecBtn *tview.Button
var HelpBtn *tview.Button

var ProjectsLastFocus *tview.Primitive
var TasksLastFocus *tview.Primitive
var RunLastFocus *tview.Primitive
var ExecLastFocus *tview.Primitive

// Misc
var HelpModal *tview.Modal
var Search *tview.InputField

var (
	stopErrMu sync.Mutex
	stopErr   error

	// Runs in flight, and whether a reload waits for them. Both are changed
	// together: a run ends on its own goroutine while a reload is asked for
	// on the event loop.
	runMu        sync.Mutex
	activeRuns   int
	reloadQueued bool
)

// Watching is set when the config files are watched, and reloaded when they
// change.
var Watching bool

// BeginRun and EndRun bracket a run of a task or command. BeginRun is called
// on the event loop, EndRun when the run has finished.
//
// A reload rebuilds every page. A run in flight would keep writing to an
// output view that is no longer shown, and the new page would accept another
// run next to it. A reload therefore waits until no run is active: see
// DeferReload.
func BeginRun() {
	runMu.Lock()
	activeRuns++
	runMu.Unlock()
}

func EndRun() {
	runMu.Lock()
	activeRuns--
	reload := activeRuns == 0 && reloadQueued
	if reload {
		reloadQueued = false
	}
	runMu.Unlock()

	if reload && ReloadConfig != nil {
		go ReloadConfig()
	}
}

// RunActive reports whether a task or command is running.
func RunActive() bool {
	runMu.Lock()
	defer runMu.Unlock()

	return activeRuns > 0
}

// DeferReload reports whether a reload has to wait for a run to finish. If so
// it's done when the last run ends. It's called on the event loop, where runs
// start, so that a run can't start between this check and the reload.
func DeferReload() bool {
	runMu.Lock()
	defer runMu.Unlock()

	if activeRuns > 0 {
		reloadQueued = true
		return true
	}
	return false
}

// FilesState returns a value that is different when one of the files has been
// written: their modification times and sizes.
func FilesState(files ...string) string {
	var state strings.Builder
	for _, file := range files {
		if info, err := os.Stat(file); err == nil {
			fmt.Fprintf(&state, "%s:%d:%d;", file, info.ModTime().UnixNano(), info.Size())
		} else {
			fmt.Fprintf(&state, "%s:missing;", file)
		}
	}
	return state.String()
}

// ConfigFilesState is FilesState for the files of the config.
func ConfigFilesState() string {
	if Config == nil {
		return ""
	}
	return FilesState(append([]string{Config.Path}, Config.ConfigPaths...)...)
}

// StopWithError stops the TUI because of err. The error can't be shown while
// the TUI owns the terminal, so it's kept for StopError and reported by
// RunTui after the application has stopped.
func StopWithError(err error) {
	stopErrMu.Lock()
	if stopErr == nil {
		stopErr = err
	}
	stopErrMu.Unlock()

	if App != nil {
		App.Stop()
	}
}

// StopError returns the error that stopped the TUI, if any.
func StopError() error {
	stopErrMu.Lock()
	defer stopErrMu.Unlock()

	return stopErr
}

// outputView is the output of a page: its text view, whether it's the view
// that is shown, and how to show it.
type outputView struct {
	view  *tview.TextView
	shown func() bool
	show  func()
}

var outputs = map[string]outputView{}

// RegisterOutput makes the output of a page known, so that it's kept when the
// pages are rebuilt by a reload. A page registers again when it's rebuilt.
func RegisterOutput(page string, view *tview.TextView, shown func() bool, show func()) {
	outputs[page] = outputView{view: view, shown: shown, show: show}
}

// SaveOutputs takes the output of every page before the pages are rebuilt, and
// returns a function that puts it back on the new pages. Without it a reload
// (which may have waited for a run to finish) would wipe the output of that
// run from the screen.
func SaveOutputs() (restore func()) {
	type saved struct {
		text  string
		shown bool
	}
	state := map[string]saved{}
	for page, output := range outputs {
		state[page] = saved{text: output.view.GetText(false), shown: output.shown()}
	}

	return func() {
		for page, s := range state {
			output, ok := outputs[page]
			if !ok || s.text == "" {
				continue
			}
			output.view.SetText(s.text)
			output.view.ScrollToEnd()
			if s.shown {
				output.show()
			}
		}
	}
}

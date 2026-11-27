package misc

import (
	"sync"
	"sync/atomic"

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

	activeRuns   atomic.Int32
	reloadQueued atomic.Bool
)

// BeginRun and EndRun bracket a run of a task or command. BeginRun is called
// on the event loop, EndRun when the run has finished.
//
// A reload rebuilds every page. A run in flight would keep writing to an
// output view that is no longer shown, and the new page would accept another
// run next to it. A reload therefore waits until no run is active: see
// DeferReload.
func BeginRun() {
	activeRuns.Add(1)
}

func EndRun() {
	if activeRuns.Add(-1) == 0 && reloadQueued.Swap(false) && ReloadConfig != nil {
		go ReloadConfig()
	}
}

// DeferReload reports whether a reload has to wait for a run to finish. If so
// it's done when the last run ends. It's called on the event loop, where runs
// start, so that a run can't start between this check and the reload.
func DeferReload() bool {
	if activeRuns.Load() > 0 {
		reloadQueued.Store(true)
		return true
	}
	return false
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

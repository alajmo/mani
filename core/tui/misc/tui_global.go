package misc

import (
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

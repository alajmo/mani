package misc

import (
	"github.com/alajmo/mani/core/dao"
	"github.com/rivo/tview"
)

var Config *dao.Config

var App *tview.Application
var Pages *tview.Pages
var MainPage *tview.Pages
var PreviousPage tview.Primitive

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

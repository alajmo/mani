package misc

import (
	"github.com/alajmo/mani/core/dao"
	"github.com/rivo/tview"
)

var Config *dao.Config
var Emitter *EventEmitter

var App *tview.Application
var NavPane *tview.Flex
var Pages *tview.Pages
var MainPage *tview.Pages
var PreviousPage tview.Primitive

// Nav
var ProjectBtn *tview.Button
var TaskBtn *tview.Button
var RunBtn *tview.Button
var ExecBtn *tview.Button
var HelpBtn *tview.Button

// Run
var RunPage *tview.Flex

// Exec
var ExecPage *tview.Flex

// Misc
var HelpModal *tview.Modal
var Search *tview.InputField

package tui

import (
	"os"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
	"github.com/alajmo/mani/core/tui/components"
	"github.com/alajmo/mani/core/tui/misc"
	"github.com/rivo/tview"
)

func RunTui(config *dao.Config, themeName string, reload bool) {
	app := NewApp(config, themeName)

	if reload {
		WatchFiles(app, append([]string{config.Path}, config.ConfigPaths...)...)
	}

	if err := app.Run(); err != nil {
		os.Exit(1)
	}
}

type App struct {
	App  *tview.Application
	Path string
}

func NewApp(config *dao.Config, themeName string) *App {
	app := &App{
		App:  tview.NewApplication(),
		Path: config.Path,
	}
	app.setupApp(config, themeName)

	return app
}

func (app *App) Run() error {
	return app.App.SetRoot(misc.Pages, true).EnableMouse(true).Run()
}

func (app *App) Reload() {
	// Read + parse the config off the event loop so disk I/O does not block the UI.
	config, configErr := dao.ReadConfig(app.Path, "", true)
	if configErr != nil {
		// A transient bad save while editing the watched file must not kill the
		// TUI. Surface the parse error and keep the previously loaded config.
		// The error is ANSI-colorized; strip it (a tview TextView renders tview
		// tags, not ANSI) and escape stray brackets so it shows as plain text.
		errPlain := core.Strip(configErr.Error())
		app.App.QueueUpdateDraw(func() {
			components.OpenTextModal("reload-error-modal", tview.Escape(errPlain), errPlain, "Config Reload Error")
		})
		return
	}

	// tview is single-threaded: all widget mutation (rebuilding misc.Pages,
	// re-registering the input capture, SetRoot) must run on the event loop.
	// QueueUpdateDraw is safe to call from F5's goroutine and the watcher goroutine.
	app.App.QueueUpdateDraw(func() {
		app.setupApp(&config, *misc.ThemeName)
		app.App.SetRoot(misc.Pages, true)
	})
}

func (app *App) setupApp(config *dao.Config, themeName string) {
	misc.Config = config
	misc.ReloadConfig = app.Reload
	misc.ThemeName = &themeName
	theme, err := misc.Config.GetTheme(themeName)
	core.CheckIfError(err)

	misc.LoadStyles(&theme.TUI)
	misc.TUITheme = &theme.TUI
	misc.BlockTheme = &theme.Block

	// Data
	projects := config.ProjectList
	tasks := config.TaskList
	dao.ParseTasksEnv(tasks)
	projectTags := config.GetTags()
	projectPaths := config.GetProjectPaths()

	// Styles
	setupStyles()

	// Create pages
	misc.App = app.App
	misc.Pages = createPages(projects, projectTags, projectPaths, tasks)

	// Global input handling
	HandleInput(app)
}

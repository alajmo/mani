package tui

import (
	"os"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
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
	App *tview.Application
}

func NewApp(config *dao.Config, themeName string) *App {
	app := &App{
		App: tview.NewApplication(),
	}
	app.setupApp(config, themeName)

	return app
}

func (app *App) Run() error {
	return app.App.SetRoot(misc.Pages, true).EnableMouse(true).Run()
}

func (app *App) Reload() {
	config, configErr := dao.ReadConfig(misc.Config.Path, "", true)
	if configErr != nil {
		app.App.Stop()
	}

	app.setupApp(&config, *misc.ThemeName)
	app.App.SetRoot(misc.Pages, true)
	app.App.Draw()
}

func (app *App) setupApp(config *dao.Config, themeName string) {
	misc.Config = config
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

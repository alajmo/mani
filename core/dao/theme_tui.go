package dao

import (
	"github.com/alajmo/mani/core"
)

// Not all attributes are used, but no clean way to add them since
// MergeThemeOptions initializes all of the fields.
var DefaultTUI = TUI{
	Default: &ColorOptions{
		Fg:     core.Ptr(""),
		Bg:     core.Ptr(""),
		Attr:   core.Ptr(""),
		Format: core.Ptr(""),
	},

	Border: &ColorOptions{
		Fg:     core.Ptr(""),
		Bg:     core.Ptr(""),
		Attr:   core.Ptr(""),
		Format: core.Ptr(""),
	},
	BorderFocus: &ColorOptions{
		Fg:     core.Ptr("#d787ff"),
		Bg:     core.Ptr(""),
		Attr:   core.Ptr(""),
		Format: core.Ptr(""),
	},

	Title: &ColorOptions{
		Fg:     core.Ptr(""),
		Bg:     core.Ptr(""),
		Attr:   core.Ptr(""),
		Align:  core.Ptr("center"),
		Format: core.Ptr(""),
	},
	TitleActive: &ColorOptions{
		Fg:     core.Ptr("#000000"),
		Bg:     core.Ptr("#d787ff"),
		Attr:   core.Ptr(""),
		Align:  core.Ptr("center"),
		Format: core.Ptr(""),
	},

	Button: &ColorOptions{
		Fg:     core.Ptr(""),
		Bg:     core.Ptr(""),
		Attr:   core.Ptr(""),
		Align:  core.Ptr(""),
		Format: core.Ptr(""),
	},
	ButtonActive: &ColorOptions{
		Fg:     core.Ptr("#080808"),
		Bg:     core.Ptr("#d787ff"),
		Attr:   core.Ptr(""),
		Align:  core.Ptr(""),
		Format: core.Ptr(""),
	},

	Item: &ColorOptions{
		Fg:     core.Ptr(""),
		Bg:     core.Ptr(""),
		Attr:   core.Ptr(""),
		Format: core.Ptr(""),
	},
	ItemFocused: &ColorOptions{
		Fg:     core.Ptr("#ffffff"),
		Bg:     core.Ptr("#262626"),
		Attr:   core.Ptr(""),
		Format: core.Ptr(""),
	},
	ItemSelected: &ColorOptions{
		Fg:     core.Ptr("#5f87d7"),
		Bg:     core.Ptr(""),
		Attr:   core.Ptr(""),
		Format: core.Ptr(""),
	},
	ItemDir: &ColorOptions{
		Fg:     core.Ptr("#d787ff"),
		Bg:     core.Ptr(""),
		Attr:   core.Ptr(""),
		Format: core.Ptr(""),
	},
	ItemRef: &ColorOptions{
		Fg:     core.Ptr("#d787ff"),
		Bg:     core.Ptr(""),
		Attr:   core.Ptr(""),
		Format: core.Ptr(""),
	},

	TableHeader: &ColorOptions{
		Fg:     core.Ptr("#d787ff"),
		Bg:     core.Ptr(""),
		Attr:   core.Ptr("bold"),
		Align:  core.Ptr("left"),
		Format: core.Ptr(""),
	},

	SearchLabel: &ColorOptions{
		Fg:     core.Ptr("#d7d75f"),
		Bg:     core.Ptr(""),
		Attr:   core.Ptr("bold"),
		Format: core.Ptr(""),
	},
	SearchText: &ColorOptions{
		Fg:     core.Ptr(""),
		Bg:     core.Ptr(""),
		Attr:   core.Ptr(""),
		Format: core.Ptr(""),
	},

	FilterLabel: &ColorOptions{
		Fg:     core.Ptr("#d7d75f"),
		Bg:     core.Ptr(""),
		Attr:   core.Ptr("bold"),
		Format: core.Ptr(""),
	},
	FilterText: &ColorOptions{
		Fg:     core.Ptr(""),
		Bg:     core.Ptr(""),
		Attr:   core.Ptr(""),
		Format: core.Ptr(""),
	},

	ShortcutLabel: &ColorOptions{
		Fg:     core.Ptr("#00af5f"),
		Bg:     core.Ptr(""),
		Attr:   core.Ptr(""),
		Format: core.Ptr(""),
	},
	ShortcutText: &ColorOptions{
		Fg:     core.Ptr(""),
		Bg:     core.Ptr(""),
		Attr:   core.Ptr(""),
		Format: core.Ptr(""),
	},
}

type TUI struct {
	Default *ColorOptions `yaml:"default"`

	Border      *ColorOptions `yaml:"border"`
	BorderFocus *ColorOptions `yaml:"border_focus"`

	Title       *ColorOptions `yaml:"title"`
	TitleActive *ColorOptions `yaml:"title_active"`

	TableHeader *ColorOptions `yaml:"table_header"`

	Item         *ColorOptions `yaml:"item"`
	ItemFocused  *ColorOptions `yaml:"item_focused"`
	ItemSelected *ColorOptions `yaml:"item_selected"`
	ItemDir      *ColorOptions `yaml:"item_dir"`
	ItemRef      *ColorOptions `yaml:"item_ref"`

	Button       *ColorOptions `yaml:"button"`
	ButtonActive *ColorOptions `yaml:"button_active"`

	SearchLabel *ColorOptions `yaml:"search_label"`
	SearchText  *ColorOptions `yaml:"search_text"`

	FilterLabel *ColorOptions `yaml:"filter_label"`
	FilterText  *ColorOptions `yaml:"filter_text"`

	ShortcutLabel *ColorOptions `yaml:"shortcut_label"`
	ShortcutText  *ColorOptions `yaml:"shortcut_text"`
}

func LoadTUITheme(tui *TUI) {
	tui.Default = MergeThemeOptions(tui.Default, DefaultTUI.Default)

	tui.Border = MergeThemeOptions(tui.Border, DefaultTUI.Border)
	tui.BorderFocus = MergeThemeOptions(tui.BorderFocus, DefaultTUI.BorderFocus)

	tui.Button = MergeThemeOptions(tui.Button, DefaultTUI.Button)
	tui.ButtonActive = MergeThemeOptions(tui.ButtonActive, DefaultTUI.ButtonActive)

	tui.Item = MergeThemeOptions(tui.Item, DefaultTUI.Item)
	tui.ItemFocused = MergeThemeOptions(tui.ItemFocused, DefaultTUI.ItemFocused)
	tui.ItemSelected = MergeThemeOptions(tui.ItemSelected, DefaultTUI.ItemSelected)
	tui.ItemDir = MergeThemeOptions(tui.ItemDir, DefaultTUI.ItemDir)
	tui.ItemRef = MergeThemeOptions(tui.ItemRef, DefaultTUI.ItemRef)

	tui.Title = MergeThemeOptions(tui.Title, DefaultTUI.Title)
	tui.TitleActive = MergeThemeOptions(tui.TitleActive, DefaultTUI.TitleActive)

	tui.TableHeader = MergeThemeOptions(tui.TableHeader, DefaultTUI.TableHeader)

	tui.SearchLabel = MergeThemeOptions(tui.SearchLabel, DefaultTUI.SearchLabel)
	tui.SearchText = MergeThemeOptions(tui.SearchText, DefaultTUI.SearchText)

	tui.FilterLabel = MergeThemeOptions(tui.FilterLabel, DefaultTUI.FilterLabel)
	tui.FilterText = MergeThemeOptions(tui.FilterText, DefaultTUI.FilterText)

	tui.ShortcutText = MergeThemeOptions(tui.ShortcutText, DefaultTUI.ShortcutText)
	tui.ShortcutLabel = MergeThemeOptions(tui.ShortcutLabel, DefaultTUI.ShortcutLabel)
}

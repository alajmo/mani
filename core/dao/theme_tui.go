package dao

// DefaultTUI  Not all attributes are used, but no clean way to add them since
// MergeThemeOptions initializes all of the fields.
var DefaultTUI = TUI{
	Default: &ColorOptions{
		Fg:     new(""),
		Bg:     new(""),
		Attr:   new(""),
		Format: new(""),
	},

	Border: &ColorOptions{
		Fg:     new(""),
		Bg:     new(""),
		Attr:   new(""),
		Format: new(""),
	},
	BorderFocus: &ColorOptions{
		Fg:     new("#d787ff"),
		Bg:     new(""),
		Attr:   new(""),
		Format: new(""),
	},

	Title: &ColorOptions{
		Fg:     new(""),
		Bg:     new(""),
		Attr:   new(""),
		Align:  new("center"),
		Format: new(""),
	},
	TitleActive: &ColorOptions{
		Fg:     new("#000000"),
		Bg:     new("#d787ff"),
		Attr:   new(""),
		Align:  new("center"),
		Format: new(""),
	},

	Button: &ColorOptions{
		Fg:     new(""),
		Bg:     new(""),
		Attr:   new(""),
		Align:  new(""),
		Format: new(""),
	},
	ButtonActive: &ColorOptions{
		Fg:     new("#080808"),
		Bg:     new("#d787ff"),
		Attr:   new(""),
		Align:  new(""),
		Format: new(""),
	},

	Item: &ColorOptions{
		Fg:     new(""),
		Bg:     new(""),
		Attr:   new(""),
		Format: new(""),
	},
	ItemFocused: &ColorOptions{
		Fg:     new("#ffffff"),
		Bg:     new("#262626"),
		Attr:   new(""),
		Format: new(""),
	},
	ItemSelected: &ColorOptions{
		Fg:     new("#5f87d7"),
		Bg:     new(""),
		Attr:   new(""),
		Format: new(""),
	},
	ItemDir: &ColorOptions{
		Fg:     new("#d787ff"),
		Bg:     new(""),
		Attr:   new(""),
		Format: new(""),
	},
	ItemRef: &ColorOptions{
		Fg:     new("#d787ff"),
		Bg:     new(""),
		Attr:   new(""),
		Format: new(""),
	},

	TableHeader: &ColorOptions{
		Fg:     new("#d787ff"),
		Bg:     new(""),
		Attr:   new("bold"),
		Align:  new("left"),
		Format: new(""),
	},

	SearchLabel: &ColorOptions{
		Fg:     new("#d7d75f"),
		Bg:     new(""),
		Attr:   new("bold"),
		Format: new(""),
	},
	SearchText: &ColorOptions{
		Fg:     new(""),
		Bg:     new(""),
		Attr:   new(""),
		Format: new(""),
	},

	FilterLabel: &ColorOptions{
		Fg:     new("#d7d75f"),
		Bg:     new(""),
		Attr:   new("bold"),
		Format: new(""),
	},
	FilterText: &ColorOptions{
		Fg:     new(""),
		Bg:     new(""),
		Attr:   new(""),
		Format: new(""),
	},

	ShortcutLabel: &ColorOptions{
		Fg:     new("#00af5f"),
		Bg:     new(""),
		Attr:   new(""),
		Format: new(""),
	},
	ShortcutText: &ColorOptions{
		Fg:     new(""),
		Bg:     new(""),
		Attr:   new(""),
		Format: new(""),
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

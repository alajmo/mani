package dao

import (
	"github.com/alajmo/mani/core"
)

type Theme struct {
	Name  string
	Table string
	Tree  string
}

// Populates ThemeList and creates a default theme if no default theme is set.
func (c *Config) GetThemeList() ([]Theme, error) {
	var themes []Theme
	count := len(c.Themes.Content)

	for i := 0; i < count; i += 2 {
		theme := &Theme{}
		err := c.Themes.Content[i+1].Decode(theme)
		if err != nil {
			return []Theme{}, &core.FailedToParseFile{Name: c.Path, Msg: err}
		}

		theme.Name = c.Themes.Content[i].Value
		themes = append(themes, *theme)
	}

	return themes, nil
}

func (c Config) GetTheme(name string) (*Theme, error) {
	for _, theme := range c.ThemeList {
		if name == theme.Name {
			return &theme, nil
		}
	}

	return nil, &core.ThemeNotFound{Name: name}
}

func (c Config) GetThemeNames() []string {
	names := []string{}
	for _, theme := range c.ThemeList {
		names = append(names, theme.Name)
	}

	return names
}

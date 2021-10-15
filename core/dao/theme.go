package dao

import (
	// "gopkg.in/yaml.v3"

	"github.com/alajmo/mani/core"
)

type Theme struct {
	Name   string
	Table  string
	Tree   string
}

// Populates ThemeList and creates a default theme if no default theme is set.
func (c *Config) SetThemeList() []Theme {
	var themes []Theme
	count := len(c.Themes.Content)

	for i := 0; i < count; i += 2 {
		theme := &Theme{}
		c.Themes.Content[i+1].Decode(theme)
		theme.Name = c.Themes.Content[i].Value
		themes = append(themes, *theme)
	}

	c.ThemeList = themes

	_, err := c.GetTheme(DEFAULT_THEME.Name)
	if err != nil {
		c.ThemeList = append(c.ThemeList, DEFAULT_THEME)
	}

	return themes
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

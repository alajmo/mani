package dao

import (
	// "gopkg.in/yaml.v3"

	"github.com/alajmo/mani/core"
)

type Theme struct {
	Name	string
	Table	string
	Tree	string
	Output	string
}

func (c *Config) SetThemeList() []Theme {
	var themes []Theme
	count := len(c.Themes.Content)

	for i := 0; i < count; i += 2 {
		theme := &Theme{}
		c.Themes.Content[i + 1].Decode(theme)
		theme.Name = c.Themes.Content[i].Value
		themes = append(themes, *theme)
	}

	c.ThemeList = themes

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

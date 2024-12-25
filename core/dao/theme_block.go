package dao

import (
	"github.com/alajmo/mani/core"
)

var DefaultBlock = Block{
	Key: &ColorOptions{
		Fg:     core.Ptr("#5f87d7"),
		Attr:   core.Ptr(""),
		Format: core.Ptr(""),
	},

	Separator: &ColorOptions{
		Fg:     core.Ptr("#5f87d7"),
		Attr:   core.Ptr(""),
		Format: core.Ptr(""),
	},

	Value: &ColorOptions{
		Fg:     core.Ptr(""),
		Attr:   core.Ptr(""),
		Format: core.Ptr(""),
	},

	ValueTrue: &ColorOptions{
		Fg:     core.Ptr("#00af5f"),
		Attr:   core.Ptr(""),
		Format: core.Ptr(""),
	},

	ValueFalse: &ColorOptions{
		Fg:     core.Ptr("#d75f5f"),
		Attr:   core.Ptr(""),
		Format: core.Ptr(""),
	},
}

type Block struct {
	Key        *ColorOptions `yaml:"key"`
	Separator  *ColorOptions `yaml:"separator"`
	Value      *ColorOptions `yaml:"value"`
	ValueTrue  *ColorOptions `yaml:"value_true"`
	ValueFalse *ColorOptions `yaml:"value_false"`
}

func LoadBlockTheme(block *Block) {
	if block.Key == nil {
		block.Key = DefaultBlock.Key
	} else {
		block.Key = MergeThemeOptions(block.Key, DefaultBlock.Key)
	}

	if block.Value == nil {
		block.Value = DefaultBlock.Value
	} else {
		block.Value = MergeThemeOptions(block.Value, DefaultBlock.Value)
	}

	if block.Separator == nil {
		block.Separator = DefaultBlock.Separator
	} else {
		block.Separator = MergeThemeOptions(block.Separator, DefaultBlock.Separator)
	}

	if block.ValueTrue == nil {
		block.ValueTrue = DefaultBlock.ValueTrue
	} else {
		block.ValueTrue = MergeThemeOptions(block.ValueTrue, DefaultBlock.ValueTrue)
	}

	if block.ValueFalse == nil {
		block.ValueFalse = DefaultBlock.ValueFalse
	} else {
		block.ValueFalse = MergeThemeOptions(block.ValueFalse, DefaultBlock.ValueFalse)
	}
}

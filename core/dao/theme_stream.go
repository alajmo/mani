package dao

type Stream struct {
	Prefix       bool     `yaml:"prefix"`
	PrefixColors []string `yaml:"prefix_colors"`
	Header       bool     `yaml:"header"`
	HeaderChar   string   `yaml:"header_char"`
	HeaderPrefix string   `yaml:"header_prefix"`
}

var DefaultStream = Stream{
	Prefix:       true,
	Header:       true,
	HeaderPrefix: "TASK",
	HeaderChar:   "*",
	PrefixColors: []string{"#d787ff", "#00af5f", "#d75f5f", "#5f87d7", "#00af87", "#5f00ff"},
}

func LoadStreamTheme(stream *Stream) {
	if stream.PrefixColors == nil {
		stream.PrefixColors = DefaultStream.PrefixColors
	} else {
		for j := range stream.PrefixColors {
			stream.PrefixColors[j] = *convertToHex(&stream.PrefixColors[j])
		}
	}
}

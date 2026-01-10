package core

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"
)

// Prefixer_Read: Read() with varying line counts and sizes
func BenchmarkPrefixer_Read(b *testing.B) {
	lineCounts := []int{10, 100, 1000}
	lineSizes := []int{50, 200, 500}

	for _, lineCount := range lineCounts {
		for _, lineSize := range lineSizes {
			name := fmt.Sprintf("lines_%d_size_%d", lineCount, lineSize)
			b.Run(name, func(b *testing.B) {
				// Create input with specified number of lines
				var input strings.Builder
				line := strings.Repeat("x", lineSize) + "\n"
				for i := 0; i < lineCount; i++ {
					input.WriteString(line)
				}
				inputStr := input.String()

				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					reader := strings.NewReader(inputStr)
					prefixer := NewPrefixer(reader, "[project-name] ")

					buf := make([]byte, 4096)
					for {
						_, err := prefixer.Read(buf)
						if err == io.EOF {
							break
						}
					}
				}
			},
			)
		}
	}
}

// Prefixer_WriteTo: WriteTo() with varying line counts
func BenchmarkPrefixer_WriteTo(b *testing.B) {
	lineCounts := []int{10, 100, 1000}

	for _, lineCount := range lineCounts {
		name := fmt.Sprintf("lines_%d", lineCount)
		b.Run(name, func(b *testing.B) {
			// Create input with specified number of lines
			var input strings.Builder
			line := strings.Repeat("x", 80) + "\n"
			for i := 0; i < lineCount; i++ {
				input.WriteString(line)
			}
			inputStr := input.String()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				reader := strings.NewReader(inputStr)
				prefixer := NewPrefixer(reader, "[project-name] ")

				var buf bytes.Buffer
				_, _ = prefixer.WriteTo(&buf)
			}
		})
	}
}

// Prefixer_PrefixLen: Impact of prefix length on performance
func BenchmarkPrefixer_PrefixLen(b *testing.B) {
	prefixLengths := []int{10, 50, 100}

	for _, prefixLen := range prefixLengths {
		name := fmt.Sprintf("prefix_%d", prefixLen)
		b.Run(name, func(b *testing.B) {
			// Create input with 100 lines
			var input strings.Builder
			line := strings.Repeat("x", 80) + "\n"
			for i := 0; i < 100; i++ {
				input.WriteString(line)
			}
			inputStr := input.String()
			prefix := strings.Repeat("P", prefixLen) + " "

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				reader := strings.NewReader(inputStr)
				prefixer := NewPrefixer(reader, prefix)

				var buf bytes.Buffer
				_, _ = prefixer.WriteTo(&buf)
			}
		})
	}
}

// Prefixer_Allocs: Memory allocation count (optimization target)
func BenchmarkPrefixer_Allocs(b *testing.B) {
	// Create input with 100 lines
	var input strings.Builder
	line := strings.Repeat("x", 80) + "\n"
	for i := 0; i < 100; i++ {
		input.WriteString(line)
	}
	inputStr := input.String()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := strings.NewReader(inputStr)
		prefixer := NewPrefixer(reader, "[project-name] ")

		var buf bytes.Buffer
		_, _ = prefixer.WriteTo(&buf)
	}
}

package core

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
)

//go:embed mani.1
var ConfigMan []byte

func GenManPages(dir string) error {
	manPath := filepath.Join(dir, "mani.1")
	err := os.WriteFile(manPath, ConfigMan, 0644)
	CheckIfError(err)

	fmt.Printf("Created %s\n", manPath)

	return nil
}

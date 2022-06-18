package core

import (
	_ "embed"
	"fmt"
	"io/ioutil"
	"path/filepath"
)

//go:embed mani.1
var CONFIG_MAN []byte

func GenManPages(dir string) error {
	manPath := filepath.Join(dir, "mani.1")
	err := ioutil.WriteFile(manPath, CONFIG_MAN, 0644)
	CheckIfError(err)

	fmt.Printf("Created %s\n", manPath)

	return nil
}

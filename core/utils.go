package core

import (
	"path/filepath"
	"os"
)

func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func Intersection(a []string, b []string) []string {
	var i []string
	for _, s := range a {
		if StringInSlice(s, b) {
			i = append(i, s)
		}
	}

	return i
}

func FindFileInParentDirs(path string, files []string) (string, error) {
	for _, file := range files {
		pathToFile := filepath.Join(path, file)

		if _, err := os.Stat(pathToFile); err == nil {
			return pathToFile, nil
		}
	}

	parentDir := filepath.Dir(path)

	// TODO: Check different path if on windows subsystem
	// https://stackoverflow.com/questions/151860/root-folder-equivalent-in-windows/152038
	// https://en.wikipedia.org/wiki/Directory_structure#:~:text=In%20DOS%2C%20Windows%2C%20and%20OS,to%20being%20combined%20as%20one.
	// Seems it's \ in windows
	if parentDir == "/" {
		return "", &ConfigNotFound{files}
	}

	return FindFileInParentDirs(parentDir, files)
}

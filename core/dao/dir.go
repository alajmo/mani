package dao

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/alajmo/mani/core"
)

type Dir struct {
	Name        string   `yaml:"name"`
	Path        string   `yaml:"path"`
	Description string   `yaml:"description"`
	Tags        []string `yaml:"tags"`

	RelPath string
}

func (d Dir) GetValue(key string) string {
	switch key {
	case "Name", "name":
		return d.Name
	case "Path", "path":
		return d.Path
	case "RelPath", "relpath":
		return d.RelPath
	case "Description", "description":
		return d.Description
	case "Tags", "tags":
		return strings.Join(d.Tags, ", ")
	}

	return ""
}

func GetDirRelPath(configPath string, path string) (string, error) {
	baseDir := filepath.Dir(configPath)
	relPath, err := filepath.Rel(baseDir, path)

	return relPath, err
}

func (c Config) FilterDirs(
	cwdFlag bool,
	allDirsFlag bool,
	dirPathsFlag []string,
	dirsFlag []string,
	tagsFlag []string,
) []Dir {
	var finalDirs []Dir
	if allDirsFlag {
		finalDirs = c.Dirs
	} else {
		var dirPaths []Dir
		if len(dirPathsFlag) > 0 {
			dirPaths = c.GetDirsByPath(dirPathsFlag)
		}

		var tagDirs []Dir
		if len(tagsFlag) > 0 {
			tagDirs = c.GetDirsByTags(tagsFlag)
		}

		var dirs []Dir
		if len(dirsFlag) > 0 {
			dirs = c.GetDirs(dirsFlag)
		}

		var cwdDir Dir
		if cwdFlag {
			cwdDir = c.GetCwdDir()
		}

		finalDirs = GetUnionDirs(dirPaths, tagDirs, dirs, cwdDir)
	}

	return finalDirs
}

// Dirs must have all paths to match. For instance, if --tags frontend,backend
// is passed, then a dir must have both tags.
func (c Config) GetDirsByPath(drs []string) []Dir {
	if len(drs) == 0 {
		return c.Dirs
	}

	var dirs []Dir
	for _, dir := range c.Dirs {

		// Variable use to check that all dirs are matched
		var numMatched int = 0
		for _, d := range drs {
			if strings.Contains(dir.RelPath, d) {
				numMatched = numMatched + 1
			}
		}

		if numMatched == len(drs) {
			dirs = append(dirs, dir)
		}
	}

	return dirs
}

func (c Config) GetDirs(flagDir []string) []Dir {
	var dirs []Dir

	for _, v := range flagDir {
		for _, d := range c.Dirs {
			if v == d.Name {
				dirs = append(dirs, d)
			}
		}
	}

	return dirs
}

func (c Config) GetCwdDir() Dir {
	cwd, err := os.Getwd()
	core.CheckIfError(err)

	var dir Dir
	parts := strings.Split(cwd, string(os.PathSeparator))

out:
	for i := len(parts) - 1; i >= 0; i-- {
		p := strings.Join(parts[0:i+1], string(os.PathSeparator))

		for _, pro := range c.Dirs {
			if p == pro.Path {
				dir = pro
				break out
			}
		}
	}

	return dir
}

func GetUnionDirs(a []Dir, b []Dir, c []Dir, d Dir) []Dir {
	drs := []Dir{}

	for _, dir := range a {
		if !DirInSlice(dir.Path, drs) {
			drs = append(drs, dir)
		}
	}

	for _, dir := range b {
		if !DirInSlice(dir.Path, drs) {
			drs = append(drs, dir)
		}
	}

	for _, dir := range c {
		if !DirInSlice(dir.Path, drs) {
			drs = append(drs, dir)
		}
	}

	if d.Name != "" {
		drs = append(drs, d)
	}

	dirs := []Dir{}
	dirs = append(dirs, drs...)

	return dirs
}

func DirInSlice(name string, list []Dir) bool {
	for _, d := range list {
		if d.Name == name {
			return true
		}
	}
	return false
}

func (c Config) GetDirNames() []string {
	names := []string{}
	for _, dir := range c.Dirs {
		names = append(names, dir.Name)
	}

	return names
}

/**
 * For each project path, get all the enumerations of dirnames.
 * Example:
 * Input:
 *   - /frontend/tools/project-a
 *   - /frontend/tools/project-b
 *   - /frontend/tools/node/project-c
 *   - /backend/project-d
 * Output:
 *   - /frontend
 *   - /frontend/tools
 *   - /frontend/tools/node
 *   - /backend
 */
func (c Config) GetDirPaths() []string {
	dirs := []string{}
	for _, dir := range c.Dirs {

		ps := strings.Split(filepath.Dir(dir.RelPath), string(os.PathSeparator))
		for i := 1; i <= len(ps); i++ {
			p := filepath.Join(ps[0:i]...)

			if p != "." && !core.StringInSlice(p, dirs) {
				dirs = append(dirs, p)
			}
		}
	}

	return dirs
}

func GetIntersectDirs(a []Dir, b []Dir) []Dir {
	dirs := []Dir{}

	for _, pa := range a {
		for _, pb := range b {
			if pa.Name == pb.Name {
				dirs = append(dirs, pa)
			}
		}
	}

	return dirs
}

func (c Config) GetDirsByName(names []string) []Dir {
	if len(names) == 0 {
		return c.Dirs
	}

	var filtered []Dir
	var found []string
	for _, name := range names {
		if core.StringInSlice(name, found) {
			continue
		}

		for _, dir := range c.Dirs {
			if name == dir.Name {
				filtered = append(filtered, dir)
				found = append(found, name)
			}
		}
	}

	return filtered
}

// Dirs must have all tags to match. For instance, if --tags frontend,backend
// is passed, then a dir must have both tags.
func (c Config) GetDirsByTags(tags []string) []Dir {
	if len(tags) == 0 {
		return c.Dirs
	}

	var dirs []Dir
	for _, dir := range c.Dirs {
		// Variable use to check that all tags are matched
		var numMatched int = 0
		for _, tag := range tags {
			for _, dirTag := range dir.Tags {
				if dirTag == tag {
					numMatched = numMatched + 1
				}
			}
		}

		if numMatched == len(tags) {
			dirs = append(dirs, dir)
		}
	}

	return dirs
}

func (c Config) GetDirsTree(drs []string, tags []string) []core.TreeNode {
	var tree []core.TreeNode
	var paths = []string{}

	dirPaths := c.GetDirsByPath(drs)
	dirTags := c.GetDirsByTags(tags)
	dirs := GetIntersectDirs(dirPaths, dirTags)

	for _, p := range dirs {
		if p.RelPath != "." {
			paths = append(paths, p.RelPath)
		}
	}

	for i := range paths {
		tree = core.AddToTree(tree, strings.Split(paths[i], string(os.PathSeparator)))
	}

	return tree
}

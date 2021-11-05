package dao

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/alajmo/mani/core"
)

type Dir struct {
	Name string   `yaml:"name"`
	Path string   `yaml:"path"`
	Desc string   `yaml:"desc"`
	Tags []string `yaml:"tags"`

	Context string
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
	case "Desc", "desc", "Description", "description":
		return d.Desc
	case "Tags", "tags":
		return strings.Join(d.Tags, ", ")
	}

	return ""
}

func (c *Config) GetDirList() []Dir {
	var dirs []Dir
	count := len(c.Dirs.Content)

	var err error
	for i := 0; i < count; i += 2 {
		dir := &Dir{}
		err = c.Dirs.Content[i+1].Decode(dir)
		core.CheckIfError(err)

		dir.Name = c.Dirs.Content[i].Value

		// Add absolute and relative path for each dir
		var err error
		dir.Path, err = core.GetAbsolutePath(c.Dir, dir.Path, dir.Name)
		core.CheckIfError(err)

		dir.RelPath, err = core.GetRelativePath(c.Dir, dir.Path)
		core.CheckIfError(err)

		dir.Context = c.Path

		dirs = append(dirs, *dir)
	}

	return dirs
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
		finalDirs = c.DirList
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

// DirList must have all paths to match. For instance, if --paths frontend,backend
// is passed, then a dir must have both paths.
func (c Config) GetDirsByPath(drs []string) []Dir {
	if len(drs) == 0 {
		return c.DirList
	}

	var dirs []Dir
	for _, dir := range c.DirList {

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
		for _, d := range c.DirList {
			if v == d.Name {
				dirs = append(dirs, d)
			}
		}
	}

	return dirs
}

func (c Config) GetDir(name string) (*Dir, error) {
	for _, dir := range c.DirList {
		if name == dir.Name {
			return &dir, nil
		}
	}

	return nil, &core.DirNotFound{Name: name}
}

func (c Config) GetCwdDir() Dir {
	cwd, err := os.Getwd()
	core.CheckIfError(err)

	var dir Dir
	parts := strings.Split(cwd, string(os.PathSeparator))

out:
	for i := len(parts) - 1; i >= 0; i-- {
		p := strings.Join(parts[0:i+1], string(os.PathSeparator))

		for _, pro := range c.DirList {
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

	if d.Name != "" && !DirInSlice(d.Name, drs) {
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
	for _, dir := range c.DirList {
		names = append(names, dir.Name)
	}

	return names
}

/**
 * For each project path, get all the enumerations of dirnames. Skip absolute/relative paths
 * outside of mani.yaml
 * Example:
 * Input:
 *   - /frontend/tools/project-a
 *   - /frontend/tools/project-b
 *   - /frontend/tools/node/project-c
 *   - /backend/project-d
 *
 * Output:
 *   - /frontend
 *   - /frontend/tools
 *   - /frontend/tools/node
 *   - /backend
 */
func (c Config) GetDirPaths() []string {
	dirs := []string{}
	for _, dir := range c.DirList {
		// Ignore dirs outside of mani.yaml directory
		if strings.Contains(dir.Path, c.Dir) {
			ps := strings.Split(filepath.Dir(dir.RelPath), string(os.PathSeparator))
			for i := 1; i <= len(ps); i++ {
				p := filepath.Join(ps[0:i]...)

				if p != "." && !core.StringInSlice(p, dirs) {
					dirs = append(dirs, p)
				}
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
		return c.DirList
	}

	var filtered []Dir
	var found []string
	for _, name := range names {
		if core.StringInSlice(name, found) {
			continue
		}

		for _, dir := range c.DirList {
			if name == dir.Name {
				filtered = append(filtered, dir)
				found = append(found, name)
			}
		}
	}

	return filtered
}

// DirList must have all tags to match. For instance, if --tags frontend,backend
// is passed, then a dir must have both tags.
func (c Config) GetDirsByTags(tags []string) []Dir {
	if len(tags) == 0 {
		return c.DirList
	}

	var dirs []Dir
	for _, dir := range c.DirList {
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
	dirPaths := c.GetDirsByPath(drs)
	dirTags := c.GetDirsByTags(tags)
	dirs := GetIntersectDirs(dirPaths, dirTags)

	var tree []core.TreeNode
	var paths = []string{}
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

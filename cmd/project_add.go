package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
)

func projectAddCmd(config *dao.Config, configErr *error) *cobra.Command {
	var name, path, url, desc, tags string

	cmd := cobra.Command{
		Use:   "add <name|url>",
		Short: "Add a project to mani.yaml",
		Long: `Add a project to mani.yaml.

A repository URL is required — either as the argument or via --url:

  - A git URL (https://…, ssh://…, git@host:org/repo.git) is recorded as the
    project's url and the project name is derived from the repo (the final path
    segment, minus ".git"). Override the derived name with --name.
  - A plain name needs --url to supply the repository.

The project is appended to the 'projects:' mapping and cloned on the next
'mani sync'. Comments and formatting elsewhere in mani.yaml are preserved.`,

		Example: `  # Add from a URL — name is derived (-> "mani")
  mani project add https://github.com/alajmo/mani.git

  # Add from a URL but choose the name
  mani project add https://github.com/alajmo/mani.git --name kaka

  # Add by name with an explicit url and tags
  mani project add my-repo --url git@github.com:org/my-repo.git --tags backend

  # Add a remote, cloned to a custom path with a description
  mani project add my-lib --url git@github.com:org/my-lib.git --path ./libs/my-lib --desc "shared lib"`,

		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			core.CheckIfError(*configErr)

			arg := args[0]
			var p dao.Project

			if dao.IsRemoteURL(arg) {
				if url != "" {
					core.CheckIfError(fmt.Errorf("url given both as the argument and --url; pass just one"))
				}
				p.URL = arg
				if name != "" {
					p.Name = name
				} else {
					p.Name = dao.ProjectNameFromURL(arg)
				}
				if p.Name == "" {
					core.CheckIfError(fmt.Errorf("could not derive a project name from %q; pass --name", arg))
				}
			} else {
				if name != "" && name != arg {
					core.CheckIfError(fmt.Errorf("project name given twice: %q (argument) and %q (--name)", arg, name))
				}
				p.Name = arg
				p.URL = url
			}

			if p.URL == "" {
				core.CheckIfError(fmt.Errorf("a project url is required — pass a repository URL as the argument, or a name with --url"))
			}

			p.Path = p.Name
			if path != "" {
				p.Path = path
			}
			p.Desc = desc
			if tags != "" {
				p.Tags = splitTags(tags)
			}

			if _, err := config.GetProject(p.Name); err == nil {
				core.CheckIfError(&core.ProjectAlreadyExists{Name: p.Name})
			}

			core.CheckIfError(dao.AddProject(config.Path, p))

			fmt.Printf("Added project `%s` to %s\n", p.Name, config.Path)
		},
		DisableAutoGenTag: true,
	}

	cmd.Flags().StringVar(&name, "name", "", "project name (defaults to the repo name derived from a URL argument)")
	cmd.Flags().StringVar(&path, "path", "", "local directory path relative to mani.yaml (defaults to the project name)")
	cmd.Flags().StringVar(&url, "url", "", "git repository URL (cloned on next sync)")
	cmd.Flags().StringVar(&desc, "desc", "", "project description")
	cmd.Flags().StringVar(&tags, "tags", "", "comma-separated tags")

	return &cmd
}

// splitTags parses a comma-separated --tags value into a trimmed,
// empty-skipping slice.
func splitTags(s string) []string {
	var out []string
	for _, t := range strings.Split(s, ",") {
		if t = strings.TrimSpace(t); t != "" {
			out = append(out, t)
		}
	}
	return out
}

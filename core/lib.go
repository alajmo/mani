package core

func FilterProjectOnName(projects []Project, names []string) []Project {
	if len(names) == 0 {
		return projects
	}

	var filteredProjects []Project
	var foundProjectNames []string
	for _, name := range names {
		if StringInSlice(name, foundProjectNames) {
			continue
		}

		for _, project := range projects {
			if name == project.Name {
				filteredProjects = append(filteredProjects, project)
				foundProjectNames = append(foundProjectNames, name)
			}
		}
	}

	return filteredProjects
}

func FilterCommandOnName(commands []Command, names []string) []Command {
	if len(names) == 0 {
		return commands
	}

	var filteredCommands []Command
	var foundCommands []string
	for _, name := range names {
		if StringInSlice(name, foundCommands) {
			continue
		}

		for _, project := range commands {
			if name == project.Name {
				filteredCommands = append(filteredCommands, project)
				foundCommands = append(foundCommands, name)
			}
		}
	}

	return filteredCommands
}

// Projects must have all tags to match.
func FilterProjectOnTag(projects []Project, tags []string) []Project {
	if len(tags) == 0 {
		return projects
	}

	var filteredProjects []Project
	for _, project := range projects {
		var foundTags int = 0
		for _, tag := range tags {
			for _, projectTag := range project.Tags {
				if projectTag == tag {
					foundTags = foundTags + 1
				}
			}
		}

		if foundTags == len(tags) {
			filteredProjects = append(filteredProjects, project)
		}
	}

	return filteredProjects
}

func FilterTagOnProject(projects []Project, projectNames []string) []string {
	tags := []string{}
	for _, project := range projects {
		if StringInSlice(project.Name, projectNames) {
			tags = append(tags, project.Tags...)
		}
	}

	return tags
}

func GetProjectNames(projects []Project) []string {
	projectNames := []string{}
	for _, project := range projects {
		projectNames = append(projectNames, project.Name)
	}

	return projectNames
}

func GetCommandNames(commands []Command) []string {
	commandNames := []string{}
	for _, project := range commands {
		commandNames = append(commandNames, project.Name)
	}

	return commandNames
}

func GetTags(projects []Project) []string {
	tags := []string{}
	for _, project := range projects {
		for _, tag := range project.Tags {
			if !StringInSlice(tag, tags) {
				tags = append(tags, tag)
			}
		}
	}

	return tags
}

func GetProjectUrls(projects []Project) []string {
	urls := []string{}
	for _, project := range projects {
		if (project.Url != "") {
			urls = append(urls, project.Url)
		}
	}

	return urls
}

func ProjectInSlice(name string, list []Project) bool {
	for _, p := range list {
		if p.Name == name {
			return true
		}
	}
	return false
}

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

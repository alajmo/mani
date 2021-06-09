package core

func FilterProjectOnTag(projects []Project, tags []string) []Project {
	var filteredProjects []Project
	for _, project := range projects {
		if len(tags) == 0 {
			filteredProjects = append(filteredProjects, project)
			continue
		}

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
			for _, tag := range project.Tags {
				tags = append(tags, tag)
			}
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

func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func ProjectInSlice(name string, list []Project) bool {
	for _, p := range list {
		if p.Name == name {
			return true
		}
	}
	return false
}

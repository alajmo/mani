package core

func GetAllTags(projects []Project) map[string]struct{} {
	tags := make(map[string]struct{})
	for _, project := range projects {
		for _, tag := range project.Tags {
			tags[tag] = struct{}{}
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

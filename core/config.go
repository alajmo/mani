package core

type Project struct {
	Name        string   `yaml:"name"`
	Path        string   `yaml:"path"`
	Description string   `yaml:"description"`
	Url         string   `yaml:"url"`
	Tags        []string `yaml:"tags"`
}

type Command struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Args        map[string]string `yaml:"args"`
	Shell		string            `yaml:"shell"`
	Command     string            `yaml:"command"`
}

type Config struct {
	Shell    string    `yaml:"shell"`
	Projects []Project `yaml:"projects"`
	Commands []Command `yaml:"commands"`
}

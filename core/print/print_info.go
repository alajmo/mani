package print

import (
	"github.com/alajmo/mani/core"
	"fmt"
)

func PrintInfo(configPath string, config core.Config) {
	if configPath != "" {
		tags := core.GetAllProjectTags(config.Projects)

		fmt.Printf("context %s\n", configPath)
		fmt.Printf("%d projects\n", len(config.Projects))
		fmt.Printf("%d commands\n", len(config.Commands))
		fmt.Printf("%d tags\n\n", len(tags))
	}

	fmt.Printf("mani version %s\n", version)
	cmd := exec.Command("git", "--version")
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("git not installed")
	} else {
		fmt.Println(string(stdout))
	}
}

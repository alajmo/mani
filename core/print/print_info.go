package print

import (
	"github.com/alajmo/mani/core"
	"fmt"
	"strings"
	"os/exec"
)

func PrintInfo(configPath string, config core.Config) {
	if configPath != "" {
		fmt.Printf("context: %s\n", configPath)
		fmt.Printf("shell: %v\n", strings.Split(config.Shell, " ")[0])
	}

	fmt.Printf("mani version %s\n", core.Version)
	cmd := exec.Command("git", "--version")
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("git not installed")
	} else {
		fmt.Println(string(stdout))
	}
}

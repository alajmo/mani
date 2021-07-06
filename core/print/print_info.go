package print

import (
	"fmt"
	"strings"
	"os/exec"

	"github.com/alajmo/mani/core/dao"
)

func PrintInfo(config *dao.Config) {
	if config.Path != "" {
		fmt.Printf("context: %s\n", config.Path)
		fmt.Printf("shell: %v\n", strings.Split(config.Shell, " ")[0])
	}

	fmt.Printf("mani version %s\n", dao.Version)
	cmd := exec.Command("git", "--version")
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("git not installed")
	} else {
		fmt.Println(string(stdout))
	}
}

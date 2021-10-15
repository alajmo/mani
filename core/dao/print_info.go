package dao

import (
	"fmt"
	"os/exec"
)

func PrintInfo(config *Config) {
	if config.Path != "" {
		fmt.Printf("config: %s\n", config.Path)
	}

	fmt.Printf("mani version %s\n", Version)
	cmd := exec.Command("git", "--version")
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("git not installed")
	} else {
		fmt.Println(string(stdout))
	}
}

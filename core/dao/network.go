package dao

import (
	"fmt"
	"strings"
	"os"
	"os/exec"
	"io/ioutil"

	"gopkg.in/yaml.v3"

	"github.com/alajmo/mani/core"
)

type Network struct {
	Name		string
	Description string    `yaml:"description"`
	User		string    `yaml:"user"`
	Hosts		[]string  `yaml:"hosts"`
	Tags        []string  `yaml:"tags"`
}

func (n Network) GetValue(key string) string {
	switch key {
	case "Name", "name":
		return n.Name
	case "Description", "description":
		return n.Description
	case "User", "user":
		return n.User
	case "Hosts", "hosts":
		return strings.Join(n.Hosts, "\n")
	case "Tags", "tags":
		return strings.Join(n.Tags, ", ")
	}

	return ""
}

func (c *Config) SetNetworkList() []Network {
	var networks []Network
	count := len(c.Networks.Content)

	for i := 0; i < count; i += 2 {
		network := &Network{}
		c.Networks.Content[i + 1].Decode(network)
		network.Name = c.Networks.Content[i].Value
		networks = append(networks, *network)
	}

	c.NetworkList = networks

	return networks
}

func (c Config) GetNetworkNames() []string {
	names := []string{}
	for _, network := range c.NetworkList {
		names = append(names, network.Name)
	}

	return names
}

func (c Config) GetNetworksByName(names []string) []Network {
	if len(names) == 0 {
		return c.NetworkList
	}

	var filtered []Network
	var found []string
	for _, name := range names {
		if core.StringInSlice(name, found) {
			continue
		}

		for _, network := range c.NetworkList {
			if name == network.Name {
				filtered = append(filtered, network)
				found = append(found, name)
			}
		}
	}

	return filtered
}

func (c Config) GetNetworksByHost(hosts []string) []Network {
	if len(hosts) == 0 {
		return c.NetworkList
	}

	var networks []Network
	for _, network := range c.NetworkList {
		// Variable use to check that all hosts are matched
		var numMatched int = 0
		for _, host := range hosts {
			for _, networkHost := range network.Hosts {
				if networkHost == host {
					numMatched = numMatched + 1
				}
			}
		}

		if numMatched == len(hosts) {
			networks = append(networks, network)
		}
	}

	return networks
}

// Networks must have all tags to match. For instance, if --tags frontend,backend
// is passed, then a network must have both tags.
func (c Config) GetNetworksByTag(tags []string) []Network {
	if len(tags) == 0 {
		return c.NetworkList
	}

	var networks []Network
	for _, network := range c.NetworkList {
		// Variable use to check that all tags are matched
		var numMatched int = 0
		for _, tag := range tags {
			for _, dirTag := range network.Tags {
				if dirTag == tag {
					numMatched = numMatched + 1
				}
			}
		}

		if numMatched == len(tags) {
			networks = append(networks, network)
		}
	}

	return networks
}

func (c Config) FilterNetworks(
    allNetworksFlag bool,
    networksFlag []string,
    networkHostsFlag []string,
    tagsFlag []string,
) []Network {
	var networks []Network
	if allNetworksFlag {
		networks = c.NetworkList
	} else {
		var networksName []Network
		if len(networksFlag) > 0 {
			networksName = c.GetNetworksByName(networksFlag)
		}

		var networksTag []Network
		if len(tagsFlag) > 0 {
			networksTag = c.GetNetworksByTag(tagsFlag)
		}

		var networksHost []Network
		if len(networkHostsFlag) > 0 {
			networksHost = c.GetNetworksByHost(networkHostsFlag)
		}

		networks = GetUnionNetworks(networksName, networksTag, networksHost)
	}

	return networks
}


func GetIntersectNetworks(a []Network, b []Network) []Network {
	networks := []Network{}

	for _, pa := range a {
		for _, pb := range b {
			if (pa.Name == pb.Name) {
				networks = append(networks, pa)
			}
		}
	}

	return networks
}

func (c Config) GetAllHosts() []string {
	var hosts []string
	for _, network := range c.NetworkList {
		for _, host := range network.Hosts {
			hosts = append(hosts, host)
		}
	}

	return hosts
}

// Open mani config in editor and optionally go to line matching the network name
func (c Config) EditNetworks(networkName string) {
	dat, err := ioutil.ReadFile(c.Path)
	core.CheckIfError(err)

	type ConfigTmp struct {
		Networks yaml.Node
	}

	var configTmp ConfigTmp
	err = yaml.Unmarshal([]byte(dat), &configTmp)
	core.CheckIfError(err)

	lineNr := 0
	if networkName == "" {
		lineNr = configTmp.Networks.Line - 1
	} else {
		out:
		for _, network := range configTmp.Networks.Content {
			if network.Value == networkName {
				lineNr = network.Line
				break out
			}
		}
	}

	editor := os.Getenv("EDITOR")
	var args []string
	switch editor {
	case "vim":
		args = []string{fmt.Sprintf("+%v", lineNr), c.Path}
	case "vi":
		args = []string{fmt.Sprintf("+%v", lineNr), c.Path}
	case "emacs":
		args = []string{fmt.Sprintf("+%v", lineNr), c.Path}
	case "nano":
		args = []string{fmt.Sprintf("+%v", lineNr), c.Path}
	case "code": // visual studio code
		args = []string{"--goto", fmt.Sprintf("%s:%v", c.Path, lineNr)}
	case "idea": // Intellij
		args = []string{"--line", fmt.Sprintf("%v", lineNr), c.Path}
	case "subl": // Sublime
		args = []string{fmt.Sprintf("%s:%v", c.Path, lineNr)}
	case "atom":
		args = []string{fmt.Sprintf("%s:%v", c.Path, lineNr)}
	case "notepad-plus-plus":
		args = []string{"-n", fmt.Sprintf("%v", lineNr), c.Path}
	default:
		args = []string{c.Path}
	}

	cmd := exec.Command(editor, args...)
	cmd.Env = os.Environ()
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
    err = cmd.Run()
	core.CheckIfError(err)
}

func GetUnionNetworks(a []Network, b []Network, c []Network) []Network {
	networks := []Network{}

	for _, network := range a {
		if !NetworkInSlice(network.Name, networks) {
			networks = append(networks, network)
		}
	}

	for _, network := range b {
		if !NetworkInSlice(network.Name, networks) {
			networks = append(networks, network)
		}
	}

	for _, network := range c {
		if !NetworkInSlice(network.Name, networks) {
			networks = append(networks, network)
		}
	}

	return networks
}

func NetworkInSlice(name string, list []Network) bool {
	for _, d := range list {
		if d.Name == name {
			return true
		}
	}
	return false
}


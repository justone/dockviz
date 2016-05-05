package main

import (
	"github.com/fsouza/go-dockerclient"

	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

type Container struct {
	Id      string
	Image   string
	Names   []string
	Ports   []map[string]interface{}
	Created int64
	Status  string
	Command string
}

type ContainersCommand struct {
	Dot        bool `short:"d" long:"dot" description:"Show container information as Graphviz dot."`
	NoTruncate bool `short:"n" long:"no-trunc" description:"Don't truncate the container IDs."`
}

var containersCommand ContainersCommand

func (x *ContainersCommand) Execute(args []string) error {

	var containers *[]Container

	stat, err := os.Stdin.Stat()
	if err != nil {
		return fmt.Errorf("error reading stdin stat", err)
	}

	if (stat.Mode() & os.ModeCharDevice) == 0 {
		// read in stdin
		stdin, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("error reading all input", err)
		}

		containers, err = parseContainersJSON(stdin)
		if err != nil {
			return err
		}
	} else {

		client, err := connect()
		if err != nil {
			return err
		}

		clientContainers, err := client.ListContainers(docker.ListContainersOptions{All: true})
		if err != nil {
			if in_docker := os.Getenv("IN_DOCKER"); len(in_docker) > 0 {
				return fmt.Errorf("Unable to access Docker socket, please run like this:\n  docker run --rm -v /var/run/docker.sock:/var/run/docker.sock nate/dockviz containers <args>\nFor more help, run 'dockviz help'")
			} else {
				return fmt.Errorf("Unable to connect: %s\nFor help, run 'dockviz help'", err)
			}
		}

		var conts []Container
		for _, container := range clientContainers {
			conts = append(conts, Container{
				container.ID,
				container.Image,
				container.Names,
				apiPortToMap(container.Ports),
				container.Created,
				container.Status,
				container.Command,
			})
		}

		containers = &conts
	}

	if containersCommand.Dot {
		fmt.Printf(jsonContainersToDot(containers))
	} else {
		return fmt.Errorf("Please specify --dot")
	}

	return nil
}

func apiPortToMap(ports []docker.APIPort) []map[string]interface{} {
	result := make([]map[string]interface{}, 2)
	for _, port := range ports {
		intPort := map[string]interface{}{
			"IP":          port.IP,
			"Type":        port.Type,
			"PrivatePort": port.PrivatePort,
			"PublicPort":  port.PublicPort,
		}
		result = append(result, intPort)
	}
	return result
}

func parseContainersJSON(rawJSON []byte) (*[]Container, error) {

	var containers []Container
	err := json.Unmarshal(rawJSON, &containers)

	if err != nil {
		return nil, fmt.Errorf("Error reading JSON: ", err)
	}

	return &containers, nil
}

func jsonContainersToDot(containers *[]Container) string {

	var buffer bytes.Buffer
	buffer.WriteString("digraph docker {\n")

	for _, container := range *containers {

		var containerName string

		for _, name := range container.Names {
			if strings.Count(name, "/") == 1 {
				containerName = name[1:]
			}
		}
		for _, name := range container.Names {
			nameParts := strings.Split(name, "/")
			if len(nameParts) > 2 {
				buffer.WriteString(fmt.Sprintf(" \"%s\" -> \"%s\" [label = \" %s\" ]\n", containerName, nameParts[1], nameParts[len(nameParts)-1]))

			}
		}

		var containerBackground string
		if strings.Contains(container.Status, "Exited") {
			containerBackground = "lightgrey"
		} else {
			containerBackground = "paleturquoise"
		}

		buffer.WriteString(fmt.Sprintf(" \"%s\" [label=\"%s\\n%s\",shape=box,fillcolor=\"%s\",style=\"filled,rounded\"];\n", containerName, containerName, truncate(container.Id, 12), containerBackground))
	}

	buffer.WriteString("}\n")

	return buffer.String()
}

func init() {
	parser.AddCommand("containers",
		"Visualize docker containers.",
		"",
		&containersCommand)
}

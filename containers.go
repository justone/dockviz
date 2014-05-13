package main

import (
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

	// read in stdin
	stdin, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("error reading all input", err)
	}

	containers, err := parseContainersJSON(stdin)
	if err != nil {
		return err
	}

	if containersCommand.Dot {
		fmt.Printf(jsonContainersToDot(containers))
	} else {
		return fmt.Errorf("Please specify --dot")
	}

	return nil
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

		buffer.WriteString(fmt.Sprintf(" \"%s\" [label=\"%s\\n%s\",shape=box,fillcolor=\"%s\",style=\"filled,rounded\"];\n", containerName, containerName, truncate(container.Id), containerBackground))
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

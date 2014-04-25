package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

type Image struct {
	Id          string
	ParentId    string   `json:",omitempty"`
	RepoTags    []string `json:",omitempty"`
	VirtualSize int64
	Size        int64
	Created     int64
}

type ImagesCommand struct {
	Dot  bool `short:"d" long:"dot" description:"Show image information as Graphviz dot."`
	Tree bool `short:"t" long:"tree" description:"Show image information as tree."`
}

var imagesCommand ImagesCommand

func (x *ImagesCommand) Execute(args []string) error {

	// read in stdin
	stdin, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("error reading all input", err)
	}

	images, err := parseJSON(stdin)

	if imagesCommand.Dot {
		fmt.Printf(jsonToDot(images))
	} else if imagesCommand.Tree {
		fmt.Println("Tree output not implemented yet.")
	}

	return nil
}

func truncate(id string) string {
	return id[0:12]
}

func parseJSON(rawJSON []byte) (*[]Image, error) {

	var images []Image
	err := json.Unmarshal(rawJSON, &images)

	if err != nil {
		return nil, fmt.Errorf("Error reading JSON: ", err)
	}

	return &images, nil
}

func jsonToDot(images *[]Image) string {

	var buffer bytes.Buffer
	buffer.WriteString("digraph docker {\n")

	for _, image := range *images {
		if image.ParentId == "" {
			buffer.WriteString(fmt.Sprintf(" base -> \"%s\" [style=invis]\n", truncate(image.Id)))
		} else {
			buffer.WriteString(fmt.Sprintf(" \"%s\" -> \"%s\"\n", truncate(image.ParentId), truncate(image.Id)))
		}
		if image.RepoTags[0] != "<none>:<none>" {
			buffer.WriteString(fmt.Sprintf(" \"%s\" [label=\"%s\\n%s\",shape=box,fillcolor=\"paleturquoise\",style=\"filled,rounded\"];\n", truncate(image.Id), truncate(image.Id), strings.Join(image.RepoTags, "\\n")))
		}
	}

	buffer.WriteString(" base [style=invisible]\n}\n")

	return buffer.String()
}

func init() {
	parser.AddCommand("images",
		"Visualize docker images.",
		"",
		&imagesCommand)
}

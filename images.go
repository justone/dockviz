package main

import "fmt"

type ImagesCommand struct {
	Dot  bool `short:"d" long:"dot" description:"Show image information as Graphviz dot."`
	Tree bool `short:"t" long:"tree" description:"Show image information as tree."`
}

var imagesCommand ImagesCommand

func (x *ImagesCommand) Execute(args []string) error {

	if imagesCommand.Dot {
		fmt.Println("Output dot")
	} else if imagesCommand.Tree {
		fmt.Println("Output tree")
	}

	return nil
}

func init() {
	parser.AddCommand("images",
		"Visualize docker images.",
		"",
		&imagesCommand)
}

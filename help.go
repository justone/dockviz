package main

import "fmt"

type HelpCommand struct {
	// nothing yet
}

var helpCommand HelpCommand

func (x *HelpCommand) Execute(args []string) error {
	fmt.Println("stub help is here")

	return nil
}

func init() {
	parser.AddCommand("help",
		"Help for this tool.",
		"",
		&helpCommand)
}

package main

import "fmt"

type HelpCommand struct {
	// nothing yet
}

var helpCommand HelpCommand

func (x *HelpCommand) Execute(args []string) error {
	fmt.Println(`Dockviz: Visualizing Docker Data

Connecting to Docker:

Dockviz supports connecting to the Docker daemon directly.  It defaults to
'unix:///var/run/docker.sock', but respects the following as well:

* The 'DOCKER_HOST', 'DOCKER_CERT_PATH', and 'DOCKER_TLS_VERIFY' environment
  variables, as set up by boot2docker or docker-machine.
* Command line arguments (e.g. '--tlscacert'), like those that Docker itself
  supports.

Dockviz also supports receiving Docker image or container json data on standard
input: curl -s http://localhost:4243/images/json?all=1 | dockviz images --tree

Running inside a Docker image:

Dockviz can be run as a Docker image, like this:

$ docker run -it --rm -v /var/run/docker.sock:/var/run/docker.sock nate/dockviz

If you get an error that Dockviz is unable to access the Docker socket, even
when specified, SELinux may be getting in the way.  See this issue comment for
more information:

https://github.com/justone/dockviz/issues/12#issuecomment-186682489

Visualizing:

Dockviz can visualize both images and containers. For more information on the
options each subcommand supports, run them with the '--help' flag (e.g.
'dockviz images --help').
`)

	return nil
}

func init() {
	parser.AddCommand("help",
		"Help for dockviz.",
		"",
		&helpCommand)
}

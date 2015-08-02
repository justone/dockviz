package main // import "github.com/justone/dockviz"

import (
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
)

type GlobalOptions struct {
	TLSCaCert string `long:"tlscacert" value-name:"~/.docker/ca.pem" description:"Trust certs signed only by this CA"`
	TLSCert   string `long:"tlscert" value-name:"~/.docker/cert.pem" description:"Path to TLS certificate file"`
	TLSKey    string `long:"tlskey" value-name:"~/.docker/key.pem" description:"Path to TLS key file"`
	TLSVerify bool   `long:"tlsverify" description:"Use TLS and verify the remote"`
	Host      string `long:"host" short:"H" value-name:"unix:///var/run/docker.sock" description:"Docker host to connect to"`
	Version   func() `long:"version" short:"v" description:"Display version information."`
}

var globalOptions GlobalOptions
var parser = flags.NewParser(&globalOptions, flags.Default)

var version = "v0.2.1"

func main() {
	globalOptions.Version = func() {
		fmt.Println("dockviz", version)
		os.Exit(0)
	}
	if _, err := parser.Parse(); err != nil {
		os.Exit(1)
	}
}

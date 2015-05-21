package main

import (
	"github.com/fsouza/go-dockerclient"

	"os"
	"path"
)

func connect() (*docker.Client, error) {

	// grab directly from docker daemon
	endpoint := os.Getenv("DOCKER_HOST")
	if len(endpoint) == 0 {
		endpoint = "unix:///var/run/docker.sock"
	}

	var client *docker.Client
	var err error
	if dockerCertPath := os.Getenv("DOCKER_CERT_PATH"); len(dockerCertPath) > 0 {
		cert := path.Join(dockerCertPath, "cert.pem")
		key := path.Join(dockerCertPath, "key.pem")
		ca := path.Join(dockerCertPath, "ca.pem")
		client, err = docker.NewTLSClient(endpoint, cert, key, ca)
		if err != nil {
			return nil, err
		}
	} else {
		client, err = docker.NewClient(endpoint)
		if err != nil {
			return nil, err
		}
	}
	return client, nil
}

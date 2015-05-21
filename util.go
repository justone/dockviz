package main

import (
	"errors"
	"os"
	"path"

	"github.com/fsouza/go-dockerclient"
)

func connect() (*docker.Client, error) {

	// grab directly from docker daemon
	endpoint := os.Getenv("DOCKER_HOST")
	if len(endpoint) == 0 {
		endpoint = "unix:///var/run/docker.sock"
	}

	var client *docker.Client
	var err error
	dockerTlsVerifyEnv := os.Getenv("DOCKER_TLS_VERIFY")
	if dockerTlsVerifyEnv == "1" || globalOptions.TLSVerify {
		if dockerCertPath := os.Getenv("DOCKER_CERT_PATH"); len(dockerCertPath) > 0 {
			cert := path.Join(dockerCertPath, "cert.pem")
			key := path.Join(dockerCertPath, "key.pem")
			ca := path.Join(dockerCertPath, "ca.pem")
			client, err = docker.NewTLSClient(endpoint, cert, key, ca)
			if err != nil {
				return nil, err
			}
		} else if len(globalOptions.TLSCert) > 0 && len(globalOptions.TLSKey) > 0 && len(globalOptions.TLSCaCert) > 0 {
			client, err = docker.NewTLSClient(endpoint, globalOptions.TLSCert, globalOptions.TLSKey, globalOptions.TLSCaCert)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, errors.New("TLS Verification requested but certs not specified")
		}
	} else {
		client, err = docker.NewClient(endpoint)
		if err != nil {
			return nil, err
		}
	}
	return client, nil
}

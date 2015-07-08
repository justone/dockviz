# dockviz: Visualizing Docker Data

This command takes Docker image and container information and presents in
different ways, to help you understand what's going on inside the system.

# Quick Start

1. Install dockviz.  Either:
  * Download the [latest release](https://github.com/justone/dockviz/releases).
  * Set up an alias to run it from the (5.8 MB) docker image: 

  ```
  alias dockviz="docker run --rm -v /var/run/docker.sock:/var/run/docker.sock nate/dockviz"
  ```
2. Visualize images by running `dockviz images -t`, which has similar output to `docker images -t`.

Image can be visualized as [Graphviz](http://www.graphviz.org), or as a tree or short summary in the terminal.  Only Graphviz output has been implemented for containers.

# Output Examples

## Containers

Currently, containers are visualized with labeled lines for links.  Containers that aren't running are greyed out.

```
$ dockviz containers -d | dot -Tpng -o containers.png
```

![](sample/containers.png "Container")

## Images

Image info is visualized with lines indicating parent images:

```
$ dockviz images -d | dot -Tpng -o images.png
```

![](sample/images.png "Image")

Or in short form:

```
$ dockviz images -s
nate/mongodb: latest
redis: latest
ubuntu: 12.04, precise, 12.10, quantal, 13.04, raring
```

Or as a tree in the terminal:

```
$ dockviz images -t
└─511136ea3c5a Virtual Size: 0.0 B
  |─f10ebce2c0e1 Virtual Size: 103.7 MB
  | └─82cdea7ab5b5 Virtual Size: 103.9 MB
  |   └─5dbd9cb5a02f Virtual Size: 103.9 MB
  |     └─74fe38d11401 Virtual Size: 209.6 MB Tags: ubuntu:12.04, ubuntu:precise
  |─ef519c9ee91a Virtual Size: 100.9 MB
  | └─07302703becc Virtual Size: 101.2 MB
  |   └─cf8dc907452c Virtual Size: 101.2 MB
  |     └─a7cf8ae4e998 Virtual Size: 171.3 MB Tags: ubuntu:12.10, ubuntu:quantal
  |       |─e18d8001204e Virtual Size: 171.3 MB
  |       | └─d0525208a46c Virtual Size: 171.3 MB
  |       |   └─59dac4bae93b Virtual Size: 242.5 MB
  |       |     └─89541b3b35f2 Virtual Size: 511.8 MB
  |       |       └─7dac4e98548e Virtual Size: 511.8 MB
  |       |         └─341d0cc3fac8 Virtual Size: 511.8 MB
  |       |           └─2f96171d2098 Virtual Size: 511.8 MB
  |       |             └─67b8b7262a67 Virtual Size: 513.7 MB
  |       |               └─0fe9a2bc50fe Virtual Size: 513.7 MB
  |       |                 └─8c32832f07ba Virtual Size: 513.7 MB
  |       |                   └─cc4e1358bc80 Virtual Size: 513.7 MB
  |       |                     └─5c0d04fba9df Virtual Size: 513.7 MB Tags: nate/mongodb:latest
  |       └─398d592f2009 Virtual Size: 242.2 MB
  |         └─0cd8e7f50270 Virtual Size: 243.6 MB
  |           └─594b6f8e6f92 Virtual Size: 243.6 MB
  |             └─f832a63e87a4 Virtual Size: 243.6 MB Tags: redis:latest
  └─02dae1c13f51 Virtual Size: 98.3 MB
    └─e7206bfc66aa Virtual Size: 98.5 MB
      └─cb12405ee8fa Virtual Size: 98.5 MB
        └─316b678ddf48 Virtual Size: 169.4 MB Tags: ubuntu:13.04, ubuntu:raring
```

# Running

Dockviz supports connecting to the Docker daemon directly.  It defaults to `unix:///var/run/docker.sock`, but respects the following as well:

* The `DOCKER_HOST`, `DOCKER_CERT_PATH`, and `DOCKER_TLS_VERIFY` environment variables, as set up by [boot2docker](http://boot2docker.io/) or [docker-machine](https://docs.docker.com/machine/).
* Command line arguments (e.g. `--tlscacert`), like those that Docker itself supports.

Dockviz also supports receiving Docker image or container json data on standard input.

```
$ curl -s http://localhost:4243/images/json?all=1 | dockviz images --tree
$ curl -s http://localhost:4243/containers/json?all=1 | dockviz containers --dot | dot -Tpng -o containers.png
$ echo -e "GET /images/json?all=1 HTTP/1.0\r\n" | nc -U /var/run/docker.sock | tail -n +5 | dockviz images --tree
$ echo -e "GET /containers/json?all=1 HTTP/1.0\r\n" | nc -U /var/run/docker.sock | tail -n +5 | dockviz containers --dot | dot -Tpng -o containers.png
```

Note: GNU netcat doesn't support `-U` (UNIX socket) flag, so OpenBSD variant can be used.

# Binaries

See the [releases](https://github.com/justone/dockviz/releases) area for binaries.

# Build

```bash
go get ./...
go build
```

.PHONY: build cross image

build:
	go build

cross:
	gox -osarch="darwin/amd64 linux/amd64 linux/arm"

image:
	docker run --rm -v $(shell pwd):/src -v /var/run/docker.sock:/var/run/docker.sock centurylink/golang-builder
	docker tag dockviz nate/dockviz

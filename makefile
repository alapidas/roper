IMAGE_NAME := roper-build
GO_PREFIX  := godep


# Dockerized targets
docker_image:
	docker build -t $(IMAGE_NAME) hack/

docker_test: docker_image
	docker run --rm -v $(shell pwd):/go/src/github.com/alapidas/roper -w /go/src/github.com/alapidas/roper/ $(IMAGE_NAME) make test

docker_shell: docker_image
	docker run -it --rm -v $(shell pwd):/go/src/github.com/alapidas/roper/ -w /go/src/github.com/alapidas/roper/ $(IMAGE_NAME) bash

docker_run: docker_image
	docker run -it --rm -v $(shell pwd):/go/src/github.com/alapidas/roper/ -w /go/src/github.com/alapidas/roper/ -p 3001:3001 $(IMAGE_NAME) make run


# Regular targets
test:
	$(GO_PREFIX) go test ./...

build:
	$(GO_PREFIX) go build main.go

run: build
	./main serve

# Godep targets
godep_save:
	godep save ./...

godep_restore:
	godep restore

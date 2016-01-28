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
	$(GO_PREFIX) go build -o roper main.go

run: build
	./roper serve

run_bootstrapped: build
	./roper repo add /Users/alapidas/goWorkspace/src/github.com/alapidas/roper/hack/test_repos/epel TestEpel
	./roper repo add /Users/alapidas/goWorkspace/src/github.com/alapidas/roper/hack/test_repos/docker/7 Docker
	./roper serve

# Godep targets
godep_save:
	godep save ./...

godep_restore:
	godep restore

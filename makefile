IMAGE_NAME := roper-build
GO_PREFIX  := godep
MKFILE_DIR  = $(shell pwd)
ROPER_OPTS  =

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
	$(GO_PREFIX) go test -race ./...

build:
	$(GO_PREFIX) go build -race -o roper main.go

run: build
	./roper $(ROPER_OPTS) serve

run_bootstrapped: build
	./roper $(ROPER_OPTS) repo add /Users/alapidas/goWorkspace/src/github.com/alapidas/roper/hack/test_repos/epel TestEpel
	./roper $(ROPER_OPTS) repo add /Users/alapidas/goWorkspace/src/github.com/alapidas/roper/hack/test_repos/docker/7 Docker
	./roper $(ROPER_OPTS) serve

# Godep targets
godep_save:
	godep save ./...

godep_restore:
	godep restore

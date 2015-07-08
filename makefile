IMAGE_NAME := roper-build


# Docker targets
docker_image:
	docker build -t $(IMAGE_NAME) hack/

docker_test: docker_image
	docker run --rm -v $(shell pwd):/go/src/github.com/alapidas/roper/ -w /go/src/github.com/alapidas/roper/ $(IMAGE_NAME) make test

docker_shell: docker_image
	docker run -it --rm -v $(shell pwd):/go/src/github.com/alapidas/roper/ -w /go/src/github.com/alapidas/roper/ $(IMAGE_NAME) bash


# Regular targets
test:
	godep restore
	go test ./...


# Godep targets
godep_save:
	godep save ./...

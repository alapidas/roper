IMAGE_NAME := roper-build
GO_PREFIX  := godep
MKFILE_DIR  = $(shell pwd)
ROPER_OPTS  =

# Regular targets
test:
	$(GO_PREFIX) go test -race ./...

build:
	$(GO_PREFIX) go build -race -o roper main.go

run: build
	./roper $(ROPER_OPTS) serve

# Godep targets
godep_save:
	godep save ./...

godep_restore:
	godep restore

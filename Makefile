SHELL := /bin/bash

all: fmt lint vet test build

build: 
	@go build ./hashtable

test:
	@go test ./...

fmt:
	@gofmt -d ./

lint:
	@golint -set_exit_status $$(go list ./...)

vet:
	@go tool vet ./

SHELL := /bin/bash

all: fmt lint vet test build

build: 
	@go build ./cache

test:
	@go test -race -coverprofile=coverage.txt -covermode=atomic ./...

fmt:
	@gofmt -d -s ./

lint:
	@golint -set_exit_status $$(go list ./...)

vet:
	@go tool vet ./

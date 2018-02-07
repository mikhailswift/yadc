SHELL := /bin/bash

all: fmt lint vet dep test build

build: 
	@go build ./cache
	@go build ./commands

test:
	@./tests.sh
fmt:
	@gofmt -d -s ./

lint:
	@golint -set_exit_status $$(go list ./...)

vet:
	@go tool vet ./

dep:
	@go get -d github.com/mikhailswift/yadc/cache
	@go get -d github.com/mikhailswift/yadc/commands

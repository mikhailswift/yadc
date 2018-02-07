SHELL := /bin/bash

all: fmt lint vet test build

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

SHELL := /bin/bash

.PHONY: build push

build:
	@docker build -f build/Dockerfile -t yadc:build build

.ONESHELL:
push:
	@if [[ -z "${IMAGE}" ]]; then
	    echo "Required env IMAGE not set"
	    exit 1
	fi
	docker tag yadc:${IMAGE} scoodah/yadc:${IMAGE}
	docker push scoodah/yadc:${IMAGE}

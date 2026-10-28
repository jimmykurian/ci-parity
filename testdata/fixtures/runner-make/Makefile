.PHONY: build test lint
GO ?= go

build:
	$(GO) build ./...

test:
	go test ./...

lint:
	golangci-lint run

include mk/release.mk

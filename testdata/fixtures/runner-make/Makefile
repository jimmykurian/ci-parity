.PHONY: build test lint
GO ?= go

build:
	$(GO) build ./...

test:
	go test ./...

lint:
	golangci-lint run

include mk/release.mk

check: build
	go vet ./...

tidy:
	go mod tidy
	go generate ./...

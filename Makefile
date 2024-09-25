PKG := github.com/kittipat1413/go-common
GOLINT ?= golangci-lint
GO_FILES = $(shell go list ./... | grep -v -e /mocks -e /example)

install:
	@test -e $(shell go env GOPATH)/bin/mockgen || go install github.com/golang/mock/mockgen@v1.7.0-rc.1

precommit: lint test

lint:
	$(GOLINT) run

test:
	@go test $(GO_FILES)/... -cover --race; \

generate-mock:
	@go generate ./...
PKG := github.com/kittipat1413/go-common
GOLINT ?= golangci-lint
GO_FILES = $(shell go list ./... | grep -v -e /mocks -e /example)
GO_BIN = $(shell go env GOPATH)/bin

install:
	@test -e $(GO_BIN)/mockgen || go install github.com/golang/mock/mockgen@v1.7.0-rc.1

precommit: lint test

lint:
	@$(GOLINT) run

test:
	@go test $(GO_FILES)/... -cover --race

test-coverage:
	@go test $(GO_FILES)/... -race -covermode=atomic -coverprofile coverage.out

open-coverage-report:
	@go tool cover -html coverage.out -o coverage.html;
	@open coverage.html

generate-mock:
	@go generate ./...
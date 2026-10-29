PKG := github.com/kittipat1413/go-common
GOLINT ?= golangci-lint
GO_FILES = $(shell go list ./... | grep -v -e /mocks -e /example)
GO_BIN = $(shell go env GOPATH)/bin

all: generate-mock precommit

install:
	@echo "Installing tools... 🚀"
	@test -e $(GO_BIN)/mockgen || go install github.com/golang/mock/mockgen@v1.7.0-rc.1

precommit: lint test trivy-scan

lint:
	@echo "Running linters... 🧹"
	@$(GOLINT) run

trivy-scan:
	@echo "Running Trivy scan... 🔍"
	@trivy fs --exit-code 1 --severity LOW,MEDIUM,HIGH,CRITICAL --no-progress .

test:
	@echo "Running tests... 🧪"
	@go test $(GO_FILES)/... -cover --race

test-coverage:
	@echo "Running tests with coverage... 🧪"
	@go test $(GO_FILES)/... -race -covermode=atomic -coverprofile coverage.out
	@go tool cover -func=coverage.out -o=coverage_summary.out
	@cat coverage_summary.out | grep total | awk '{print "Total coverage: " $$3}'
	
open-coverage-report:
	@echo "Opening coverage report... 📊"
	@go tool cover -html coverage.out -o coverage.html;
	@open coverage.html

generate-mock:
	@echo "Generating mock for interfaces... 🛠️"
	@go generate ./...
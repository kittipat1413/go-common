PKG := github.com/kittipat1413/go-common
GOLINT ?= golangci-lint

install:
	go install github.com/golang/mock/mockgen@v1.7.0-rc.1

precommit: lint test

lint:
	$(GOLINT) run

test:
	go test -count=1 $(PKG)/... -cover; \

generate-mock:
	go generate ./...
default: build test

build:
	go build ./...

test:
	go test ./.

mod:
	go mod tidy

.PHONY: default build test mod

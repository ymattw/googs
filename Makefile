default: build test

build:
	go build ./...

test:
	go test ./.

mod:
	go mod tidy

publish:
	go list -m github.com/ymattw/googs@$(shell git rev-parse HEAD)

.PHONY: default build test mod publish

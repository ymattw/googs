build:
	go build ./...

mod:
	go mod tidy

.PHONY: cli mod

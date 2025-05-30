cli:
	go build -o cli ./example/main.go

mod:
	go mod tidy

.PHONY: cli mod

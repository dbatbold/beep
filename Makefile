.PHONY: build format

build:
	go build ./cmd/beep

format:
	gofmt -w -l .

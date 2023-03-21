.PHONY: build format test

build:
	go build ./cmd/beep

format:
	gofmt -w -l .

test:
	go test

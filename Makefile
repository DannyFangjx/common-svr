.PHONY: fmt test build run

fmt:
	gofmt -w $$(find . -name '*.go' -not -path './vendor/*')

test:
	go test ./...

build:
	mkdir -p output
	go build -o output/common-svr .

run:
	go run .

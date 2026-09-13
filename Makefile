.PHONY: fmt test test-integration test-e2e build run

fmt:
	gofmt -w $$(find . -name '*.go' -not -path './vendor/*')

test:
	go test ./...

test-integration:
	go test -tags=integration ./tests/integration/...

test-e2e:
	go test -tags=e2e ./tests/e2e/...

build:
	mkdir -p output
	go build -o output/common-svr .

run:
	go run .

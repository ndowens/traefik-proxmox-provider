.PHONY: build test lint vendor clean

build:
	go build ./...

test:
	go test ./...

lint:
	golangci-lint run ./...

vendor:
	go mod vendor

clean:
	rm -rf dist/

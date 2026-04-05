.PHONY: test lint build ci

test:
	go test -race -count=1 ./...

lint:
	golangci-lint run

build:
	go build -o tick .

ci: lint test

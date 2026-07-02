.PHONY: build test lint vet check

build:
	go build -o gpc ./cmd/gpc

test:
	go test ./...

vet:
	go vet ./...

lint: vet

check: build vet test

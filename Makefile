.PHONY: build test lint all install clean

build:
	go build ./cmd/...

test:
	go test ./...

lint:
	golangci-lint run

all: lint test build

install:
	go install ./cmd/...

clean:
	rm -f hocon2json hocon2yaml hocon2toml hocon2properties

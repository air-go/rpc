.PHONY: all build test clean

all: build lint test

lint:
	go vet ./...
	gofmt -l .
	golint ./...
build:
	go mod download 

test:
	go test -race -cover -coverpkg=./... ./...  -gcflags="-N -l"

clean:
	go clean -i -n -r
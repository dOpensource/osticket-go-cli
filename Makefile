.PHONY: build clean install test all

BINARY=osticket
VERSION=1.0.0

all: build

build:
	go build -ldflags="-s -w" -o $(BINARY) ./cmd/osticket

install: build
	sudo cp $(BINARY) /usr/local/bin/

clean:
	rm -f $(BINARY) $(BINARY)-*

test:
	go test -v ./...

# Cross-compilation targets
build-all: build-linux build-darwin build-windows

build-linux:
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $(BINARY)-linux-amd64 ./cmd/osticket
	GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o $(BINARY)-linux-arm64 ./cmd/osticket

build-darwin:
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o $(BINARY)-darwin-amd64 ./cmd/osticket
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o $(BINARY)-darwin-arm64 ./cmd/osticket

build-windows:
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o $(BINARY).exe ./cmd/osticket

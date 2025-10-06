
.PHONY: build run test

build:
	mkdir -p build
	go build -o build/watchalong cmd/server.go

run:
	go run cmd/server.go

test:
	go test -v ./tests

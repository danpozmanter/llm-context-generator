BINARY=context-generator

.PHONY: all build run test clean

all: build

build:
	go build -o $(BINARY) .

run: build
	./$(BINARY)

test:
	go test ./... -cover

testreport:
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out
	rm -fv coverage.out

clean:
	rm -f $(BINARY)

lint:
	golangci-lint run
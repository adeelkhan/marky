.PHONY: build test coverage lint clean install

BIN := bin/marky

build:
	go build -o $(BIN) .

test:
	go test ./...

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

lint:
	golangci-lint run

clean:
	rm -rf bin/ coverage.out

install:
	go install .

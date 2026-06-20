BINARY := lode
PKG := ./...

.PHONY: build test test-short lint oracle bench clean

build:
	CGO_ENABLED=0 go build -ldflags "-s -w" -o $(BINARY) ./cmd/lode

test:
	go test $(PKG)

test-short:
	go test -short $(PKG)

oracle:
	go test ./tests/oracle/...

bench:
	go test -bench=. -benchmem ./tests/integration/...

lint:
	golangci-lint run

clean:
	rm -f $(BINARY)
	rm -rf dist/

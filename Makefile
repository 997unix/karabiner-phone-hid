.PHONY: test build clean

BIN := bin/karabiner-phone-hid

test:
	go test ./internal/... -v

build:
	go build -o $(BIN) ./cmd/server

clean:
	rm -rf bin/

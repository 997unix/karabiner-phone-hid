.PHONY: test build clean shim

BIN := bin/karabiner-phone-hid

# Build the C shim static library (stub mode by default).
# Set STUB=0 and ensure vendor-cpp/ has Karabiner headers for real build.
shim:
	$(MAKE) -C cshim

test: shim
	go test ./internal/... -v

build: shim
	go build -o $(BIN) ./cmd/server

clean:
	rm -rf bin/
	$(MAKE) -C cshim clean

.PHONY: all build shim test clean install uninstall sudoers unsudoers

BIN      := bin/karabiner-phone-hid
PREFIX   ?= /usr/local
STUB     ?= 0

all: build

# Build C shim → static library, then Go binary.
shim:
	$(MAKE) -C cshim STUB=$(STUB)

build: shim
	@mkdir -p $(dir $(BIN))
	go build -o $(BIN).tmp ./cmd/server
	mv $(BIN).tmp $(BIN)
	@echo "Built $(BIN)"

test:
	$(MAKE) -C cshim STUB=1
	go test ./internal/... -v

clean:
	rm -rf bin/
	$(MAKE) -C cshim clean

install: build
	install -d $(PREFIX)/bin
	install -m 755 $(BIN) $(PREFIX)/bin/
	install -d $(PREFIX)/share/karabiner-phone-hid/web
	install -m 644 web/index.html $(PREFIX)/share/karabiner-phone-hid/web/
	@echo "Installed to $(PREFIX)/bin/karabiner-phone-hid"
	@echo "Run with: sudo karabiner-phone-hid -web $(PREFIX)/share/karabiner-phone-hid/web"

uninstall:
	rm -f $(PREFIX)/bin/karabiner-phone-hid
	rm -rf $(PREFIX)/share/karabiner-phone-hid

SUDOERS_FILE := scripts/sudoers/karabiner-phone-hid
SUDOERS_DEST := /etc/sudoers.d/karabiner-phone-hid

sudoers:
	visudo -cf $(SUDOERS_FILE)
	sudo install -m 440 -o root -g wheel $(SUDOERS_FILE) $(SUDOERS_DEST)
	@echo "Installed $(SUDOERS_DEST) — run without password: sudo karabiner-phone-hid"

unsudoers:
	sudo rm -f $(SUDOERS_DEST)
	@echo "Removed $(SUDOERS_DEST)"

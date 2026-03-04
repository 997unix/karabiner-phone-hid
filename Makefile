.PHONY: all build shim test clean install uninstall sudoers unsudoers s6-up s6-down s6-status s6-log

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

# s6 supervision — uses s6-svscan to wire service + log pipeline
S6_SCANDIR := scripts/s6
S6_SVCDIR  := $(S6_SCANDIR)/karabiner-phone-hid
S6_LOGDIR  := $(HOME)/.local/log/karabiner-phone-hid

s6-up: build
	@mkdir -p $(S6_LOGDIR)
	@if s6-svstat $(S6_SVCDIR) 2>/dev/null; then \
		echo "Already running — sending restart"; \
		s6-svc -r $(S6_SVCDIR); \
	else \
		s6-svscan $(S6_SCANDIR) & \
		echo "s6-svscan started — use 'make s6-status' or 'make s6-log'"; \
	fi

s6-down:
	@s6-svc -d $(S6_SVCDIR) 2>/dev/null; true
	@s6-svscanctl -q $(S6_SCANDIR) 2>/dev/null; true
	@echo "Service stopped"

s6-status:
	s6-svstat $(S6_SVCDIR)

s6-log:
	tail -f $(S6_LOGDIR)/current | s6-tai64nlocal

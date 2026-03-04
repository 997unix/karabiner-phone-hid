.PHONY: all build shim test clean install uninstall sudoers unsudoers \
       s6-install s6-uninstall s6-scan s6-up s6-down s6-restart s6-status s6-log

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

# s6 supervision
#   Scan dir: ~/.local/share/s6/scan/   (s6-svscan watches this)
#   Service:  symlinked from project's scripts/s6/karabiner-phone-hid/
#   Logs:     ~/.local/log/karabiner-phone-hid/current
S6_SCANDIR := $(HOME)/.local/share/s6/scan
S6_SVCNAME := karabiner-phone-hid
S6_SVCLINK := $(S6_SCANDIR)/$(S6_SVCNAME)
S6_LOGDIR  := $(HOME)/.local/log/$(S6_SVCNAME)
S6_SRCDIR  := $(CURDIR)/scripts/s6/$(S6_SVCNAME)

s6-install: build
	@mkdir -p $(S6_SCANDIR) $(S6_LOGDIR)
	@ln -sfn $(S6_SRCDIR) $(S6_SVCLINK)
	@echo "Installed: $(S6_SVCLINK) → $(S6_SRCDIR)"
	@echo "Logs:      $(S6_LOGDIR)/current"
	@echo ""
	@echo "Start the scanner (once, or add to login items):"
	@echo "  s6-svscan $(S6_SCANDIR) &"
	@echo ""
	@echo "Then use:"
	@echo "  make s6-status    # check service"
	@echo "  make s6-log       # tail logs"
	@echo "  make s6-down      # stop service"

s6-uninstall:
	@s6-svc -d $(S6_SVCLINK) 2>/dev/null; true
	@rm -f $(S6_SVCLINK)
	@echo "Removed $(S6_SVCLINK)"

s6-scan:
	@mkdir -p $(S6_SCANDIR)
	s6-svscan $(S6_SCANDIR) &
	@echo "s6-svscan running on $(S6_SCANDIR)"

s6-up:
	s6-svc -u $(S6_SVCLINK)

s6-down:
	s6-svc -d $(S6_SVCLINK)

s6-restart: build
	s6-svc -r $(S6_SVCLINK)

s6-status:
	s6-svstat $(S6_SVCLINK)
	@s6-svstat $(S6_SVCLINK)/log 2>/dev/null; true

s6-log:
	tail -f $(S6_LOGDIR)/current | s6-tai64nlocal

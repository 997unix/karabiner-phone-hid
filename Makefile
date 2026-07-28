.PHONY: all build shim test clean install uninstall sudoers unsudoers status \
       s6-install s6-uninstall s6-up s6-down s6-restart s6-status s6-log

BIN      := bin/karabiner-phone-hid
PREFIX   ?= /usr/local
STUB     ?= 0
PORT     ?= 8765

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
#   Service:  run scripts copied from project's scripts/s6/karabiner-phone-hid/
#   Logs:     ~/.local/log/karabiner-phone-hid/current
S6_SCANDIR := $(HOME)/.local/share/s6/scan
S6_SVCNAME := karabiner-phone-hid
S6_SVCDIR  := $(S6_SCANDIR)/$(S6_SVCNAME)
S6_LOGDIR  := $(HOME)/.local/log/$(S6_SVCNAME)
S6_SRCDIR  := $(CURDIR)/scripts/s6/$(S6_SVCNAME)

s6-install: build
	@mkdir -p $(S6_SVCDIR)/log $(S6_LOGDIR)
	install -m 755 $(S6_SRCDIR)/run $(S6_SVCDIR)/run
	install -m 755 $(S6_SRCDIR)/log/run $(S6_SVCDIR)/log/run
	@s6-svc -u $(S6_SVCDIR) 2>/dev/null; true
	@echo "Installed s6 service: $(S6_SVCDIR)"
	@echo "Logs: $(S6_LOGDIR)/current"

s6-uninstall:
	@s6-svc -dx $(S6_SVCDIR) 2>/dev/null; true
	rm -rf $(S6_SVCDIR)
	@echo "Removed $(S6_SVCDIR)"

s6-up:
	s6-svc -u $(S6_SVCDIR)

s6-down:
	s6-svc -d $(S6_SVCDIR)

s6-restart: build s6-install

s6-status:
	s6-svstat $(S6_SVCDIR)
	@s6-svstat $(S6_SVCDIR)/log 2>/dev/null; true

# "Is it working right now?" -- supervision state plus live HID readiness.
#
# Deliberately not a log query. s6-log is size-bounded (10MB x 5), so startup
# lines age out, and readiness is logged only on transitions -- a healthy
# server writes nothing for days. Silence in the log is ambiguous; /readyz
# is not.
status:
	@printf 's6:   '
	@s6-svstat $(S6_SVCDIR) 2>/dev/null || echo "service not installed"
	@printf 'hid:  '
	@curl -s -m 2 http://127.0.0.1:$(PORT)/readyz 2>/dev/null \
	  | jq -er '"ready=\(.ready) connected=\(.connected) keyboard=\(.keyboard_ready) pointing=\(.pointing_ready) up=\(.uptime) build=\(.checksum)"' \
	  || echo "no answer on port $(PORT)"

s6-log:
	tail -f $(S6_LOGDIR)/current | s6-tai64nlocal

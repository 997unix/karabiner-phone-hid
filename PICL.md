# Pre-Install Checklist (PICL) — karabiner-phone-hid

Project-local checklist. Overrides the global `~/.claude/PICL.md`, so the
general items are repeated here.

## Before Completing Any Task

- [ ] All tests pass (`make test`)
- [ ] No new test gaps — if you added code, you added tests
- [ ] Documentation updated for any new or changed functionality
- [ ] No secrets, credentials, or tokens in committed files
- [ ] No TODO/FIXME left behind without a tracking issue

## Before Committing

- [ ] `git diff` reviewed — no accidental files, debug prints, or commented-out code
- [ ] Commit message follows project conventions (`[component] description`)

## Before Calling Task Complete

- [ ] User has been notified with test counts and commit hash
- [ ] Acceptance testing requested

## Project-Specific

### Build

- [ ] The shipped binary is the **real** shim, not the stub. `make test` builds
      with `STUB=1`; a real build needs `STUB=0` (the default at top level).
      Verify with:
      `strings bin/karabiner-phone-hid | grep -c 'karabiner_shim] STUB'` → must be 0.
      A stub binary acks websocket keystrokes and types nothing, which looks
      exactly like the daemon being unreachable.
- [ ] `make test` leaves the shim built with `STUB=1`. Re-run `make` before
      restarting the service, or s6 will keep running the previous binary.

### Karabiner daemon compatibility

- [ ] After any Karabiner-Elements update, confirm the vendored headers still
      match the daemon. The daemon logs what it expects on startup:
      `grep client_protocol_version /var/log/karabiner/virtual_hid_device_service.log`
      Compare against
      `vendor-headers/karabiner/pqrs/karabiner/driverkit/client_protocol_version.hpp`.
      A mismatch shows up as a once-a-second
      `[Karabiner] Connection failed, re-initializing pointing device` loop.
      Re-vendoring steps are in README → "Karabiner Version Compatibility".

### Verifying a change actually works

- [ ] `make status` reports `s6: up` and `hid: ready=true` with all three of
      connected / keyboard / pointing true, and a `build=` checksum matching
      the binary you just deployed. This is the quick check; the log lines
      below are for reconstructing what happened, not for current state.
- [ ] Server side: `~/.local/log/karabiner-phone-hid/current` should show
      `Connected to daemon` → `Keyboard ready` → `Pointing ready` → `Listening on port`.
- [ ] Daemon side: `/var/log/karabiner/virtual_hid_device_service.log` should show
      a new `peer_connected`, `driver_version_ is changed: 10800`, and both
      `virtual_hid_keyboard_ready_` / `virtual_hid_pointing_ready_` true, with no
      `driver_version_mismatched`. Server logs alone do not prove the daemon
      accepted the client.

### Gotchas

- Apple's `/usr/bin/make` is GNU Make 3.81 — whole-second mtime granularity, and
  it stats targets before running recipes. Don't rely on a stamp file's mtime, or
  on a recipe deleting artifacts, to force a rebuild. Do that work at parse time
  with `$(shell ...)`.
- `sudo` is passwordless only for the server binary itself (see
  `scripts/sudoers/`), so the daemon's root-only socket directory
  `/Library/Application Support/org.pqrs/tmp/rootonly` can't be inspected
  directly. Use `/var/log/karabiner/*.log`, which is world-readable, instead.

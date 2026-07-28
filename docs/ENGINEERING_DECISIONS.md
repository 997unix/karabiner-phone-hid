# Engineering Decisions

A running log of the non-obvious choices behind this repo. Each entry names the decision, the alternative that was rejected, and why. New decisions go at the top.

The aim is to make the "why" recoverable later — both for humans and for LLMs collaborating on the code — without cluttering source files with long comments.

---

## ED-008 — Device recovery keys off readiness state, not a heartbeat

**Decision:** Connection and device readiness are tracked as explicit state
(`linkState` in `internal/hid/link_state.go`). Recovery triggers on "connected
but pointing not ready"; every log line is emitted on a state *transition*.
The shim forwards both edges of the driver's readiness signals — see the
`*_NOT_READY` status codes in `cshim/karabiner_shim.h`.

**Rejected:**
- Keeping the ED-006 heartbeat and re-initializing when callbacks go quiet.
- Forcing re-initialization on a timer regardless of state.

**Why:** Karabiner-DriverKit-VirtualHIDDevice 8.x reports readiness only when
it changes. Pre-8.0 re-broadcast it periodically, which is what made ED-006's
heartbeat interpretation work. Under 8.x that heartbeat never arrives, so the
old watchdog fired every 30s forever while the device was perfectly healthy,
and its non-forced re-init was a silent no-op (the 8.x client drops an
initialize request for a device that is already up). Silence now correctly
means "nothing changed".

Two supporting fixes fell out of this. The shim previously discarded
`ready == false`, so the server could see a device appear but never learn it
had gone — the exact condition the watchdog existed to catch. And the connect
handler only created the keyboard, so a device lost to a reconnect was never
re-created; it now creates both on every connect, which is the primary
recovery path. The watchdog is the backstop for a device dying while the
connection survives.

Logging on transitions also fixes the log volume: the 8.x client retries the
socket once a second while the daemon is down, and the old code logged every
attempt. That is how the log reached 8MB.

---

## ED-007 — Self-exec on binary rebuild

**Decision:** The server watches its own file checksum every 30s (`internal/buildinfo/watch.go`). When the on-disk binary changes, it calls `syscall.Exec` on itself to hot-swap into the new build.

**Rejected:**
- Requiring the user to `s6-svc -r` after every `make`.
- A file-watch triggered rebuild on the source tree (fragile; makes the dev loop depend on the server being running).

**Why:** `make` alone is the whole dev loop. You rebuild, the running process notices within 30s and replaces itself. Combined with s6, crashes are also recovered — so the supervisor and the self-exec cover different failure modes without overlap.

---

## ED-006 — Pointing-device watchdog

> **Superseded by ED-008.** The heartbeat reading below was correct against the
> pre-8.0 driver, which re-broadcast readiness periodically. Karabiner 8.x
> reports it only on change, so the watchdog now keys off readiness state.
> The underlying problem — and the watchdog itself — remain real.

**Decision:** After `InitPointing()`, a watchdog goroutine (`StartPointingWatchdog` in `internal/hid/karabiner_poster.go`) re-initializes the pointing device if we stop getting `POINTING_READY` callbacks.

**Rejected:** Trusting that once-ready stays ready.

**Why:** The DriverKit daemon can stall the pointing endpoint without dropping the whole connection. Without the watchdog, the mouse tab on the phone silently stops working until the server is restarted. Commit `4123702` added this after real-world flakiness.

---

## ED-005 — CGo + C++20 header-only client, wrapped behind a plain-C shim

**Decision:** `cshim/karabiner_shim.cpp` compiles the pqrs C++ headers into a static library that exposes a minimal plain-C ABI (`cshim/karabiner_shim.h`). Go imports that via `#cgo LDFLAGS: -lkarabiner_shim -lc++`.

**Rejected:**
- SwigGo / cgo-directly-against-C++ — fragile, painful template instantiation errors at cgo parse time.
- A separate Swift/Obj-C helper process talking to Go over a Unix socket — two binaries to ship, two sets of logs, two supervisors.

**Why:** The Karabiner client is C++-only and header-only. CGo cannot consume C++ directly, but it consumes plain C fine. A single `.cpp` + `.h` pair gives us one static library with a stable C ABI; every other layer (dispatcher, router, protocol) stays pure Go and testable.

**Corollary (ED-005a):** A `KARABINER_STUB` define in the shim produces no-op C functions. `make test` builds the shim with `STUB=1` so `go test ./internal/...` runs on any Mac — including CI without Karabiner installed.

---

## ED-004 — `HIDPoster` interface in front of the CGo layer

**Decision:** `internal/hid/poster.go` defines a small `HIDPoster` interface. The real implementation (`karabiner_poster.go`) is CGo. The dispatcher (`dispatcher.go`) depends only on the interface.

**Rejected:** Having the dispatcher call CGo directly.

**Why:** The CGo side cannot be unit-tested portably. The dispatcher can — `hid_test.go` substitutes a mock poster and gets 21 tests. This is the pattern that lets us have 80 tests on a project whose core functionality talks to a kernel extension.

---

## ED-003 — Passwordless sudo scoped to one binary

**Decision:** `scripts/sudoers/karabiner-phone-hid` lets the current user run exactly `/Users/.../bin/karabiner-phone-hid` without a password prompt. Installed via `make sudoers`, validated with `visudo -cf` before install.

**Rejected:**
- Running the whole s6 scan tree as root.
- Making the user type a password every time they want to restart the service.

**Why:** The service *must* be root to talk to the Karabiner daemon socket. Everything else (builds, s6, logs) is fine as the user. Scoping the sudo exemption to one absolute path keeps the blast radius tiny; validating with `visudo` before install keeps a typo from locking the user out of sudo.

---

## ED-002 — s6 instead of launchd for supervision

**Decision:** The service is managed by s6 (`brew install s6`) with run scripts in `scripts/s6/karabiner-phone-hid/`. Everything is under `~/.local/share/s6/scan/` — no root-owned system state.

**Rejected:**
- `launchd` with a `.plist` in `~/Library/LaunchAgents/` or `/Library/LaunchDaemons/`.
- Running the server in a terminal tab and remembering to restart it.

**Why:**
- `launchd` plists are XML, the schema is big, and the failure modes (`launchctl bootstrap`, `disable`, label collisions) are confusing.
- The s6 `run` script is 4 lines. The log `run` script is 2 lines. Anyone can read them.
- `s6-log T s10000000 n5` gives TAI64N-timestamped, size-rotated logs for free. No `log stream --predicate` incantations.
- All state lives under `$HOME`. Uninstalling is `rm -rf`.
- Portable to Linux and to other users' machines without touching `/Library`.

**Trade-off:** Users must `brew install s6`. Worth it.

---

## ED-001 — Single-file phone UI, vanilla JS + p5.js

**Decision:** `web/index.html` is one file containing HTML + CSS + vanilla JS + a p5.js include for the mouse canvas. Served as a static file by the Go server.

**Rejected:**
- A React/Vite build pipeline.
- Progressive Web App with service worker, manifest, etc.

**Why:** Zero install on the phone is the core UX promise — just open the URL. A build pipeline would mean a `node_modules` in this repo and a deploy step. The file is ugly at ~2000 lines, but the cost of that ugliness is borne by one file. The cost of the alternative would be borne by every future contributor.

**Known debt:** This file will be the first thing to split when it grows past what one person can hold in their head. See `docs/REVIEW_BY_FILE.md` for the current rating.

---

## Template for New Entries

```
## ED-NNN — Short decision name

**Decision:** What we do.

**Rejected:** What we considered and did not do.

**Why:** The constraint, incident, or trade-off that forced the choice.
```

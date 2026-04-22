# Review by File

A per-file snapshot of the repository used as a talking point for "what can a non-coder build with an LLM, and how do you critique it?" Ratings are subjective. Test counts are approximate — they come from `^func Test` matches in `_test.go` files, not benchmark or fuzz counts.

## Executive Summary

Go + C++ project that turns a phone browser into a real HID keyboard/mouse for macOS via Karabiner's virtual HID DriverKit. CGo bridges Go to a C++ shim; a WebSocket server takes JSON events from the phone UI and posts them as hardware input.

Roughly **~7,200 lines of code**, **80 Go test functions** across 8 test files. Every Go package in `internal/` has tests. The CGo poster, `main.go`, and the web UI are intentionally uncovered — they touch hardware, process startup, and a browser.

## The s6 Trick (for the demo)

macOS ships `launchd`, but writing a `.plist`, wrestling with `launchctl load/unload/bootstrap`, log redirection quirks, and permission surprises is genuinely ugly for a hobby project. This repo sidesteps all of that by using the s6 process supervisor from Homebrew:

- `scripts/s6/karabiner-phone-hid/run` — **4 lines**. `exec sudo -n .../karabiner-phone-hid …`.
- `scripts/s6/karabiner-phone-hid/log/run` — **2 lines**. `exec s6-log T s10000000 n5 …` gives timestamped, rotated logs automatically.
- Scan directory lives under `~/.local/share/s6/scan/` — no root-owned system state, no `sudo launchctl`.
- Makefile targets (`s6-install`, `s6-status`, `s6-log`, `s6-restart`, `s6-down`, `s6-uninstall`) wrap the whole lifecycle. `s6-restart` rebuilds + reinstalls, and `s6-svc -u` brings the service back up.
- Passwordless `sudo` is isolated to exactly this one binary via `scripts/sudoers/karabiner-phone-hid`, validated with `visudo -cf` before install.

The clever piece: the binary also watches its own checksum (`internal/buildinfo/watch.go`) and `exec()`s the replacement when you rebuild — so `make` alone, without touching s6, picks up code changes. s6 is just the crash-restart + logging safety net.

## File Table

| # | File | Lines | Tests | Rating | Notes |
|---|------|-------|-------|--------|-------|
| 1 | `cmd/server/main.go` | 138 | 0 | Pretty | Clean wiring: flag parse → load config → init HID → start WS → Bonjour → signal wait. Readable top-to-bottom. |
| 2 | `cshim/karabiner_shim.cpp` | 231 | 0 | OK | C++20 wrapping the pqrs header-only client into plain C for cgo. Has a `KARABINER_STUB` mode so the Go tests compile without the real driver. |
| 3 | `cshim/karabiner_shim.h` | 77 | 0 | Pretty | Small, focused C ABI. |
| 4 | `cshim/Makefile` | 29 | 0 | Pretty | Builds static lib; `STUB=1` path for tests. |
| 5 | `internal/buildinfo/identity.go` | 69 | — | Pretty | Reports `checksum/git/go` of the running binary. |
| 6 | `internal/buildinfo/identity_test.go` | 94 | **6** | Well-tested | |
| 7 | `internal/buildinfo/watch.go` | 79 | — | Pretty | Self-exec on binary rebuild — the "auto-reload" trick. |
| 8 | `internal/buildinfo/watch_test.go` | 125 | **5** | Well-tested | |
| 9 | `internal/config/config.go` | **567** | 0 | Heavy | Biggest non-HTML file. Loads/validates JSON actions, builds a registry. Could be split. |
| 10 | `internal/config/config_test.go` | 398 | **16** | Well-tested | Heaviest test file by count — compensates for the source size. |
| 11 | `internal/discovery/bonjour.go` | 43 | 0 | Pretty | Tiny mDNS wrapper. |
| 12 | `internal/discovery/bonjour_test.go` | 19 | **2** | Minimal | Thin, matches surface area. |
| 13 | `internal/hid/dispatcher.go` | 68 | — | Pretty | Translates protocol steps → poster calls. Clear. |
| 14 | `internal/hid/hid_test.go` | 400 | **21** | Well-tested | Keycode + dispatcher coverage via a mock poster. |
| 15 | `internal/hid/karabiner_poster.go` | 260 | 0 | Busy | CGo + C callback bridge + `sync.Mutex` + watchdog + heartbeat-dot printing interleaved. Cohesive but the most intricate Go file in the tree. |
| 16 | `internal/hid/keycodes.go` | 103 | — | Pretty | Flat lookup tables. |
| 17 | `internal/hid/poster.go` | 74 | — | Pretty | `HIDPoster` interface — makes the dispatcher testable. |
| 18 | `internal/protocol/message.go` | 108 | 0 | Pretty | JSON envelope + payload types. |
| 19 | `internal/protocol/message_test.go` | 322 | **11** | Well-tested | |
| 20 | `internal/server/router.go` | 123 | 0 | Pretty | Message → handler dispatch. |
| 21 | `internal/server/router_test.go` | 269 | **13** | Well-tested | |
| 22 | `internal/server/websocket.go` | 121 | 0 | Pretty | `nhooyr.io/websocket` wrapper + static file serving. |
| 23 | `internal/server/websocket_test.go` | 190 | **6** | Well-tested | |
| 24 | `web/index.html` | **1974** | 0 | Monolith | Single-file UI — HTML + CSS + vanilla JS + p5.js for the mouse canvas. Five swipeable tabs in one doc. Impressive it works, but it is the classic LLM single-file artifact. |
| 25 | `web/mouse.html` | 349 | 0 | OK | Standalone mouse view. |
| 26 | `scripts/test_send.py` | 88 | 0 | Pretty | Dev helper for firing WS messages. |
| 27 | `scripts/s6/karabiner-phone-hid/run` | 4 | — | Tiny | The whole service definition. |
| 28 | `scripts/s6/karabiner-phone-hid/log/run` | 2 | — | Tiny | Logging with rotation + TAI64N timestamps. |
| 29 | `scripts/sudoers/karabiner-phone-hid` | 4 | — | Pretty | Minimal, scoped to one binary. |
| 30 | `Makefile` | 88 | — | Pretty | Clean targets for build/test/install/sudoers/s6-*. |

## Rating Legend

- **Pretty** — small, focused, obvious intent; easy for a reader (or an LLM) to modify safely.
- **OK** — reasonable, not beautiful; would benefit from a split or a rename pass but not urgent.
- **Heavy** — too much in one file; hard to hold in your head. First candidate for refactoring.
- **Busy** — intricate for a real reason (CGo lifetimes, concurrency). Looks ugly but the ugliness is in the domain, not the author.
- **Monolith** — one giant file doing five jobs. Works; doesn't scale.
- **Minimal** — thin by design, matches a small surface area.
- **Tiny** — a handful of lines that replaces something much larger (e.g., the s6 `run` script vs. a launchd plist).
- **Well-tested** — a test file that pulls its weight for its paired source file.

## Story Arc for the Demo

1. **Real system software.** Root, CGo, C++, DriverKit, Unix sockets, Bonjour, WebSockets, HID protocol. Not a toy.
2. **Tests where they matter.** 80 tests concentrated on the pure-Go logic; the hardware edge has a stub so tests still run on any Mac.
3. **The ugly parts are the ugly parts of the domain.** `karabiner_poster.go` and `web/index.html` are the two files a senior reviewer would point at — and both are ugly for *real* reasons (CGo callback lifetime management; single-file UI for easy serving), not sloppiness.
4. **s6 over launchd is the punchline.** A 4-line `run` script + a Makefile replaces the entire macOS daemon ritual. That is the "LLM helped me pick the right tool" story, not just the "LLM wrote Go" story.

## Obvious Next Iterations

These are cheap wins to demonstrate the "critique and fix" half of the LLM workflow:

- **Split `internal/config/config.go` (567 lines)** into `types.go`, `load.go`, `registry.go`.
- **Split `web/index.html` (1974 lines)** into per-tab partials or move the mouse canvas JS to a separate file.
- **Add a regression test for the pointing-device watchdog** in `karabiner_poster.go` — the watchdog is the most recent feature (commit `4123702`) and has zero coverage.
- **Remove the stray `./server` binary** from the working tree and add `/server` to `.gitignore` so future accidental builds don't show up as untracked.

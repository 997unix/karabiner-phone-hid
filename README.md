# karabiner-phone-hid

Turn your phone into a wireless keyboard for your Mac — keystrokes appear as real hardware input through [Karabiner's virtual HID device](https://github.com/pqrs-org/Karabiner-DriverKit-VirtualHIDDevice).

No app to install. Open a browser on your phone, tap buttons, keystrokes appear on your Mac.

## Prerequisites

- **macOS 13+**
- **[Karabiner-Elements](https://karabiner-elements.pqrs.org/)** installed and running (provides the DriverKit virtual HID daemon), shipping **Karabiner-DriverKit-VirtualHIDDevice 8.x** — see [Karabiner version compatibility](#karabiner-version-compatibility)
- **Go 1.23+** (`brew install go`)
- **Xcode Command Line Tools** (`xcode-select --install`)
- **A phone/tablet on the same network** with any modern browser

Optional, for process supervision:

- **[s6](https://skarnet.org/software/s6/)** (`brew install s6`) — recommended for running as a long-lived service

## Why

- Keystrokes go through Karabiner's virtual keyboard, so they get full remapping support
- Events look like physical keyboard input to every app (not Accessibility API injection)
- Five swipeable tabs — coding assistant, terminal, YouTube remote, full keyboard, numpad
- Works from any device with a browser — phone, tablet, another computer

## Install

```bash
git clone https://github.com/997unix/karabiner-phone-hid.git
cd karabiner-phone-hid
make
```

This builds the C++ shim against Karabiner's headers and compiles the Go server into `bin/karabiner-phone-hid`.

## Usage

### Run directly

```bash
sudo ./bin/karabiner-phone-hid
```

Root is required to talk to Karabiner's daemon via its Unix socket.

To run without password prompts, install the sudoers rule:

```bash
make sudoers
```

Then you can start without typing your password:

```bash
sudo ./bin/karabiner-phone-hid
```

Stop with `Ctrl+C`.

### Run with s6 (recommended)

[s6](https://skarnet.org/software/s6/) keeps the server running in the background, restarts it on crashes, and captures logs with timestamps. If you want the server always available, this is the way to go.

```bash
make sudoers         # one-time: passwordless sudo for the binary
make s6-install      # build, install, and start the supervised service
```

Useful commands:

```bash
make s6-status       # check if the service is running
make s6-log          # tail logs (timestamped)
make s6-restart      # rebuild + reinstall (picks up code changes)
make s6-down         # stop the service
make s6-uninstall    # remove the service entirely
```

Logs go to `~/.local/log/karabiner-phone-hid/current`.

### Startup output

```
[Server] start checksum=a1b2c3d4e5f6 git=7a21ceb go=go1.23.5
[Karabiner] Connected to daemon
[Karabiner] Keyboard ready
[Server] Serving web UI from web
[Server] Listening on port 8765
[Bonjour] Published: Tonys-Mac on port 8765
```

Then open your phone browser to:

```
http://<your-mac-ip>:8765
```

Find your Mac's IP with `ipconfig getifaddr en0`.

Tap a button. The keystroke appears on your Mac. Swipe to switch tabs.

## Tabs

### tmate/tmux

Coding assistant controls — SuperWhisper voice input, accept/reject inline suggestions, tmux copy mode, arrow keys.

### Terminal

tmux session management — prefix, copy mode, scroll, split panes, navigate panes and windows.

### YouTube Remote

YouTube keyboard shortcuts — play/pause, seek ±10s, volume, speed ±, fullscreen, captions, next/prev video.

### Keyboard

Full programmer's keyboard in portrait mode:

- SuperWhisper button at the top (long-press to paste)
- F1–F12 function key row
- Number row, QWERTY layout, symbols
- Ctrl / Opt / Cmd modifier keys on both sides
- Inverted-T arrow cluster, PgUp/PgDn

### Numpad

Standard numpad layout per the [W3C UIEvents spec](https://www.w3.org/TR/uievents-code/):

- Control pad: Insert, Home, PgUp / Delete, End, PgDn
- Numpad: Esc / * − across top, 7-8-9 / 4-5-6 / 1-2-3 / 0 . with + and Enter spanning two rows

All numpad keys send distinct USB HID numpad scancodes (not the number row keys).

## Custom Actions

Create `~/.config/karabiner-phone-hid/actions.json`:

```json
{
  "actions": [
    {
      "name": "screenshot",
      "label": "Screenshot",
      "steps": [
        {"key": "4", "modifiers": ["command", "shift"]}
      ]
    },
    {
      "name": "lock_screen",
      "label": "Lock",
      "steps": [
        {"key": "q", "modifiers": ["command", "control"]}
      ]
    }
  ]
}
```

Custom actions merge with defaults. To override a default, use the same `name`.

Key names are USB HID standard — see [`internal/hid/keycodes.go`](internal/hid/keycodes.go) for the full list.

Modifiers: `control`, `shift`, `option`, `command`

## Flags

```
-port 8765    WebSocket listen port
-name "Mac"   Server name for Bonjour discovery
-web ./web    Path to web UI directory
```

## Architecture

```
Phone Browser  ──WebSocket──▶  Go Server  ──CGo──▶  Karabiner VHD Daemon  ──▶  macOS HID
                                 :8765       C shim      Unix socket
```

The Go server receives JSON messages over WebSocket, translates them to USB HID keyboard reports, and sends them to Karabiner's DriverKit daemon via a C++ client library (wrapped in a C shim for CGo). The daemon feeds them to a virtual keyboard device that macOS treats as real hardware.

### Auto-reload

The server hashes its own binary on startup. Every 30 seconds it re-checks the hash — if the binary changed (i.e. you ran `make`), it `exec()`s the new version in place. The new process logs its provenance:

```
[Server] re-exec from checksum=a1b2c3d4e5f6 to checksum=f6e5d4c3b2a1 git=deadbee go=go1.23.5
```

## Wire Protocol

The WebSocket protocol is simple JSON. See [`shared/protocol.md`](shared/protocol.md).

**Send a keypress:**
```json
{"type":"action","id":"1","action":"keypress","payload":{"key":"a","modifiers":["shift"]}}
```

**Send a named action:**
```json
{"type":"action","id":"2","action":"superwhisper_toggle","payload":{}}
```

**Send a sequence:**
```json
{"type":"action","id":"3","action":"sequence","payload":{"steps":[
  {"key":"backslash","modifiers":["control"],"delay_ms":100},
  {"key":"open_bracket","modifiers":[]}
]}}
```

Responses are `{"type":"ack","id":"1","status":"ok"}`.

## Development

Run tests (no root or Karabiner needed):

```bash
make test
```

80 tests across 6 packages — protocol serialization, key code lookups, HID dispatch, message routing, WebSocket integration, build identity.

Build with stub mode (no Karabiner headers needed):

```bash
make STUB=1
```

The build tracks the `STUB` value, so switching between stub and real
rebuilds the shim rather than silently reusing the previous object.

## Karabiner Version Compatibility

`vendor-headers/` holds a pinned copy of the Karabiner-DriverKit-VirtualHIDDevice
client headers. They are currently on **v8.0.0** (`client_protocol_version` 7,
`driver_version` 10800 — matching driver extension 1.8.0). This covers the whole
8.x line; v8.1 and v8.2 use the same protocol version.

Updating Karabiner-Elements can bump the daemon's protocol and break the
connection. The symptom is the server looping on:

```
[Karabiner] Connection failed, re-initializing pointing device
```

The daemon reports what it expects on startup:

```bash
grep -E 'version|client_protocol_version' /var/log/karabiner/virtual_hid_device_service.log
```

If `client_protocol_version` there differs from
`vendor-headers/karabiner/pqrs/karabiner/driverkit/client_protocol_version.hpp`,
re-vendor headers matching the installed daemon:

```bash
git clone --depth 1 --branch v<VERSION> --recurse-submodules \
  https://github.com/pqrs-org/Karabiner-DriverKit-VirtualHIDDevice.git /tmp/kdk
rm -rf vendor-headers/karabiner vendor-headers/deps
mkdir -p vendor-headers/karabiner vendor-headers/deps
cp -R /tmp/kdk/include/. vendor-headers/karabiner/
cp -R /tmp/kdk/vendor/vendor/include/. vendor-headers/deps/
make && make s6-restart
```

Check upstream `NEWS.md` for breaking API changes before rebuilding; the
8.0.0 upgrade changed the transport but left the client API untouched.

## Project Structure

```
cmd/server/main.go       Entry point
internal/
  protocol/              JSON message types
  config/                Action registry + config file
  hid/                   USB keycodes, HIDPoster interface, dispatcher
  server/                WebSocket server + message router
  discovery/             Bonjour advertisement
  buildinfo/             Binary identity, self-watch + auto-reload
cshim/                   C wrapper around Karabiner C++ client
vendor-headers/          Karabiner + dependency headers (vendored copy)
web/                     Browser UI served at /
shared/protocol.md       Wire protocol spec
scripts/s6/              s6 supervision run scripts
```

## License

MIT

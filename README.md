# karabiner-phone-hid

Turn your phone into a wireless keyboard for your Mac — keystrokes appear as real hardware input through [Karabiner's virtual HID device](https://github.com/pqrs-org/Karabiner-DriverKit-VirtualHIDDevice).

No app to install. Open a browser on your phone, tap buttons, keystrokes appear on your Mac.

## Why

- Keystrokes go through Karabiner's virtual keyboard, so they get full remapping support
- Events look like physical keyboard input to every app (not Accessibility API injection)
- Five swipeable tabs — coding assistant, terminal, YouTube remote, full keyboard, numpad
- Works from any device with a browser — phone, tablet, another computer

## Prerequisites

- **macOS 13+**
- **[Karabiner-Elements](https://karabiner-elements.pqrs.org/)** installed and running
- **Go 1.23+** and **Xcode Command Line Tools** (`xcode-select --install`)

Karabiner-Elements includes the DriverKit virtual HID daemon. If Karabiner is running, you're good.

## Install

```bash
git clone --recurse-submodules https://github.com/997unix/karabiner-phone-hid.git
cd karabiner-phone-hid
make
```

This builds the C++ shim against Karabiner's headers and compiles the Go server into `bin/karabiner-phone-hid`.

## Usage

```bash
sudo ./bin/karabiner-phone-hid
```

Root is required to talk to Karabiner's daemon via its Unix socket.

You'll see:

```
[Server] start checksum=a1b2c3d4e5f6 git=1155fdd go=go1.23.0
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

### Teammate

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

Standalone number pad — 0–9, period, Enter, Backspace, Delete, Esc.

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
[Server] re-exec from checksum=a1b2c3d4e5f6 to checksum=f6e5d4c3b2a1 git=deadbee go=go1.23.0
```

### s6 supervision (optional)

```bash
make s6-install       # install + start supervised service
make s6-status        # check service status
make s6-log           # tail logs
make s6-restart       # rebuild + reinstall service
make s6-down          # stop service
make s6-uninstall     # remove service
```

Logs go to `~/.local/log/karabiner-phone-hid/current`.

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

66 tests across 6 packages — protocol serialization, key code lookups, HID dispatch, message routing, WebSocket integration, build identity.

Build with stub mode (no Karabiner headers needed):

```bash
make STUB=1
```

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
vendor-cpp/              Karabiner headers (git submodule)
web/                     Browser UI served at /
shared/protocol.md       Wire protocol spec
scripts/s6/              s6 supervision run scripts
```

## License

MIT

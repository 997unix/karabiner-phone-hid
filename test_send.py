#!/usr/bin/env python3
"""Send "robot test loop <date>\n" via WebSocket to karabiner-phone-hid server."""

import asyncio
import json
import sys
from datetime import datetime

try:
    import websockets
except ImportError:
    print("Installing websockets...")
    import subprocess
    subprocess.check_call([sys.executable, "-m", "pip", "install", "websockets", "-q"])
    import websockets


async def main():
    uri = "ws://127.0.0.1:8765/ws"
    text = f"robot test loop {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}"

    async with websockets.connect(uri) as ws:
        # Read config message
        config = await ws.recv()
        print(f"Connected: {json.loads(config).get('type')}")

        # Type each character
        for ch in text:
            key, mods = char_to_key(ch)
            msg = json.dumps({
                "type": "action",
                "id": f"t-{id(ch)}",
                "action": "keypress",
                "payload": {"key": key, "modifiers": mods}
            })
            await ws.send(msg)
            ack = await ws.recv()
            status = json.loads(ack).get("status")
            if status != "ok":
                print(f"Error on '{ch}': {ack}")
                return
            await asyncio.sleep(0.03)  # small delay between keys

        # Press enter
        msg = json.dumps({
            "type": "action",
            "id": "t-enter",
            "action": "keypress",
            "payload": {"key": "return_or_enter", "modifiers": []}
        })
        await ws.send(msg)
        ack = await ws.recv()
        print(f"Sent: \"{text}\" + Enter")
        print(f"Final ack: {json.loads(ack).get('status')}")


def char_to_key(ch):
    """Map a character to (key_name, modifiers) for USB HID."""
    if ch.isalpha():
        if ch.isupper():
            return ch.lower(), ["shift"]
        return ch, []
    if ch.isdigit():
        return ch, []
    special = {
        " ": ("spacebar", []),
        "-": ("hyphen", []),
        ":": ("semicolon", ["shift"]),
        ".": ("period", []),
        ",": ("comma", []),
        "/": ("slash", []),
        "\\": ("backslash", []),
        ";": ("semicolon", []),
        "'": ("quote", []),
        "=": ("equal_sign", []),
        "[": ("open_bracket", []),
        "]": ("close_bracket", []),
        "`": ("grave_accent_and_tilde", []),
        "\t": ("tab", []),
    }
    if ch in special:
        return special[ch]
    print(f"Warning: unmapped character '{ch}'")
    return "spacebar", []


if __name__ == "__main__":
    asyncio.run(main())

# Phone-to-Mac HID Protocol

## Transport
WebSocket over TCP, discovered via Bonjour (`_phonehid._tcp.`)

## Messages

### Phone → Mac

#### Action
```json
{"type":"action","id":"<uuid>","action":"<action_name>","payload":{}}
```

#### Keypress
```json
{"type":"action","id":"<uuid>","action":"keypress","payload":{"key":"return_or_enter","modifiers":[]}}
```

#### Sequence
```json
{"type":"action","id":"<uuid>","action":"sequence","payload":{"steps":[
  {"key":"backslash","modifiers":["control"],"delay_ms":100},
  {"key":"open_bracket","modifiers":[]}
]}}
```

### Mac → Phone

#### Ack
```json
{"type":"ack","id":"<uuid>","status":"ok"}
```

#### Error
```json
{"type":"ack","id":"<uuid>","status":"error","error":"description"}
```

#### Config
```json
{"type":"config","payload":{"server_name":"Tony's Mac","actions":[...]}}
```

## Modifier Keys
- control, option, command, shift

## Key Names
USB HID usage code names (see internal/hid/keycodes.go)

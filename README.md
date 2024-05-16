# Spotify-Server

## Socket API

The websocket is available at `ws://localhost:5050/socket`.

Once connected, you will receive `Opcode 1: Hello`.

You should send `Opcode 2: Initialize` immediately after receiving Opcode 1.

### List of Opcodes
| Opcode | Name         | Description                                             | Client Send/Receive |
| ------ | ------------ | ------------------------------------------------------- | ---------------- |
| 0      | Dispatch     | Default Opcode when receiving core events.              | Receive only |
| 1      | Hello        | Sends this when clients initially connect               | Receive only |
| 2      | Initialize   | This is what the client sends when receiving opcode `1` | Send only |
| 3      | Heartbeat    | Clients should send Opcode 3                            | Send / Receive | 
| 4      | HeartbeatACK | Sends when clients sends heartbeat                      | Receive only |
| 5      | Error        | Sent to the client when an error occurs                 | Receive only |

### Events

Events are received on `Opcode 0: Event` - the event type will be part of the root message object under the `t` key.

#### Example Event Message Objects

#### `INITIAL_STATE`
```json
{
  "op": 0,
  "t": "INITIAL_STATE",
  "d": {
    "id": "track id",
    "title": "track title",
    "url": "track url",
    "is_playing": true,
    "artist": {
      "name": "artist name",
      "url": "artist spotify url"
    },
    "album": {
      "name": "album name",
      "url": "album spotify url",
      "art_url": "album art url"
    },
    "timestamp": { // maybe undefined
      "progress": 123,
      "duration": 224747
    }
  }
}
```

#### `TRACK_CHANGE`
```json
{
 "op": 2,
 "t": "TRACK_CHANGE",
 "d": {
  // same object of INITIAL_STATE
 } 
}
```

#### `TRACK_PROGRESS`
```json
{
  "op": 2,
  "t": "TRACK_PROGRESS",
  "d": 728 // now_playing -> progress_ms
}
```

#### `TRACK_STATE`
```json
{
  "op": 2,
  "t": "TRACK_STATE",
  "d": {
    "is_playing": true // true/false
  }
}
```

### Error Codes

Server can disconnect clients for multiple reasons, usually to do with messages being badly formatted. Please refer to your WebSocket client to see how you should handle errors - they do not get received as regular messages.

#### Types of Errors
| Name                    | Code |
| ----------------------- | ---- |
| Invalid/Unknown Opcode  | 4001 |
| Invalid message/payload | 4002 |
| Not Authenticated       | 4003 |
| By Server Request       | 4004 |
| Already authenticated   | 4005 |

## Todo

- [ ] API doc
- [ ] Config Doc
- [ ] Track object doc
- [ ] Track state object doc
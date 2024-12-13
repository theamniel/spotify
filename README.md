# Spotify-Server

## Socket API

The websocket is available at `ws://localhost:5050/socket`.

Once connected, you will receive `Opcode 1: Hello`.

You should send `Opcode 2: Initialize` immediately after receiving Opcode 1.

### Configuration
It is configured using the file [spotify-server.toml](./bin/spotify-server.toml) to avoid recompiling in exchange of some variable.

Additionally, using `#{var}` replaces it with values that are in the environment variable.

```toml
[server]
host = "localhost"
port = "5050"
token = "#{SPOTIFY_SERVER_TOKEN}"
prefork = false
timeZone = "#{SPOTIFY_SERVER_TIMEZONE}"

[grpc]
host = "localhost"
port = "4040"

[socket]
origins = ["*"]
readBufferSize = 2048
writeBufferSize = 2048

[spotify]
clientID = "#{SPOTIFY_CLIENT_ID}"
clientSecret = "#{SPOTIFY_CLIENT_SECRET}"
refreshToken = "#{SPOTIFY_REFRESH_TOKEN}"
```

#### Configuration types
| Name | Type | Descrption |
| ---- | ---- | ---------- |
| server.host | `String` | The host to listen on. |
| server.port | `String` | The port to listen on. |
| server.prefork | `Boolean` | Whether to use preforking. |
| server.timeZone | `String` | The time zone to use. |
| grpc.host | `String` | The host to listen on for gRPC. |
| grpc.port | `String` | The port to listen on for gRPC. |
| socket.origins | `Array` | The origins to allow. |
| socket.readBufferSize | `Integer` | The read buffer size. |
| socket.writeBufferSize | `Integer` | The write buffer size. |
| spotify.clientID | `String` | The Spotify client ID. |
| spotify.clientSecret | `String` | The Spotify client secret. |
| spotify.refreshToken | `String` | The Spotify refresh token. |


### Opcodes
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
##### `INITIAL_STATE`
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
      "id": "album id"
      "name": "album name",
      "url": "album spotify url",
      "art_url": "album art url"
    },
    "timestamp?": {
      "progress": 123,
      "duration": 224747
    }
  }
}
```

##### `TRACK_CHANGE`
Triggers when the song changes returning the object of the new song
```json
{
 "op": 2,
 "t": "TRACK_CHANGE",
 "d": {
  "id": "track id",
  "title": "track title",
  "...": "..."
 } 
}
```

##### `TRACK_PROGRESS`
It fires each in the 5-second range with the current progress of the song
```json
{
  "op": 2,
  "t": "TRACK_PROGRESS",
  "d": 728
}
```

### Error Codes
Server can disconnect clients for multiple reasons, usually to do with messages being badly formatted. Please refer to your WebSocket client to see how you should handle errors - they do not get received as regular messages.

#### Errors
| Name                    | Code |
| ----------------------- | ---- |
| Invalid/Unknown Opcode  | 4001 |
| Invalid message/payload | 4002 |
| Not Authenticated       | 4003 |
| By Server Request       | 4004 |
| Already authenticated   | 4005 |

### API Doc
#### `GET` /now-playing
Retrive the information player state.

#### `Queries`
| Name | Type | Description |
| ------ | --------- | ----------------------------------------------- |
| `raw`  | `boolean` | raw output directly from spotify ([see spotify documentation](https://developer.spotify.com/documentation/web-api/reference/get-information-about-the-users-current-playback)) |
| `open` | `boolean` | Redirects to the URL of the song                |

eg:
```json
{
  "album": {
    "image_url": "https://i.scdn.co/image/ab67616d0000b273b3de5764cc02f94714487c86",
    "name": "ily (i love you baby) (feat. Emilee)",
    "id": "4MHHajvRTUHItDsvfdIC8B",
    "url": "https://open.spotify.com/album/4MHHajvRTUHItDsvfdIC8B"
  },
  "artists": [
    {
      "name": "Surf Mesa",
      "url": "https://open.spotify.com/artist/1lmU3giNF3CSbkVSQmLpHQ"
    },
    {
      "name": "Emilee",
      "url": "https://open.spotify.com/artist/4ArPQ1Opcksbbf3CPwEjWE"
    }
  ],
  "id": "62aP9fBQKYKxi7PDXwcUAS",
  "is_playing": true,
  "timestamp": {
    "progress": 16338,
    "duration": 176546
  },
  "title": "ily (i love you baby) (feat. Emilee)",
  "url": "https://open.spotify.com/track/62aP9fBQKYKxi7PDXwcUAS"
}
```

#### `GET` /recently-played
Retrive the information of recently played songs.

#### `Queries`
| Name | Type | Description |
| ------ | --------- | ----------------------------------------------- |
| `raw`  | `boolean` | raw output of first track directly from spotify ([see spotify documentation](https://developer.spotify.com/documentation/web-api/reference/get-recently-played)) |
| `open` | `boolean` | Redirects to the URL of the first song                |

eg:
```json
{
  "album": {
    "image_url": "https://i.scdn.co/image/ab67616d0000b273b3de5764cc02f94714487c86",
    "name": "ily (i love you baby) (feat. Emilee)",
    "id": "4MHHajvRTUHItDsvfdIC8B",
    "url": "https://open.spotify.com/album/4MHHajvRTUHItDsvfdIC8B"
  },
  "artists": [
    {
      "name": "Surf Mesa",
      "url": "https://open.spotify.com/artist/1lmU3giNF3CSbkVSQmLpHQ"
    },
    {
      "name": "Emilee",
      "url": "https://open.spotify.com/artist/4ArPQ1Opcksbbf3CPwEjWE"
    }
  ],
  "id": "62aP9fBQKYKxi7PDXwcUAS",
  "is_playing": false,
  "played_at": "2024-07-08T22:03:03.308Z",
  "title": "ily (i love you baby) (feat. Emilee)",
  "url": "https://open.spotify.com/track/62aP9fBQKYKxi7PDXwcUAS"
}
```

# License
Spotify-server is under the license Apache License 2.0, read [here](./LICENSE) for more information.

# Disclaimer
This project is not affiliated with or endorsed by Spotify. It is a fan-created project and does not have the official backing of the company.

All rights to the music, images, and other materials used in this project belong to their respective owners. Spotify® and its logos are registered trademarks of Spotify AB.

This project is used solely for entertainment purposes and has no commercial intent. No copyright or intellectual property infringement is intended.

If you have any questions or concerns about this project, please contact the developers.

For more information about Spotify, please visit the official website: https://developer.spotify.com

It is strongly recommended that you use the official Spotify app for the best music experience.

Thank you for your understanding!
# Starship Sorades Server

A Go HTTP server that serves the game and provides WebRTC signaling for multiplayer.

## Building

```bash
cd server
go build -o starship-server .
```

## Running

```bash
# Run from project root
./server/starship-server -port 8080 -static .

# Or run with defaults
./server/starship-server
```

## Endpoints

### Static Files
- `GET /` - Serves game files (index.html, game.js, etc.)

### WebRTC Signaling
- `GET /api/signal?room=ROOM_ID&peer=PEER_ID` - SSE connection for receiving signaling messages
- `POST /api/signal?room=ROOM_ID&peer=PEER_ID` - Send signaling message

### Utility
- `GET /api/rooms` - List active rooms (for lobby)
- `GET /api/health` - Health check

## WebRTC Signaling Protocol

### Message Types

```json
{
  "type": "join|leave|offer|answer|candidate",
  "roomId": "game-session-id",
  "peerId": "sender-peer-id",
  "targetId": "target-peer-id (optional)",
  "payload": { /* SDP or ICE candidate */ },
  "timestamp": 1234567890
}
```

### Client Usage

1. Connect to SSE endpoint:
```javascript
const eventSource = new EventSource('/api/signal?room=my-room&peer=my-id');
eventSource.onmessage = (event) => {
  const msg = JSON.parse(event.data);
  handleSignalingMessage(msg);
};
```

2. Send signaling messages:
```javascript
fetch('/api/signal?room=my-room&peer=my-id', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    type: 'offer',
    targetId: 'other-peer-id',
    payload: { sdp: rtcOffer }
  })
});
```

### Connection Flow

1. Peer A joins room → receives `peers` list
2. Peer A creates RTCPeerConnection for each existing peer
3. Peer A sends `offer` to each peer
4. Peer B receives `offer`, creates answer, sends `answer`
5. Both peers exchange `candidate` messages
6. Connection established

## Development

The server uses Server-Sent Events (SSE) instead of WebSockets for simplicity and better compatibility.
Stale peers are automatically cleaned up after 60 seconds of inactivity.

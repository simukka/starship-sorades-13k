package game

import (
	"encoding/json"
	"math"
	"strconv"
	"time"

	"github.com/gopherjs/gopherjs/js"
)

// Network constants
const (
	MaxPlayers        = 20
	NetworkTickRate   = 50 * time.Millisecond // 20 Hz state broadcast
	InputTickRate     = 16 * time.Millisecond // 60 Hz input send
	InterpolationTime = 100 * time.Millisecond
)

// MessageType identifies the type of network message
type MessageType string

const (
	MsgPlayerInput    MessageType = "input"
	MsgWorldState     MessageType = "state"
	MsgPlayerJoin     MessageType = "join"
	MsgPlayerLeave    MessageType = "leave"
	MsgSpawnEnemy     MessageType = "enemy"
	MsgSpawnBonus     MessageType = "bonus"
	MsgDamage         MessageType = "damage"
	MsgHostMigrate    MessageType = "migrate"
	MsgSpawnExplosion MessageType = "explosion"
)

// NetworkMessage is the base message structure
type NetworkMessage struct {
	Type      MessageType     `json:"t"`
	PlayerID  string          `json:"p"`
	Timestamp int64           `json:"ts"`
	Data      json.RawMessage `json:"d,omitempty"`
}

// PlayerInputData contains player input state
type PlayerInputData struct {
	Keys     uint16  `json:"k"`  // Bitmask of pressed keys
	Angle    float64 `json:"a"`  // Ship angle
	Firing   bool    `json:"f"`  // Is firing
	TargetID int     `json:"ti"` // Target enemy index (-1 if none)
	SeqNum   uint32  `json:"s"`  // Sequence number for reconciliation
}

// Key bitmasks for compact input encoding
const (
	KeyLeft  uint16 = 1 << 0
	KeyRight uint16 = 1 << 1
	KeyUp    uint16 = 1 << 2
	KeyDown  uint16 = 1 << 3
	KeyFire  uint16 = 1 << 4
	KeyLock  uint16 = 1 << 5
)

// ShipState contains networked ship state
type ShipState struct {
	ID       string  `json:"id"`
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	VelX     float64 `json:"vx"`
	VelY     float64 `json:"vy"`
	Angle    float64 `json:"a"`
	Health   int     `json:"h"`
	Shield   int     `json:"s"`
	Weapons  int     `json:"w"`
	Points   int     `json:"pt"`
	InBase   bool    `json:"ib"`
	TargetID int     `json:"ti"` // Target enemy index
}

// EnemyState contains networked enemy state
type EnemyState struct {
	ID     int       `json:"id"`
	Kind   EnemyKind `json:"k"`
	X      float64   `json:"x"`
	Y      float64   `json:"y"`
	VelX   float64   `json:"vx"`
	VelY   float64   `json:"vy"`
	Health int       `json:"h"`
	Angle  float64   `json:"a"`
}

// BulletState contains networked bullet state
type BulletState struct {
	ID   int        `json:"id"`
	Kind BulletKind `json:"k"`
	X    float64    `json:"x"`
	Y    float64    `json:"y"`
	VelX float64    `json:"vx"`
	VelY float64    `json:"vy"`
	T    int        `json:"t"` // Lifetime remaining
	E    int        `json:"e"` // Health (for torpedoes)
}

// ExplosionState contains networked explosion state
type ExplosionState struct {
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	Size  float64 `json:"s"`
	Angle float64 `json:"a"`
	D     float64 `json:"d"` // Rotation speed
	Alpha float64 `json:"alpha"`
}

// WorldStateData contains the full world state from host
type WorldStateData struct {
	Tick       uint32           `json:"t"`  // Server tick number
	Ships      []ShipState      `json:"s"`  // All player ships
	Enemies    []EnemyState     `json:"e"`  // All enemies
	Bullets    []BulletState    `json:"b"`  // Active bullets/torpedos
	Explosions []ExplosionState `json:"ex"` // Active explosions
	InputAck   uint32           `json:"ia"` // Last processed input seq for this player
}

// PlayerJoinData contains info about a joining player
type PlayerJoinData struct {
	PlayerID string `json:"id"`
	Name     string `json:"name"`
	IsHost   bool   `json:"host"`
}

// SpawnEnemyData contains enemy spawn info from host
type SpawnEnemyData struct {
	Kind   EnemyKind `json:"k"`
	X      float64   `json:"x"`
	Y      float64   `json:"y"`
	Health int       `json:"h"`
}

// DamageData contains damage event info
type DamageData struct {
	TargetType string `json:"tt"` // "ship" or "enemy"
	TargetID   string `json:"id"`
	Damage     int    `json:"d"`
	SourceID   string `json:"src"`
}

// NetworkManager handles all multiplayer networking
type NetworkManager struct {
	game      *Game
	playerID  string
	roomID    string
	isHost    bool
	connected bool

	// Connections
	peers     map[string]*PeerConnection
	signaling *js.Object // EventSource for signaling

	// ICE configuration (fetched from server)
	iceConfig map[string]interface{}

	// State
	serverTick    uint32
	inputSeqNum   uint32
	lastInputSent time.Time
	lastStateSent time.Time
	pendingInputs []PlayerInputData // Unacknowledged inputs for reconciliation

	// Interpolation buffers for other players
	stateBuffer []WorldStateData
	interpTime  float64
}

// PeerConnection wraps a WebRTC peer connection
type PeerConnection struct {
	ID                string
	conn              *js.Object // RTCPeerConnection
	dataChannel       *js.Object // RTCDataChannel
	isConnected       bool
	lastReceived      time.Time
	remoteDescSet     bool                     // Whether remote description has been set
	pendingCandidates []map[string]interface{} // Buffered candidates waiting for remote desc
}

// NewNetworkManager creates a new network manager
func NewNetworkManager(game *Game) *NetworkManager {
	nm := &NetworkManager{
		game:          game,
		peers:         make(map[string]*PeerConnection),
		pendingInputs: make([]PlayerInputData, 0, 64),
		stateBuffer:   make([]WorldStateData, 0, 10),
	}
	// Fetch ICE config from server
	nm.fetchICEConfig()
	return nm
}

// fetchICEConfig retrieves ICE server configuration from the server
func (nm *NetworkManager) fetchICEConfig() {
	// Make synchronous XHR request to get ICE servers
	xhr := js.Global.Get("XMLHttpRequest").New()
	xhr.Call("open", "GET", "/api/ice-servers", false) // false = synchronous
	xhr.Call("send")

	if xhr.Get("status").Int() == 200 {
		response := xhr.Get("responseText").String()
		var config map[string]interface{}
		if err := json.Unmarshal([]byte(response), &config); err == nil {
			nm.iceConfig = config
			netDebug("Fetched ICE config from server")
			return
		}
	}

	// Fallback to default config
	netDebug("Using default ICE config")
	nm.iceConfig = map[string]interface{}{
		"iceServers": []interface{}{
			map[string]interface{}{
				"urls": "stun:stun.l.google.com:19302",
			},
		},
	}
}

// getICEConfig returns the ICE configuration for peer connections
func (nm *NetworkManager) getICEConfig() map[string]interface{} {
	if nm.iceConfig != nil {
		return nm.iceConfig
	}
	// Fallback
	return map[string]interface{}{
		"iceServers": []interface{}{
			map[string]interface{}{
				"urls": "stun:stun.l.google.com:19302",
			},
		},
	}
}

// GeneratePlayerID creates a random player ID
func GeneratePlayerID() string {
	chars := "abcdefghijklmnopqrstuvwxyz0123456789"
	id := make([]byte, 8)
	for i := range id {
		id[i] = chars[int(js.Global.Get("Math").Call("random").Float()*float64(len(chars)))]
	}
	return string(id)
}

// JoinRoom connects to a game room
func (nm *NetworkManager) JoinRoom(roomID string) {
	nm.roomID = roomID
	nm.playerID = GeneratePlayerID()

	// Connect to signaling server via SSE
	url := "/api/signal?room=" + roomID + "&peer=" + nm.playerID
	nm.signaling = js.Global.Get("EventSource").New(url)

	nm.signaling.Set("onmessage", func(event *js.Object) {
		data := event.Get("data").String()
		nm.handleSignalingMessage(data)
	})

	nm.signaling.Set("onerror", func(event *js.Object) {
		netDebug("Signaling connection error")
		nm.connected = false
	})

	nm.signaling.Set("onopen", func(event *js.Object) {
		netDebug("Connected to signaling server")
	})
}

// handleSignalingMessage processes messages from the signaling server
func (nm *NetworkManager) handleSignalingMessage(data string) {
	var msg map[string]interface{}
	if err := json.Unmarshal([]byte(data), &msg); err != nil {
		netDebug("Failed to parse signaling message: " + err.Error())
		return
	}

	msgType, _ := msg["type"].(string)
	netDebug("Signaling message received: " + msgType)

	switch msgType {
	case "peers":
		// Initial peer list - we're the host if we're the only one
		peers, _ := msg["peers"].([]interface{})
		nm.isHost = len(peers) == 1
		if nm.isHost {
			netDebug("We are the host")
		}
		// Connect to existing peers
		for _, p := range peers {
			peerID, _ := p.(string)
			if peerID != nm.playerID {
				nm.createPeerConnection(peerID, true)
			}
		}
		nm.connected = true

	case "join":
		// New peer joined - create connection (they will send offer)
		peerID, _ := msg["peerId"].(string)
		if peerID != nm.playerID {
			netDebug("Peer joined: " + peerID)
			nm.createPeerConnection(peerID, false)
		}

	case "leave":
		peerID, _ := msg["peerId"].(string)
		nm.removePeer(peerID)

	case "offer":
		nm.handleOffer(msg)

	case "answer":
		nm.handleAnswer(msg)

	case "candidate":
		nm.handleCandidate(msg)
	}
}

// createPeerConnection creates a new WebRTC connection to a peer
func (nm *NetworkManager) createPeerConnection(peerID string, initiator bool) {
	// Use ICE servers from server if available, otherwise fallback to defaults
	config := nm.getICEConfig()

	// Debug: log the ICE config being used
	configJSON, _ := json.Marshal(config)
	netDebug("Using ICE config: " + string(configJSON))

	pc := js.Global.Get("RTCPeerConnection").New(config)

	peer := &PeerConnection{
		ID:   peerID,
		conn: pc,
	}
	nm.peers[peerID] = peer

	// Handle ICE candidates
	pc.Set("onicecandidate", func(event *js.Object) {
		candidate := event.Get("candidate")
		if candidate != nil && candidate != js.Undefined {
			candidateStr := candidate.Get("candidate").String()
			netDebug("ICE candidate: " + candidateStr)

			// Convert JS object to Go map for proper JSON serialization
			candidateJSON := candidate.Call("toJSON")
			payload := map[string]interface{}{
				"candidate":     candidateJSON.Get("candidate").String(),
				"sdpMid":        candidateJSON.Get("sdpMid").String(),
				"sdpMLineIndex": candidateJSON.Get("sdpMLineIndex").Int(),
			}
			// Only include usernameFragment if present
			if uf := candidateJSON.Get("usernameFragment"); uf != nil && uf != js.Undefined {
				payload["usernameFragment"] = uf.String()
			}

			nm.sendSignaling(map[string]interface{}{
				"type":     "candidate",
				"targetId": peerID,
				"payload":  payload,
			})
		} else {
			netDebug("ICE gathering complete")
		}
	})

	// Handle ICE connection state changes
	pc.Set("oniceconnectionstatechange", func() {
		state := pc.Get("iceConnectionState").String()
		netDebug("ICE connection state: " + state)
	})

	// Handle ICE gathering state changes
	pc.Set("onicegatheringstatechange", func() {
		state := pc.Get("iceGatheringState").String()
		netDebug("ICE gathering state: " + state)
	})

	// Handle connection state
	pc.Set("onconnectionstatechange", func() {
		state := pc.Get("connectionState").String()
		netDebug("Connection to " + peerID + ": " + state)
		peer.isConnected = state == "connected"
	})

	// Handle data channel
	pc.Set("ondatachannel", func(event *js.Object) {
		channel := event.Get("channel")
		nm.setupDataChannel(peer, channel)
	})

	// If we're the initiator, create data channel and send offer
	if initiator {
		channelConfig := map[string]interface{}{
			"ordered": false, // Unreliable for low latency
		}
		channel := pc.Call("createDataChannel", "game", channelConfig)
		nm.setupDataChannel(peer, channel)

		// Create and send offer
		pc.Call("createOffer").Call("then", func(offer *js.Object) {
			pc.Call("setLocalDescription", offer).Call("then", func() {
				nm.sendSignaling(map[string]interface{}{
					"type":     "offer",
					"targetId": peerID,
					"payload": map[string]interface{}{
						"type": offer.Get("type").String(),
						"sdp":  offer.Get("sdp").String(),
					},
				})
			})
		})
	}
}

// setupDataChannel configures a data channel for game messages
func (nm *NetworkManager) setupDataChannel(peer *PeerConnection, channel *js.Object) {
	peer.dataChannel = channel

	channel.Set("onopen", func() {
		netDebug("Data channel open to " + peer.ID)
		peer.isConnected = true

		// When data channel opens, send join message to announce ourselves
		joinData, _ := json.Marshal(PlayerJoinData{
			PlayerID: nm.playerID,
			Name:     "Player " + nm.playerID[:4],
			IsHost:   nm.isHost,
		})
		msg := &NetworkMessage{
			Type:      MsgPlayerJoin,
			PlayerID:  nm.playerID,
			Timestamp: time.Now().UnixMilli(),
			Data:      joinData,
		}
		nm.sendTo(peer.ID, msg)

		// If we're the host, create a ship for this new peer
		if nm.isHost {
			// Check if ship already exists for this peer
			shipExists := false
			for _, s := range nm.game.Ships {
				if s.NetworkID == peer.ID {
					shipExists = true
					break
				}
			}

			if !shipExists {
				ship := &Ship{
					NetworkID: peer.ID,
					X:         nm.game.Ship.X + (js.Global.Get("Math").Call("random").Float()-0.5)*200,
					Y:         nm.game.Ship.Y + (js.Global.Get("Math").Call("random").Float()-0.5)*200,
					E:         100,
					local:     false,
					Shield:    Shield{MaxT: ShipMaxShield},
					Image:     nm.game.Ship.Image,
				}
				ship.Shield.Image = nm.game.Ship.Shield.Image
				ship.AddWeapon()
				nm.game.Ships = append(nm.game.Ships, ship)
				netDebug("Host created ship for peer " + peer.ID)
			}
		}
	})

	channel.Set("onclose", func() {
		netDebug("Data channel closed to " + peer.ID)
		peer.isConnected = false
	})

	channel.Set("onmessage", func(event *js.Object) {
		data := event.Get("data").String()
		nm.handleGameMessage(peer.ID, data)
	})
}

// handleOffer processes an SDP offer from a peer
func (nm *NetworkManager) handleOffer(msg map[string]interface{}) {
	peerID, _ := msg["peerId"].(string)
	payload, _ := msg["payload"].(map[string]interface{})

	netDebug("Received offer from " + peerID)

	peer, exists := nm.peers[peerID]
	if !exists {
		nm.createPeerConnection(peerID, false)
		peer = nm.peers[peerID]
	}

	sdp := map[string]interface{}{
		"type": payload["type"],
		"sdp":  payload["sdp"],
	}

	peer.conn.Call("setRemoteDescription", sdp).Call("then", func() {
		netDebug("Set remote description from " + peerID)
		peer.remoteDescSet = true
		// Process any buffered candidates
		nm.processPendingCandidates(peer)

		peer.conn.Call("createAnswer").Call("then", func(answer *js.Object) {
			peer.conn.Call("setLocalDescription", answer).Call("then", func() {
				netDebug("Sending answer to " + peerID)
				nm.sendSignaling(map[string]interface{}{
					"type":     "answer",
					"targetId": peerID,
					"payload": map[string]interface{}{
						"type": answer.Get("type").String(),
						"sdp":  answer.Get("sdp").String(),
					},
				})
			})
		})
	}).Call("catch", func(err *js.Object) {
		netDebug("Error setting remote description: " + err.Call("toString").String())
	})
}

// handleAnswer processes an SDP answer from a peer
func (nm *NetworkManager) handleAnswer(msg map[string]interface{}) {
	peerID, _ := msg["peerId"].(string)
	payload, _ := msg["payload"].(map[string]interface{})

	netDebug("Received answer from " + peerID)

	peer, exists := nm.peers[peerID]
	if !exists {
		netDebug("No peer connection for " + peerID)
		return
	}

	sdp := map[string]interface{}{
		"type": payload["type"],
		"sdp":  payload["sdp"],
	}

	peer.conn.Call("setRemoteDescription", sdp).Call("then", func() {
		netDebug("Set remote description (answer) from " + peerID)
		peer.remoteDescSet = true
		// Process any buffered candidates
		nm.processPendingCandidates(peer)
	}).Call("catch", func(err *js.Object) {
		netDebug("Error setting answer: " + err.Call("toString").String())
	})
}

// handleCandidate processes an ICE candidate from a peer
func (nm *NetworkManager) handleCandidate(msg map[string]interface{}) {
	peerID, _ := msg["peerId"].(string)
	payload, _ := msg["payload"].(map[string]interface{})

	peer, exists := nm.peers[peerID]
	if !exists {
		netDebug("Received candidate but no peer for " + peerID)
		return
	}

	// If remote description not set yet, buffer the candidate
	if !peer.remoteDescSet {
		netDebug("Buffering ICE candidate from " + peerID + " (remote desc not set)")
		peer.pendingCandidates = append(peer.pendingCandidates, payload)
		return
	}

	// Add candidate immediately
	nm.addIceCandidate(peer, payload)
}

// processPendingCandidates adds all buffered candidates after remote description is set
func (nm *NetworkManager) processPendingCandidates(peer *PeerConnection) {
	if len(peer.pendingCandidates) > 0 {
		netDebug("Processing " + strconv.Itoa(len(peer.pendingCandidates)) + " buffered candidates for " + peer.ID)
		for _, candidate := range peer.pendingCandidates {
			nm.addIceCandidate(peer, candidate)
		}
		peer.pendingCandidates = nil
	}
}

// addIceCandidate adds a single ICE candidate to the peer connection
func (nm *NetworkManager) addIceCandidate(peer *PeerConnection, payload map[string]interface{}) {
	// Debug: log the payload structure
	payloadJSON, _ := json.Marshal(payload)
	netDebug("addIceCandidate payload: " + string(payloadJSON))

	if candidate, ok := payload["candidate"].(string); ok && len(candidate) > 0 {
		// Truncate for logging
		displayLen := len(candidate)
		if displayLen > 50 {
			displayLen = 50
		}
		netDebug("Adding ICE candidate from " + peer.ID + ": " + candidate[:displayLen])
	} else {
		netDebug("No candidate string in payload")
	}

	peer.conn.Call("addIceCandidate", payload).Call("then", func() {
		netDebug("ICE candidate added successfully")
	}).Call("catch", func(err *js.Object) {
		netDebug("Error adding ICE candidate: " + err.Call("toString").String())
	})
}

// removePeer cleans up a disconnected peer
func (nm *NetworkManager) removePeer(peerID string) {
	peer, exists := nm.peers[peerID]
	if !exists {
		return
	}

	if peer.dataChannel != nil {
		peer.dataChannel.Call("close")
	}
	if peer.conn != nil {
		peer.conn.Call("close")
	}

	delete(nm.peers, peerID)
	netDebug("Removed peer: " + peerID)

	// Remove ship from game
	for i, ship := range nm.game.Ships {
		if !ship.local && ship.NetworkID == peerID {
			nm.game.Ships = append(nm.game.Ships[:i], nm.game.Ships[i+1:]...)
			break
		}
	}

	// TODO: Host migration if host left
}

// sendSignaling sends a message via the signaling server
func (nm *NetworkManager) sendSignaling(msg map[string]interface{}) {
	data, _ := json.Marshal(msg)

	js.Global.Call("fetch", "/api/signal?room="+nm.roomID+"&peer="+nm.playerID, map[string]interface{}{
		"method": "POST",
		"headers": map[string]interface{}{
			"Content-Type": "application/json",
		},
		"body": string(data),
	})
}

// broadcast sends a game message to all connected peers
func (nm *NetworkManager) broadcast(msg *NetworkMessage) {
	data, _ := json.Marshal(msg)
	dataStr := string(data)

	for _, peer := range nm.peers {
		if peer.isConnected && peer.dataChannel != nil {
			peer.dataChannel.Call("send", dataStr)
		}
	}
}

// sendTo sends a game message to a specific peer
func (nm *NetworkManager) sendTo(peerID string, msg *NetworkMessage) {
	peer, exists := nm.peers[peerID]
	if !exists || !peer.isConnected || peer.dataChannel == nil {
		return
	}

	data, _ := json.Marshal(msg)
	peer.dataChannel.Call("send", string(data))
}

// handleGameMessage processes a game message from a peer
func (nm *NetworkManager) handleGameMessage(peerID string, data string) {
	var msg NetworkMessage
	if err := json.Unmarshal([]byte(data), &msg); err != nil {
		return
	}

	switch msg.Type {
	case MsgPlayerInput:
		if nm.isHost {
			nm.handlePlayerInput(peerID, msg.Data)
		}

	case MsgWorldState:
		if !nm.isHost {
			nm.handleWorldState(msg.Data)
		}

	case MsgPlayerJoin:
		nm.handlePlayerJoin(peerID, msg.Data)

	case MsgDamage:
		nm.handleDamage(msg.Data)
	}
}

// handlePlayerInput processes input from a client (host only)
func (nm *NetworkManager) handlePlayerInput(peerID string, data json.RawMessage) {
	var input PlayerInputData
	if err := json.Unmarshal(data, &input); err != nil {
		return
	}

	// Find or create ship for this player
	var ship *Ship
	for _, s := range nm.game.Ships {
		if s.NetworkID == peerID {
			ship = s
			break
		}
	}

	if ship == nil {
		return
	}

	// Apply input to ship
	nm.applyInputToShip(ship, &input)
}

// applyInputToShip applies network input to a ship
func (nm *NetworkManager) applyInputToShip(ship *Ship, input *PlayerInputData) {
	// Rotation
	if input.Keys&KeyLeft != 0 {
		ship.Angle -= ShipRotationSpeed
	}
	if input.Keys&KeyRight != 0 {
		ship.Angle += ShipRotationSpeed
	}

	// Thrust
	if input.Keys&KeyUp != 0 {
		ship.VelX += math.Sin(ship.Angle) * ShipThrustAcc
		ship.VelY -= math.Cos(ship.Angle) * ShipThrustAcc
	}
	if input.Keys&KeyDown != 0 {
		ship.VelX -= math.Sin(ship.Angle) * ShipThrustAcc * 0.5
		ship.VelY += math.Cos(ship.Angle) * ShipThrustAcc * 0.5
	}

	// Clamp velocity
	speed := math.Sqrt(ship.VelX*ship.VelX + ship.VelY*ship.VelY)
	if speed > ShipMaxSpeed {
		scale := ShipMaxSpeed / speed
		ship.VelX *= scale
		ship.VelY *= scale
	}

	// Apply velocity
	ship.X += ship.VelX
	ship.Y += ship.VelY

	// Apply drag
	ship.VelX *= ShipACCFactor
	ship.VelY *= ShipACCFactor

	// Handle firing (host processes client fire input)
	if input.Firing {
		ship.Fire(nm.game)
	}

	// Handle target lock (T key)
	if input.Keys&KeyLock != 0 {
		ship.InitiateTargetLock(nm.game)
	}

	// Set target if provided
	if input.TargetID >= 0 && input.TargetID < len(nm.game.Enemies) {
		ship.Target = nm.game.Enemies[input.TargetID]
	}
}

// handleWorldState processes world state from host (clients only)
func (nm *NetworkManager) handleWorldState(data json.RawMessage) {
	var state WorldStateData
	if err := json.Unmarshal(data, &state); err != nil {
		return
	}

	// Add to interpolation buffer
	nm.stateBuffer = append(nm.stateBuffer, state)
	if len(nm.stateBuffer) > 10 {
		nm.stateBuffer = nm.stateBuffer[1:]
	}

	// Reconcile local player's ship
	nm.reconcileLocalShip(&state)

	// Update remote ships
	nm.updateRemoteShips(&state)

	// Update enemies
	nm.updateEnemies(&state)

	// Update bullets (so clients see all projectiles)
	nm.updateBullets(&state)

	// Update explosions (so clients see all explosions)
	nm.updateExplosions(&state)
}

// reconcileLocalShip handles server reconciliation for local player
func (nm *NetworkManager) reconcileLocalShip(state *WorldStateData) {
	// Find our ship in the state
	var serverShip *ShipState
	for i := range state.Ships {
		if state.Ships[i].ID == nm.playerID {
			serverShip = &state.Ships[i]
			break
		}
	}

	if serverShip == nil || nm.game.Ship == nil {
		return
	}

	// Remove acknowledged inputs
	newPending := make([]PlayerInputData, 0, len(nm.pendingInputs))
	for _, input := range nm.pendingInputs {
		if input.SeqNum > state.InputAck {
			newPending = append(newPending, input)
		}
	}
	nm.pendingInputs = newPending

	// Reset to server state
	nm.game.Ship.X = serverShip.X
	nm.game.Ship.Y = serverShip.Y
	nm.game.Ship.VelX = serverShip.VelX
	nm.game.Ship.VelY = serverShip.VelY
	nm.game.Ship.Angle = serverShip.Angle
	nm.game.Ship.E = serverShip.Health
	nm.game.Ship.Shield.T = serverShip.Shield

	// Re-apply unacknowledged inputs
	for _, input := range nm.pendingInputs {
		nm.applyInputToShip(nm.game.Ship, &input)
	}
}

// updateRemoteShips updates other players' ships with interpolation
func (nm *NetworkManager) updateRemoteShips(state *WorldStateData) {
	for _, shipState := range state.Ships {
		if shipState.ID == nm.playerID {
			continue // Skip local player
		}

		// Find or create ship
		var ship *Ship
		for _, s := range nm.game.Ships {
			if s.NetworkID == shipState.ID {
				ship = s
				break
			}
		}

		if ship == nil {
			// Create new remote ship with images from local ship
			ship = &Ship{
				NetworkID: shipState.ID,
				local:     false,
				E:         100,
				Shield:    Shield{MaxT: ShipMaxShield},
				Image:     nm.game.Ship.Image, // Copy image from local ship
			}
			ship.Shield.Image = nm.game.Ship.Shield.Image // Copy shield image
			nm.game.Ships = append(nm.game.Ships, ship)
			netDebug("Created remote ship for " + shipState.ID)
		}

		// Update ship state (with interpolation in future)
		ship.X = shipState.X
		ship.Y = shipState.Y
		ship.VelX = shipState.VelX
		ship.VelY = shipState.VelY
		ship.Angle = shipState.Angle
		ship.E = shipState.Health
		ship.Shield.T = shipState.Shield
		ship.InBase = shipState.InBase
	}
}

// updateEnemies syncs enemy state from host
func (nm *NetworkManager) updateEnemies(state *WorldStateData) {
	// Create map of existing enemies
	existing := make(map[int]*Enemy)
	for _, e := range nm.game.Enemies {
		existing[e.NetworkID] = e
	}

	// Update or create enemies
	for _, es := range state.Enemies {
		enemy, exists := existing[es.ID]
		if !exists {
			// Create new enemy
			enemy = &Enemy{
				NetworkID: es.ID,
				Kind:      es.Kind,
				Radius:    nm.game.EnemyTypes[es.Kind].R,
				Image:     nm.game.EnemyTypes[es.Kind].Image,
				MaxHealth: es.Health,
			}
			nm.game.Enemies = append(nm.game.Enemies, enemy)
		}

		enemy.X = es.X
		enemy.Y = es.Y
		enemy.VelX = es.VelX
		enemy.VelY = es.VelY
		enemy.Health = es.Health
		enemy.Angle = es.Angle
		delete(existing, es.ID)
	}

	// Remove enemies that no longer exist
	for id := range existing {
		for i, e := range nm.game.Enemies {
			if e.NetworkID == id {
				nm.game.Enemies = append(nm.game.Enemies[:i], nm.game.Enemies[i+1:]...)
				break
			}
		}
	}
}

// updateBullets syncs bullet state from host (clients only)
func (nm *NetworkManager) updateBullets(state *WorldStateData) {
	// Clear all bullets and replace with server state
	nm.game.Bullets.Clear()

	for _, bs := range state.Bullets {
		var bullet *Bullet
		if bs.Kind == StandardBullet {
			bullet = nm.game.Bullets.AcquireKind(StandardBullet)
		} else {
			bullet = nm.game.Bullets.AcquireKind(TorpedoBullet)
		}
		if bullet == nil {
			continue
		}

		bullet.X = bs.X
		bullet.Y = bs.Y
		bullet.XAcc = bs.VelX
		bullet.YAcc = bs.VelY
		bullet.T = bs.T
		bullet.E = bs.E
	}
}

// updateExplosions syncs explosion state from host (clients only)
func (nm *NetworkManager) updateExplosions(state *WorldStateData) {
	// Clear all explosions and replace with server state
	nm.game.Explosions.Clear()

	for _, es := range state.Explosions {
		exp := nm.game.Explosions.Acquire()
		if exp == nil {
			continue
		}

		exp.X = es.X
		exp.Y = es.Y
		exp.Size = es.Size
		exp.Angle = es.Angle
		exp.D = es.D
		exp.Alpha = es.Alpha
	}
}

// handlePlayerJoin processes a new player joining
func (nm *NetworkManager) handlePlayerJoin(peerID string, data json.RawMessage) {
	var joinData PlayerJoinData
	if err := json.Unmarshal(data, &joinData); err != nil {
		return
	}

	// Create ship for new player if host
	if nm.isHost {
		ship := &Ship{
			NetworkID: peerID,
			X:         nm.game.Ship.X + (js.Global.Get("Math").Call("random").Float()-0.5)*200,
			Y:         nm.game.Ship.Y + (js.Global.Get("Math").Call("random").Float()-0.5)*200,
			E:         100,
			local:     false,
			Shield:    Shield{MaxT: ShipMaxShield},
			Image:     nm.game.Ship.Image, // Copy image from local ship
		}
		ship.Shield.Image = nm.game.Ship.Shield.Image // Copy shield image
		ship.AddWeapon()
		nm.game.Ships = append(nm.game.Ships, ship)
		netDebug("Created ship for joining player " + peerID)
	}
}

// handleDamage processes damage events
func (nm *NetworkManager) handleDamage(data json.RawMessage) {
	var dmg DamageData
	if err := json.Unmarshal(data, &dmg); err != nil {
		return
	}

	if dmg.TargetType == "ship" {
		for _, ship := range nm.game.Ships {
			if ship.NetworkID == dmg.TargetID {
				ship.E -= dmg.Damage
				if ship.E < 0 {
					ship.E = 0
				}
				break
			}
		}
	}
}

// Update is called each frame to process networking
func (nm *NetworkManager) Update() {
	if !nm.connected {
		return
	}

	now := time.Now()

	if nm.isHost {
		// Host: broadcast world state periodically
		if now.Sub(nm.lastStateSent) >= NetworkTickRate {
			nm.broadcastWorldState()
			nm.lastStateSent = now
		}
	} else {
		// Client: send input periodically
		if now.Sub(nm.lastInputSent) >= InputTickRate {
			nm.sendLocalInput()
			nm.lastInputSent = now
		}
	}
}

// sendLocalInput sends local player's input to host
func (nm *NetworkManager) sendLocalInput() {
	if nm.game.Ship == nil {
		return
	}

	// Encode current keys
	var keys uint16
	if nm.game.Keys[37] {
		keys |= KeyLeft
	}
	if nm.game.Keys[39] {
		keys |= KeyRight
	}
	if nm.game.Keys[38] {
		keys |= KeyUp
	}
	if nm.game.Keys[40] {
		keys |= KeyDown
	}
	if nm.game.Keys[88] {
		keys |= KeyFire
	}
	if nm.game.Keys[84] {
		keys |= KeyLock
	}

	nm.inputSeqNum++
	input := PlayerInputData{
		Keys:   keys,
		Angle:  nm.game.Ship.Angle,
		Firing: nm.game.Keys[88],
		SeqNum: nm.inputSeqNum,
	}
	if nm.game.Ship.Target != nil {
		input.TargetID = nm.game.Ship.Target.NetworkID
	} else {
		input.TargetID = -1
	}

	// Store for reconciliation
	nm.pendingInputs = append(nm.pendingInputs, input)
	if len(nm.pendingInputs) > 64 {
		nm.pendingInputs = nm.pendingInputs[1:]
	}

	// Send to host
	inputData, _ := json.Marshal(input)
	msg := &NetworkMessage{
		Type:      MsgPlayerInput,
		PlayerID:  nm.playerID,
		Timestamp: time.Now().UnixMilli(),
		Data:      inputData,
	}

	// Find host and send
	for _, peer := range nm.peers {
		if peer.isConnected {
			nm.sendTo(peer.ID, msg)
			break // Send to first connected peer (host)
		}
	}
}

// broadcastWorldState sends current world state to all clients (host only)
func (nm *NetworkManager) broadcastWorldState() {
	nm.serverTick++

	// Collect ship states
	ships := make([]ShipState, 0, len(nm.game.Ships))
	for _, ship := range nm.game.Ships {
		targetID := -1
		if ship.Target != nil {
			targetID = ship.Target.NetworkID
		}
		ships = append(ships, ShipState{
			ID:       ship.NetworkID,
			X:        ship.X,
			Y:        ship.Y,
			VelX:     ship.VelX,
			VelY:     ship.VelY,
			Angle:    ship.Angle,
			Health:   ship.E,
			Shield:   ship.Shield.T,
			Weapons:  len(ship.Weapons),
			Points:   ship.Points,
			InBase:   ship.InBase,
			TargetID: targetID,
		})
	}

	// Collect enemy states
	enemies := make([]EnemyState, 0, len(nm.game.Enemies))
	for i, enemy := range nm.game.Enemies {
		enemy.NetworkID = i // Ensure IDs are assigned
		enemies = append(enemies, EnemyState{
			ID:     i,
			Kind:   enemy.Kind,
			X:      enemy.X,
			Y:      enemy.Y,
			VelX:   enemy.VelX,
			VelY:   enemy.VelY,
			Health: enemy.Health,
			Angle:  enemy.Angle,
		})
	}

	// Collect ALL bullet states (both standard bullets and torpedoes)
	bullets := make([]BulletState, 0, nm.game.Bullets.ActiveCount)
	for i := 0; i < nm.game.Bullets.ActiveCount; i++ {
		b := nm.game.Bullets.Pool[i]
		bullets = append(bullets, BulletState{
			ID:   i,
			Kind: b.Kind,
			X:    b.X,
			Y:    b.Y,
			VelX: b.XAcc,
			VelY: b.YAcc,
			T:    b.T,
			E:    b.E,
		})
	}

	// Collect explosion states
	explosions := make([]ExplosionState, 0, nm.game.Explosions.ActiveCount)
	for i := 0; i < nm.game.Explosions.ActiveCount; i++ {
		exp := nm.game.Explosions.Pool[i]
		explosions = append(explosions, ExplosionState{
			X:     exp.X,
			Y:     exp.Y,
			Size:  exp.Size,
			Angle: exp.Angle,
			D:     exp.D,
			Alpha: exp.Alpha,
		})
	}

	state := WorldStateData{
		Tick:       nm.serverTick,
		Ships:      ships,
		Enemies:    enemies,
		Bullets:    bullets,
		Explosions: explosions,
	}

	stateData, _ := json.Marshal(state)
	msg := &NetworkMessage{
		Type:      MsgWorldState,
		PlayerID:  nm.playerID,
		Timestamp: time.Now().UnixMilli(),
		Data:      stateData,
	}

	nm.broadcast(msg)
}

// IsHost returns whether this client is the host
func (nm *NetworkManager) IsHost() bool {
	return nm.isHost
}

// IsConnected returns whether we're connected to a room
func (nm *NetworkManager) IsConnected() bool {
	return nm.connected
}

// GetPlayerCount returns the number of connected players
func (nm *NetworkManager) GetPlayerCount() int {
	count := 1 // Self
	for _, peer := range nm.peers {
		if peer.isConnected {
			count++
		}
	}
	return count
}

// Disconnect closes all connections
func (nm *NetworkManager) Disconnect() {
	for _, peer := range nm.peers {
		if peer.dataChannel != nil {
			peer.dataChannel.Call("close")
		}
		if peer.conn != nil {
			peer.conn.Call("close")
		}
	}
	nm.peers = make(map[string]*PeerConnection)

	if nm.signaling != nil {
		nm.signaling.Call("close")
	}

	nm.connected = false
}

// GetPlayerID returns the local player's ID
func (nm *NetworkManager) GetPlayerID() string {
	return nm.playerID
}

// GetRoomID returns the current room ID
func (nm *NetworkManager) GetRoomID() string {
	return nm.roomID
}

// netDebug logs a network message (JS console)
func netDebug(msg string) {
	js.Global.Get("console").Call("log", "[Network] "+msg)
}

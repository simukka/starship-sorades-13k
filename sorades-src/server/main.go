//go:build !js
// +build !js

package main

import (
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/pion/turn/v3"
)

//go:embed index.html
var indexHTML []byte

// TURN server configuration
var (
	turnRealm    = "starship-sorades"
	turnUsername = "starship"
	turnPassword = "sorades13k"
)

// WebRTC Signaling Types

// SignalMessage represents a WebRTC signaling message
type SignalMessage struct {
	Type      string          `json:"type"`      // "offer", "answer", "candidate", "join", "leave"
	RoomID    string          `json:"roomId"`    // Room/game session ID
	PeerID    string          `json:"peerId"`    // Sender's peer ID
	TargetID  string          `json:"targetId"`  // Target peer ID (for direct messages)
	Payload   json.RawMessage `json:"payload"`   // SDP or ICE candidate data
	Timestamp int64           `json:"timestamp"` // Unix timestamp
}

// Peer represents a connected peer in a room
type Peer struct {
	ID       string
	RoomID   string
	Messages chan []byte
	LastSeen time.Time
	mu       sync.Mutex
}

// Room represents a game session
type Room struct {
	ID      string
	Peers   map[string]*Peer
	Created time.Time
	mu      sync.RWMutex
}

// SignalingServer manages WebRTC signaling
type SignalingServer struct {
	rooms map[string]*Room
	mu    sync.RWMutex
}

// NewSignalingServer creates a new signaling server
func NewSignalingServer() *SignalingServer {
	s := &SignalingServer{
		rooms: make(map[string]*Room),
	}
	// Start cleanup goroutine
	go s.cleanup()
	return s
}

// cleanup removes stale peers and empty rooms
func (s *SignalingServer) cleanup() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		for roomID, room := range s.rooms {
			room.mu.Lock()
			for peerID, peer := range room.Peers {
				if time.Since(peer.LastSeen) > 60*time.Second {
					close(peer.Messages)
					delete(room.Peers, peerID)
					log.Printf("Removed stale peer %s from room %s", peerID, roomID)
				}
			}
			if len(room.Peers) == 0 {
				delete(s.rooms, roomID)
				log.Printf("Removed empty room %s", roomID)
			}
			room.mu.Unlock()
		}
		s.mu.Unlock()
	}
}

// GetOrCreateRoom gets or creates a room
func (s *SignalingServer) GetOrCreateRoom(roomID string) *Room {
	s.mu.Lock()
	defer s.mu.Unlock()

	if room, exists := s.rooms[roomID]; exists {
		return room
	}

	room := &Room{
		ID:      roomID,
		Peers:   make(map[string]*Peer),
		Created: time.Now(),
	}
	s.rooms[roomID] = room
	log.Printf("Created room %s", roomID)
	return room
}

// AddPeer adds a peer to a room
func (s *SignalingServer) AddPeer(roomID, peerID string) *Peer {
	room := s.GetOrCreateRoom(roomID)

	room.mu.Lock()
	defer room.mu.Unlock()

	// Remove existing peer with same ID if exists
	if existing, exists := room.Peers[peerID]; exists {
		close(existing.Messages)
	}

	peer := &Peer{
		ID:       peerID,
		RoomID:   roomID,
		Messages: make(chan []byte, 100),
		LastSeen: time.Now(),
	}
	room.Peers[peerID] = peer
	log.Printf("Peer %s joined room %s", peerID, roomID)

	return peer
}

// RemovePeer removes a peer from a room
func (s *SignalingServer) RemovePeer(roomID, peerID string) {
	s.mu.RLock()
	room, exists := s.rooms[roomID]
	s.mu.RUnlock()

	if !exists {
		return
	}

	room.mu.Lock()
	defer room.mu.Unlock()

	if peer, exists := room.Peers[peerID]; exists {
		close(peer.Messages)
		delete(room.Peers, peerID)
		log.Printf("Peer %s left room %s", peerID, roomID)
	}
}

// BroadcastToRoom sends a message to all peers in a room except sender
func (s *SignalingServer) BroadcastToRoom(roomID, senderID string, msg []byte) {
	s.mu.RLock()
	room, exists := s.rooms[roomID]
	s.mu.RUnlock()

	if !exists {
		return
	}

	room.mu.RLock()
	defer room.mu.RUnlock()

	for peerID, peer := range room.Peers {
		if peerID != senderID {
			select {
			case peer.Messages <- msg:
			default:
				log.Printf("Message buffer full for peer %s", peerID)
			}
		}
	}
}

// SendToPeer sends a message to a specific peer
func (s *SignalingServer) SendToPeer(roomID, targetID string, msg []byte) {
	s.mu.RLock()
	room, exists := s.rooms[roomID]
	s.mu.RUnlock()

	if !exists {
		return
	}

	room.mu.RLock()
	peer, exists := room.Peers[targetID]
	room.mu.RUnlock()

	if exists {
		select {
		case peer.Messages <- msg:
		default:
			log.Printf("Message buffer full for peer %s", targetID)
		}
	}
}

// GetPeersInRoom returns list of peer IDs in a room
func (s *SignalingServer) GetPeersInRoom(roomID string) []string {
	s.mu.RLock()
	room, exists := s.rooms[roomID]
	s.mu.RUnlock()

	if !exists {
		return nil
	}

	room.mu.RLock()
	defer room.mu.RUnlock()

	peers := make([]string, 0, len(room.Peers))
	for peerID := range room.Peers {
		peers = append(peers, peerID)
	}
	return peers
}

// Global signaling server instance
var signaling = NewSignalingServer()

// HTTP Handlers

// handleSignaling handles WebRTC signaling via Server-Sent Events (SSE)
func handleSignaling(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	roomID := r.URL.Query().Get("room")
	peerID := r.URL.Query().Get("peer")

	if roomID == "" || peerID == "" {
		http.Error(w, "room and peer query parameters required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case "GET":
		// SSE connection for receiving messages
		handleSSE(w, r, roomID, peerID)
	case "POST":
		// Send signaling message
		handleSignalPost(w, r, roomID, peerID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleSSE handles Server-Sent Events for a peer
func handleSSE(w http.ResponseWriter, r *http.Request, roomID, peerID string) {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	// Add peer to room
	peer := signaling.AddPeer(roomID, peerID)

	// Send current peers list
	peers := signaling.GetPeersInRoom(roomID)
	peersJSON, _ := json.Marshal(map[string]interface{}{
		"type":  "peers",
		"peers": peers,
	})
	fmt.Fprintf(w, "data: %s\n\n", peersJSON)
	flusher.Flush()

	// Notify other peers about new peer
	joinMsg, _ := json.Marshal(SignalMessage{
		Type:      "join",
		RoomID:    roomID,
		PeerID:    peerID,
		Timestamp: time.Now().Unix(),
	})
	signaling.BroadcastToRoom(roomID, peerID, joinMsg)

	// Stream messages to peer
	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			// Client disconnected
			signaling.RemovePeer(roomID, peerID)

			// Notify other peers
			leaveMsg, _ := json.Marshal(SignalMessage{
				Type:      "leave",
				RoomID:    roomID,
				PeerID:    peerID,
				Timestamp: time.Now().Unix(),
			})
			signaling.BroadcastToRoom(roomID, peerID, leaveMsg)
			return

		case msg, ok := <-peer.Messages:
			if !ok {
				return
			}
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()

			// Update last seen
			peer.mu.Lock()
			peer.LastSeen = time.Now()
			peer.mu.Unlock()
		}
	}
}

// handleSignalPost handles incoming signaling messages
func handleSignalPost(w http.ResponseWriter, r *http.Request, roomID, peerID string) {
	var msg SignalMessage
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	msg.RoomID = roomID
	msg.PeerID = peerID
	msg.Timestamp = time.Now().Unix()

	msgBytes, _ := json.Marshal(msg)

	// Debug logging
	log.Printf("Signal %s -> %s: type=%s", peerID, msg.TargetID, msg.Type)

	// Route message
	if msg.TargetID != "" {
		// Direct message to specific peer
		signaling.SendToPeer(roomID, msg.TargetID, msgBytes)
	} else {
		// Broadcast to room
		signaling.BroadcastToRoom(roomID, peerID, msgBytes)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

// handleRooms returns list of active rooms (for debugging/lobby)
func handleRooms(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	signaling.mu.RLock()
	defer signaling.mu.RUnlock()

	rooms := make([]map[string]interface{}, 0, len(signaling.rooms))
	for roomID, room := range signaling.rooms {
		room.mu.RLock()
		rooms = append(rooms, map[string]interface{}{
			"id":        roomID,
			"peerCount": len(room.Peers),
			"created":   room.Created.Unix(),
		})
		room.mu.RUnlock()
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"rooms": rooms,
	})
}

func main() {
	port := flag.Int("port", 8080, "HTTP server port")
	turnPort := flag.Int("turn-port", 3478, "TURN server port")
	staticDir := flag.String("static", ".", "Directory to serve static files from")
	publicIP := flag.String("public-ip", "", "Public IP address for TURN server (auto-detected if empty)")
	flag.Parse()

	// Determine the IP to use for TURN
	turnIP := *publicIP
	if turnIP == "" {
		// Auto-detect
		if ip := getOutboundIP(); ip != nil {
			turnIP = ip.String()
		} else {
			turnIP = "127.0.0.1"
		}
	}
	log.Printf("TURN server IP: %s", turnIP)

	// Start TURN server
	go startTURNServer(*turnPort, *publicIP)

	// Serve embedded index.html at root path
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || r.URL.Path == "/index.html" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write(indexHTML)
			return
		}
		// Serve other static files from disk
		http.FileServer(http.Dir(*staticDir)).ServeHTTP(w, r)
	})

	// WebRTC signaling endpoint
	http.HandleFunc("/api/signal", handleSignaling)

	// Room list endpoint (for lobby/debugging)
	http.HandleFunc("/api/rooms", handleRooms)

	// ICE server configuration endpoint (returns TURN credentials)
	http.HandleFunc("/api/ice-servers", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		iceServers := map[string]interface{}{
			"iceServers": []interface{}{
				map[string]interface{}{
					"urls": "stun:stun.l.google.com:19302",
				},
				map[string]interface{}{
					"urls": []interface{}{
						fmt.Sprintf("turn:%s:%d", turnIP, *turnPort),
						fmt.Sprintf("turn:%s:%d?transport=tcp", turnIP, *turnPort),
					},
					"username":   turnUsername,
					"credential": turnPassword,
				},
			},
		}

		json.NewEncoder(w).Encode(iceServers)
	})

	// Health check
	http.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"healthy"}`))
	})

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Starship Sorades server starting on http://localhost%s", addr)
	log.Printf("TURN server running on port %d", *turnPort)
	log.Printf("Serving static files from: %s", *staticDir)
	log.Printf("WebRTC signaling endpoint: /api/signal?room=ROOM_ID&peer=PEER_ID")
	log.Printf("ICE servers endpoint: /api/ice-servers")

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}

// startTURNServer starts a Pion TURN server
func startTURNServer(port int, publicIP string) {
	// Create a UDP listener
	udpListener, err := net.ListenPacket("udp4", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		log.Printf("Failed to create TURN UDP listener: %v", err)
		return
	}

	// Create a TCP listener
	tcpListener, err := net.Listen("tcp4", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		log.Printf("Failed to create TURN TCP listener: %v", err)
		return
	}

	// Determine public IP
	var relayIP net.IP
	if publicIP != "" {
		relayIP = net.ParseIP(publicIP)
	} else {
		// Try to auto-detect public IP
		relayIP = getOutboundIP()
	}

	if relayIP == nil {
		log.Printf("Could not determine public IP, TURN relay may not work")
		relayIP = net.ParseIP("127.0.0.1")
	}

	log.Printf("TURN server relay IP: %s", relayIP.String())

	// Create TURN server config
	s, err := turn.NewServer(turn.ServerConfig{
		Realm: turnRealm,
		// AuthHandler is called for every TURN allocation
		AuthHandler: func(username, realm string, srcAddr net.Addr) ([]byte, bool) {
			if username == turnUsername {
				return turn.GenerateAuthKey(turnUsername, turnRealm, turnPassword), true
			}
			return nil, false
		},
		PacketConnConfigs: []turn.PacketConnConfig{
			{
				PacketConn: udpListener,
				RelayAddressGenerator: &turn.RelayAddressGeneratorStatic{
					RelayAddress: relayIP,
					Address:      "0.0.0.0",
				},
			},
		},
		ListenerConfigs: []turn.ListenerConfig{
			{
				Listener: tcpListener,
				RelayAddressGenerator: &turn.RelayAddressGeneratorStatic{
					RelayAddress: relayIP,
					Address:      "0.0.0.0",
				},
			},
		},
	})

	if err != nil {
		log.Printf("Failed to start TURN server: %v", err)
		return
	}

	log.Printf("TURN server started on UDP/TCP port %d", port)

	// Keep server running
	_ = s
}

// getOutboundIP gets the preferred outbound IP of this machine
func getOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return nil
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP
}

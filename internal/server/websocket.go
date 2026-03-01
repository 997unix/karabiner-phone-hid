package server

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/tonyjiang/karabiner-phone-hid/internal/protocol"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Server is a WebSocket server that routes messages.
type Server struct {
	router     *Router
	serverName string
	actions    []protocol.ActionInfo
	listener   net.Listener
	mu         sync.Mutex
}

// NewServer creates a WebSocket server.
func NewServer(router *Router, serverName string, actions []protocol.ActionInfo) *Server {
	return &Server{
		router:     router,
		serverName: serverName,
		actions:    actions,
	}
}

// Start begins listening. Returns the bound port. Use Stop to shut down.
func (s *Server) Start(addr string) (int, error) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleWS)

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return 0, err
	}

	s.mu.Lock()
	s.listener = ln
	s.mu.Unlock()

	port := ln.Addr().(*net.TCPAddr).Port

	go func() {
		if err := http.Serve(ln, mux); err != nil {
			// net.ErrClosed expected on Stop
			log.Printf("[WS] server stopped: %v", err)
		}
	}()

	return port, nil
}

// Stop shuts down the server.
func (s *Server) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[WS] upgrade error: %v", err)
		return
	}
	defer conn.Close()

	// Send config on connect
	cfg := protocol.NewConfigMessage(s.serverName, s.actions)
	cfgData, _ := json.Marshal(cfg)
	if err := conn.WriteMessage(websocket.TextMessage, cfgData); err != nil {
		log.Printf("[WS] config send error: %v", err)
		return
	}

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Printf("[WS] read error: %v", err)
			}
			return
		}

		resp := s.router.Route(message)
		if err := conn.WriteMessage(websocket.TextMessage, resp); err != nil {
			log.Printf("[WS] write error: %v", err)
			return
		}
	}
}

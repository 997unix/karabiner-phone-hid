package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/tonyjiang/karabiner-phone-hid/internal/config"
	"github.com/tonyjiang/karabiner-phone-hid/internal/discovery"
	"github.com/tonyjiang/karabiner-phone-hid/internal/hid"
	"github.com/tonyjiang/karabiner-phone-hid/internal/server"
)

func main() {
	port := flag.Int("port", 8765, "WebSocket listen port")
	name := flag.String("name", "", "Server name for Bonjour (default: hostname)")
	flag.Parse()

	// Check root
	if os.Geteuid() != 0 {
		fmt.Fprintln(os.Stderr, "Warning: not running as root. Karabiner VHD requires root privileges.")
		fmt.Fprintln(os.Stderr, "Run with: sudo ./bin/karabiner-phone-hid")
	}

	// Load config
	cfg, err := config.LoadActions()
	if err != nil {
		log.Printf("[Config] Error loading config: %v (using defaults)", err)
	}
	registry := config.NewRegistry(cfg)

	// Initialize Karabiner
	hid.InitGlobal()
	defer hid.CleanupGlobal()

	poster := hid.NewKarabinerPoster()
	defer poster.Close()

	log.Println("[Server] Waiting for Karabiner keyboard...")
	poster.WaitReady()
	log.Println("[Server] Keyboard ready")

	// Create dispatcher and router
	dispatcher := hid.NewDispatcher(poster)
	router := server.NewRouter(dispatcher, registry)

	// Start WebSocket server
	actions := registry.AllActions()
	serverName := *name
	if serverName == "" {
		serverName, _ = os.Hostname()
	}

	ws := server.NewServer(router, serverName, actions)
	boundPort, err := ws.Start(fmt.Sprintf("0.0.0.0:%d", *port))
	if err != nil {
		log.Fatalf("[Server] Failed to start: %v", err)
	}
	defer ws.Stop()
	log.Printf("[Server] Listening on port %d", boundPort)

	// Advertise via Bonjour
	adv, err := discovery.Start(boundPort, serverName)
	if err != nil {
		log.Printf("[Bonjour] Failed to advertise: %v", err)
	} else {
		defer adv.Stop()
	}

	// Wait for shutdown signal
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	log.Println("[Server] Shutting down...")
}

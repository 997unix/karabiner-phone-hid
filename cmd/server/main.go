package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/tonyjiang/karabiner-phone-hid/internal/buildinfo"
	"github.com/tonyjiang/karabiner-phone-hid/internal/config"
	"github.com/tonyjiang/karabiner-phone-hid/internal/discovery"
	"github.com/tonyjiang/karabiner-phone-hid/internal/hid"
	"github.com/tonyjiang/karabiner-phone-hid/internal/server"
)

func main() {
	port := flag.Int("port", 8765, "WebSocket listen port")
	name := flag.String("name", "", "Server name for Bonjour (default: hostname)")
	webDir := flag.String("web", "", "Directory to serve web UI from (default: web/ next to binary)")
	flag.Parse()

	// Log build identity and re-exec provenance
	id := buildinfo.Self()
	if prev := buildinfo.PrevChecksum(); prev != "" {
		prevShort := prev
		if len(prevShort) > 12 {
			prevShort = prevShort[:12]
		}
		log.Printf("[Server] re-exec from checksum=%s to %s", prevShort, id)
	} else {
		log.Printf("[Server] start %s", id)
	}

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

	poster.InitPointing()
	log.Println("[Server] Waiting for Karabiner pointing device...")
	poster.WaitPointingReady()
	log.Println("[Server] Pointing device ready")

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

	// Serve web UI
	wd := *webDir
	if wd == "" {
		// Try relative to working directory first
		if _, err := os.Stat("web/index.html"); err == nil {
			wd = "web"
		} else if exe, err := os.Executable(); err == nil {
			// Then try relative to the binary
			candidate := filepath.Join(filepath.Dir(exe), "..", "web")
			if _, err := os.Stat(filepath.Join(candidate, "index.html")); err == nil {
				wd = candidate
			}
		}
	}
	if wd != "" {
		ws.SetWebDir(wd)
		log.Printf("[Server] Serving web UI from %s", wd)
		if html, err := os.ReadFile(filepath.Join(wd, "index.html")); err == nil {
			re := regexp.MustCompile(`<button data-idx="\d+"[^>]*>([^<]+)</button>`)
			var names []string
			for _, m := range re.FindAllSubmatch(html, -1) {
				names = append(names, string(m[1]))
			}
			if len(names) > 0 {
				log.Printf("[Server] Tabs: %s", strings.Join(names, " | "))
			}
		}
	}

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

	// Watch for binary changes — exec() new version on rebuild
	stopWatch := buildinfo.WatchSelfExec(30 * time.Second)
	defer stopWatch()

	// Wait for shutdown signal
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	log.Println("[Server] Shutting down...")
}

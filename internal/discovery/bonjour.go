package discovery

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

const ServiceType = "_phonehid._tcp"

// Advertiser publishes the server via macOS dns-sd command.
type Advertiser struct {
	cmd *exec.Cmd
}

// Start begins advertising the service using macOS dns-sd.
func Start(port int, serverName string) (*Advertiser, error) {
	if serverName == "" {
		name, _ := os.Hostname()
		serverName = name
	}

	cmd := exec.Command("dns-sd", "-R", serverName, ServiceType, "local", fmt.Sprintf("%d", port))
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("dns-sd start: %w", err)
	}

	log.Printf("[Bonjour] Published: %s on port %d", serverName, port)
	return &Advertiser{cmd: cmd}, nil
}

// Stop removes the service advertisement.
func (a *Advertiser) Stop() {
	if a.cmd != nil && a.cmd.Process != nil {
		a.cmd.Process.Kill()
		a.cmd.Wait()
		log.Printf("[Bonjour] Stopped")
	}
}

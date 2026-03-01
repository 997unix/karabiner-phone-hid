package discovery

import (
	"testing"
)

func TestServiceType(t *testing.T) {
	if ServiceType != "_phonehid._tcp" {
		t.Errorf("ServiceType = %q, want %q", ServiceType, "_phonehid._tcp")
	}
}

func TestStartAndStop(t *testing.T) {
	adv, err := Start(0, "TestServer")
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	adv.Stop()
}

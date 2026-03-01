package server

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/tonyjiang/karabiner-phone-hid/internal/hid"
	"github.com/tonyjiang/karabiner-phone-hid/internal/protocol"
)

func startTestServer(t *testing.T) (*Server, *hid.MockPoster, string) {
	t.Helper()
	mock := &hid.MockPoster{}
	dispatcher := hid.NewDispatcher(mock)
	resolver := &mockResolver{
		actions: map[string][]protocol.KeyStep{
			"test_action": {{Key: "a", Modifiers: []string{}}},
		},
	}
	router := NewRouter(dispatcher, resolver)
	actions := []protocol.ActionInfo{{Name: "test_action", Label: "Test"}}
	srv := NewServer(router, "Test Mac", actions)

	port, err := srv.Start("127.0.0.1:0")
	if err != nil {
		t.Fatalf("start server: %v", err)
	}
	t.Cleanup(func() { srv.Stop() })

	url := fmt.Sprintf("ws://127.0.0.1:%d/ws", port)
	return srv, mock, url
}

func connectWS(t *testing.T, url string) *websocket.Conn {
	t.Helper()
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	return conn
}

func readJSON(t *testing.T, conn *websocket.Conn) map[string]interface{} {
	t.Helper()
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, data, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	return result
}

func TestWSReceivesConfigOnConnect(t *testing.T) {
	_, _, url := startTestServer(t)
	conn := connectWS(t, url)

	msg := readJSON(t, conn)
	if msg["type"] != "config" {
		t.Errorf("type = %v, want config", msg["type"])
	}

	payload, ok := msg["payload"].(map[string]interface{})
	if !ok {
		t.Fatal("payload not an object")
	}
	if payload["server_name"] != "Test Mac" {
		t.Errorf("server_name = %v, want Test Mac", payload["server_name"])
	}

	actions, ok := payload["actions"].([]interface{})
	if !ok {
		t.Fatal("actions not an array")
	}
	if len(actions) != 1 {
		t.Fatalf("actions len = %d, want 1", len(actions))
	}
}

func TestWSSendKeypressAndReceiveAck(t *testing.T) {
	_, mock, url := startTestServer(t)
	conn := connectWS(t, url)

	// Read config first
	readJSON(t, conn)

	// Send keypress
	msg := `{"type":"action","id":"ws-1","action":"keypress","payload":{"key":"a","modifiers":[]}}`
	if err := conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
		t.Fatalf("write: %v", err)
	}

	ack := readJSON(t, conn)
	if ack["type"] != "ack" {
		t.Errorf("type = %v, want ack", ack["type"])
	}
	if ack["id"] != "ws-1" {
		t.Errorf("id = %v, want ws-1", ack["id"])
	}
	if ack["status"] != "ok" {
		t.Errorf("status = %v, want ok", ack["status"])
	}

	if len(mock.Calls) != 2 {
		t.Errorf("mock calls = %d, want 2", len(mock.Calls))
	}
}

func TestWSSendNamedActionAndReceiveAck(t *testing.T) {
	_, mock, url := startTestServer(t)
	conn := connectWS(t, url)

	// Read config
	readJSON(t, conn)

	msg := `{"type":"action","id":"ws-2","action":"test_action","payload":{}}`
	if err := conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
		t.Fatalf("write: %v", err)
	}

	ack := readJSON(t, conn)
	if ack["status"] != "ok" {
		t.Errorf("status = %v, want ok", ack["status"])
	}

	if len(mock.Calls) != 2 {
		t.Errorf("mock calls = %d, want 2", len(mock.Calls))
	}
}

func TestWSUnknownActionReturnsError(t *testing.T) {
	_, _, url := startTestServer(t)
	conn := connectWS(t, url)

	// Read config
	readJSON(t, conn)

	msg := `{"type":"action","id":"ws-3","action":"nonexistent","payload":{}}`
	if err := conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
		t.Fatalf("write: %v", err)
	}

	ack := readJSON(t, conn)
	if ack["status"] != "error" {
		t.Errorf("status = %v, want error", ack["status"])
	}
}

func TestWSInvalidJSONReturnsError(t *testing.T) {
	_, _, url := startTestServer(t)
	conn := connectWS(t, url)

	// Read config
	readJSON(t, conn)

	if err := conn.WriteMessage(websocket.TextMessage, []byte("not json")); err != nil {
		t.Fatalf("write: %v", err)
	}

	ack := readJSON(t, conn)
	if ack["status"] != "error" {
		t.Errorf("status = %v, want error", ack["status"])
	}
}

func TestWSMultipleClients(t *testing.T) {
	_, _, url := startTestServer(t)

	// Connect two clients
	conn1 := connectWS(t, url)
	conn2 := connectWS(t, url)

	// Both should receive config
	cfg1 := readJSON(t, conn1)
	cfg2 := readJSON(t, conn2)

	if cfg1["type"] != "config" {
		t.Errorf("client1 type = %v, want config", cfg1["type"])
	}
	if cfg2["type"] != "config" {
		t.Errorf("client2 type = %v, want config", cfg2["type"])
	}
}

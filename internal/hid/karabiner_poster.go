package hid

/*
#cgo CFLAGS: -I${SRCDIR}/../../cshim
#cgo LDFLAGS: -L${SRCDIR}/../../cshim -lkarabiner_shim -lc++
#include "karabiner_shim.h"
#include <stdlib.h>

// Forward declaration of Go callback.
extern void goKarabinerCallback(karabiner_status_t status, void* context);

// C wrapper that calls the Go callback.
static void cgo_status_callback(karabiner_status_t status, void* context) {
    goKarabinerCallback(status, context);
}

static karabiner_client_t* cgo_client_create(void* context) {
    return karabiner_client_create(cgo_status_callback, context);
}
*/
import "C"
import (
	"fmt"
	"log"
	"sync"
	"time"
	"unsafe"
)

// KarabinerPoster sends HID reports via the Karabiner DriverKit virtual HID device.
type KarabinerPoster struct {
	client        *C.karabiner_client_t
	mu            sync.Mutex
	ready         chan struct{}
	pointingReady chan struct{}
	state         linkState
	watchdogStop  chan struct{}
}

// posters tracks active KarabinerPoster instances for the C callback.
var (
	postersMu sync.Mutex
	posters   = map[uintptr]*KarabinerPoster{}
	posterID  uintptr
)

//export goKarabinerCallback
func goKarabinerCallback(status C.karabiner_status_t, context unsafe.Pointer) {
	id := uintptr(context)
	postersMu.Lock()
	poster, ok := posters[id]
	postersMu.Unlock()
	if !ok {
		return
	}

	// Every branch logs only on a state change. The client retries the socket
	// once a second while the daemon is unreachable, so logging each callback
	// unconditionally buries everything else in the log.
	switch status {
	case C.KARABINER_STATUS_CONNECTED:
		if poster.state.setConnected(true) {
			log.Println("[Karabiner] Connected to daemon")
		}
		// Re-create both devices on every connect, not just the first. The
		// client reconnects on its own after an outage, and the devices do not
		// survive it. Requests for a device that is already up are dropped by
		// the client, so this is free when nothing was lost.
		C.karabiner_client_init_keyboard(poster.client)
		C.karabiner_client_init_pointing(poster.client)

	case C.KARABINER_STATUS_KEYBOARD_READY:
		if poster.state.setKeyboardReady(true) {
			log.Println("[Karabiner] Keyboard ready")
		}
		select {
		case poster.ready <- struct{}{}:
		default:
		}

	case C.KARABINER_STATUS_KEYBOARD_NOT_READY:
		if poster.state.setKeyboardReady(false) {
			log.Println("[Karabiner] Keyboard went away")
		}

	case C.KARABINER_STATUS_POINTING_READY:
		if poster.state.setPointingReady(true) {
			log.Println("[Karabiner] Pointing ready")
		}
		select {
		case poster.pointingReady <- struct{}{}:
		default:
		}

	case C.KARABINER_STATUS_POINTING_NOT_READY:
		if poster.state.setPointingReady(false) {
			log.Println("[Karabiner] Pointing device went away, will re-initialize")
		}

	case C.KARABINER_STATUS_ERROR:
		if poster.state.setConnected(false) {
			log.Println("[Karabiner] Connection error, waiting for reconnect")
		}

	case C.KARABINER_STATUS_CONNECT_FAILED:
		// Fires once per retry, once a second, for as long as the daemon is
		// down. Only the first is worth a line.
		if poster.state.setConnected(false) {
			log.Println("[Karabiner] Cannot reach daemon, retrying")
		}

	case C.KARABINER_STATUS_CLOSED:
		if poster.state.setConnected(false) {
			log.Println("[Karabiner] Connection closed, waiting for reconnect")
		}
	}
}

// InitGlobal initializes the Karabiner dispatcher. Call once at startup.
func InitGlobal() {
	C.karabiner_global_init()
}

// CleanupGlobal cleans up the Karabiner dispatcher. Call once at shutdown.
func CleanupGlobal() {
	C.karabiner_global_cleanup()
}

// NewKarabinerPoster creates and starts a KarabinerPoster.
// Call WaitReady() to block until the keyboard is initialized.
func NewKarabinerPoster() *KarabinerPoster {
	poster := &KarabinerPoster{
		ready:         make(chan struct{}, 1),
		pointingReady: make(chan struct{}, 1),
	}

	postersMu.Lock()
	posterID++
	id := posterID
	posters[id] = poster
	postersMu.Unlock()

	poster.client = C.cgo_client_create(unsafe.Pointer(id))
	C.karabiner_client_start(poster.client)
	return poster
}

// WaitReady blocks until the keyboard is ready.
func (p *KarabinerPoster) WaitReady() {
	<-p.ready
}

// PostKeyboard sends a keyboard HID report.
func (p *KarabinerPoster) PostKeyboard(modifiers uint8, keys []uint16) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.client == nil {
		return fmt.Errorf("karabiner client not initialized")
	}

	var keysPtr *C.uint16_t
	if len(keys) > 0 {
		keysPtr = (*C.uint16_t)(unsafe.Pointer(&keys[0]))
	}

	C.karabiner_send_keyboard_report(p.client, C.uint8_t(modifiers), keysPtr, C.int(len(keys)))
	return nil
}

// ReleaseAll sends an empty keyboard report (release all keys).
func (p *KarabinerPoster) ReleaseAll() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.client == nil {
		return fmt.Errorf("karabiner client not initialized")
	}

	C.karabiner_send_keyboard_release(p.client)
	return nil
}

// InitPointing initializes the virtual pointing device.
func (p *KarabinerPoster) InitPointing() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.client != nil {
		C.karabiner_client_init_pointing(p.client)
	}
}

// WaitPointingReady blocks until the pointing device is ready.
func (p *KarabinerPoster) WaitPointingReady() {
	<-p.pointingReady
}

// PostPointing sends a pointing HID report.
func (p *KarabinerPoster) PostPointing(buttons uint32, x, y, verticalWheel, horizontalWheel int8) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.client == nil {
		return fmt.Errorf("karabiner client not initialized")
	}

	C.karabiner_send_pointing_report(p.client,
		C.uint32_t(buttons),
		C.int8_t(x), C.int8_t(y),
		C.int8_t(verticalWheel), C.int8_t(horizontalWheel))
	return nil
}

// ReleasePointing sends an empty pointing report (release all buttons).
func (p *KarabinerPoster) ReleasePointing() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.client == nil {
		return fmt.Errorf("karabiner client not initialized")
	}

	C.karabiner_send_pointing_release(p.client)
	return nil
}

// StartPointingWatchdog re-initializes the pointing device whenever it is not
// ready while the daemon is connected.
//
// This is a backstop, not the primary recovery path: the callbacks above
// already re-create both devices on reconnect. It covers the case where the
// connection survives but the pointing device does not, which is what made the
// mouse die overnight.
//
// It deliberately does not watch for a heartbeat. Karabiner 8.x reports
// readiness only when it changes, so a quiet channel means the state is
// unchanged, not that anything is wrong.
func (p *KarabinerPoster) StartPointingWatchdog() {
	p.watchdogStop = make(chan struct{})
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if !p.state.pointingNeedsInit() {
					continue
				}
				// Logged once per outage; the retry itself stays silent.
				if p.state.shouldReportReinit() {
					log.Println("[Karabiner] Pointing device not ready, re-initializing")
				}
				p.mu.Lock()
				if p.client != nil {
					C.karabiner_client_init_pointing(p.client)
				}
				p.mu.Unlock()
			case <-p.watchdogStop:
				return
			}
		}
	}()
}

// StopPointingWatchdog stops the pointing device watchdog.
func (p *KarabinerPoster) StopPointingWatchdog() {
	if p.watchdogStop != nil {
		close(p.watchdogStop)
	}
}

// Close destroys the client.
func (p *KarabinerPoster) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.client != nil {
		C.karabiner_client_destroy(p.client)
		p.client = nil
	}
}

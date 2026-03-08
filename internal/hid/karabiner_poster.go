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
	"sync"
	"unsafe"
)

// KarabinerPoster sends HID reports via the Karabiner DriverKit virtual HID device.
type KarabinerPoster struct {
	client        *C.karabiner_client_t
	mu            sync.Mutex
	ready         chan struct{}
	pointingReady chan struct{}
	readyDots         int // counts "keyboard ready" heartbeat dots
	pointingReadyDots int // counts "pointing ready" heartbeat dots
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

	switch status {
	case C.KARABINER_STATUS_CONNECTED:
		fmt.Println("[Karabiner] Connected to daemon")
		C.karabiner_client_init_keyboard(poster.client)
	case C.KARABINER_STATUS_KEYBOARD_READY:
		if poster.readyDots == 0 {
			fmt.Print("[Karabiner] Keyboard ready")
		}
		poster.readyDots++
		if poster.readyDots >= 80 {
			fmt.Println()
			fmt.Print("[Karabiner] Keyboard ready")
			poster.readyDots = 1
		} else {
			fmt.Print(".")
		}
		select {
		case poster.ready <- struct{}{}:
		default:
		}
	case C.KARABINER_STATUS_POINTING_READY:
		if poster.pointingReadyDots == 0 {
			fmt.Print("[Karabiner] Pointing ready")
		}
		poster.pointingReadyDots++
		if poster.pointingReadyDots >= 80 {
			fmt.Println()
			fmt.Print("[Karabiner] Pointing ready")
			poster.pointingReadyDots = 1
		} else {
			fmt.Print(".")
		}
		select {
		case poster.pointingReady <- struct{}{}:
		default:
		}
	case C.KARABINER_STATUS_ERROR:
		fmt.Println("[Karabiner] Error occurred")
	case C.KARABINER_STATUS_CONNECT_FAILED:
		fmt.Println("[Karabiner] Connection failed")
	case C.KARABINER_STATUS_CLOSED:
		fmt.Println("[Karabiner] Connection closed")
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

// Close destroys the client.
func (p *KarabinerPoster) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.client != nil {
		C.karabiner_client_destroy(p.client)
		p.client = nil
	}
}

#ifndef KARABINER_SHIM_H
#define KARABINER_SHIM_H

#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

// Opaque client handle.
typedef struct karabiner_client karabiner_client_t;

// Status codes returned by callbacks.
typedef enum {
    KARABINER_STATUS_OK = 0,
    KARABINER_STATUS_CONNECTED = 1,
    KARABINER_STATUS_KEYBOARD_READY = 2,
    KARABINER_STATUS_ERROR = -1,
    KARABINER_STATUS_CONNECT_FAILED = -2,
    KARABINER_STATUS_CLOSED = -3,
} karabiner_status_t;

// Callback for status changes.
typedef void (*karabiner_status_callback_t)(karabiner_status_t status, void* context);

// Initialize global dispatcher. Call once at startup.
void karabiner_global_init(void);

// Cleanup global dispatcher. Call once at shutdown.
void karabiner_global_cleanup(void);

// Create a client. callback is called for status changes.
karabiner_client_t* karabiner_client_create(karabiner_status_callback_t callback, void* context);

// Start the client connection to the daemon.
void karabiner_client_start(karabiner_client_t* client);

// Initialize the virtual keyboard device.
void karabiner_client_init_keyboard(karabiner_client_t* client);

// Send a keyboard HID report.
// modifiers: HID modifier byte (bit 0=LCtrl, 1=LShift, 2=LAlt, 3=LGUI, etc.)
// keys: array of USB HID usage codes for pressed keys
// key_count: number of keys (max 32)
void karabiner_send_keyboard_report(karabiner_client_t* client,
                                     uint8_t modifiers,
                                     const uint16_t* keys,
                                     int key_count);

// Send an empty keyboard report (release all keys).
void karabiner_send_keyboard_release(karabiner_client_t* client);

// Destroy the client and free resources.
void karabiner_client_destroy(karabiner_client_t* client);

#ifdef __cplusplus
}
#endif

#endif // KARABINER_SHIM_H

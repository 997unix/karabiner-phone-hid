// C++ wrapper around Karabiner-DriverKit-VirtualHIDDevice client library.
// This file wraps the header-only C++ client into plain C functions for CGo.
//
// Build: clang++ -std=c++20 -c karabiner_shim.cpp -o karabiner_shim.o \
//        -I../vendor-cpp/include
//
// The vendor-cpp directory should contain the Karabiner-DriverKit-VirtualHIDDevice
// headers (git submodule or manual copy).

#include "karabiner_shim.h"

// Guard against building without the Karabiner headers.
// When KARABINER_STUB is defined, we build a no-op stub for development/testing.
#ifndef KARABINER_STUB

#include <pqrs/dispatcher/extra/timer.hpp>
#include <pqrs/karabiner/driverkit/virtual_hid_device_service.hpp>
#include <memory>

struct karabiner_client {
    std::shared_ptr<pqrs::karabiner::driverkit::virtual_hid_device_service::client> client;
    karabiner_status_callback_t callback;
    void* context;
};

void karabiner_global_init(void) {
    pqrs::dispatcher::extra::initialize_shared_dispatcher();
}

void karabiner_global_cleanup(void) {
    pqrs::dispatcher::extra::terminate_shared_dispatcher();
}

karabiner_client_t* karabiner_client_create(karabiner_status_callback_t callback, void* context) {
    auto c = new karabiner_client();
    c->callback = callback;
    c->context = context;

    c->client = std::make_shared<pqrs::karabiner::driverkit::virtual_hid_device_service::client>();

    c->client->connected.connect([c] {
        if (c->callback) c->callback(KARABINER_STATUS_CONNECTED, c->context);
    });

    c->client->connect_failed.connect([c](auto&& error) {
        if (c->callback) c->callback(KARABINER_STATUS_CONNECT_FAILED, c->context);
    });

    c->client->closed.connect([c] {
        if (c->callback) c->callback(KARABINER_STATUS_CLOSED, c->context);
    });

    c->client->error_occurred.connect([c](auto&& error) {
        if (c->callback) c->callback(KARABINER_STATUS_ERROR, c->context);
    });

    c->client->virtual_hid_keyboard_ready.connect([c] {
        if (c->callback) c->callback(KARABINER_STATUS_KEYBOARD_READY, c->context);
    });

    return c;
}

void karabiner_client_start(karabiner_client_t* client) {
    if (client && client->client) {
        client->client->async_start();
    }
}

void karabiner_client_init_keyboard(karabiner_client_t* client) {
    if (client && client->client) {
        pqrs::karabiner::driverkit::virtual_hid_device_service::virtual_hid_keyboard_parameters params;
        client->client->async_virtual_hid_keyboard_initialize(params);
    }
}

void karabiner_send_keyboard_report(karabiner_client_t* client,
                                     uint8_t modifiers,
                                     const uint16_t* keys,
                                     int key_count) {
    if (!client || !client->client) return;

    pqrs::karabiner::driverkit::virtual_hid_device_driver::hid_report::keyboard_input report;

    // Set modifiers
    report.modifiers.insert(pqrs::karabiner::driverkit::virtual_hid_device_driver::hid_report::modifier(modifiers));

    // Set keys
    for (int i = 0; i < key_count && i < 32; i++) {
        report.keys.insert(pqrs::hid::usage::keyboard_or_keypad::value_t(keys[i]));
    }

    client->client->async_post_report(report);
}

void karabiner_send_keyboard_release(karabiner_client_t* client) {
    if (!client || !client->client) return;

    pqrs::karabiner::driverkit::virtual_hid_device_driver::hid_report::keyboard_input report;
    client->client->async_post_report(report);
}

void karabiner_client_destroy(karabiner_client_t* client) {
    if (client) {
        client->client = nullptr;
        delete client;
    }
}

#else // KARABINER_STUB

// Stub implementation for development without Karabiner headers.
#include <cstdio>

struct karabiner_client {
    karabiner_status_callback_t callback;
    void* context;
};

void karabiner_global_init(void) {
    fprintf(stderr, "[karabiner_shim] STUB: global_init\n");
}

void karabiner_global_cleanup(void) {
    fprintf(stderr, "[karabiner_shim] STUB: global_cleanup\n");
}

karabiner_client_t* karabiner_client_create(karabiner_status_callback_t callback, void* context) {
    fprintf(stderr, "[karabiner_shim] STUB: client_create\n");
    auto c = new karabiner_client();
    c->callback = callback;
    c->context = context;
    return c;
}

void karabiner_client_start(karabiner_client_t* client) {
    fprintf(stderr, "[karabiner_shim] STUB: client_start\n");
    // Simulate connected + keyboard ready
    if (client && client->callback) {
        client->callback(KARABINER_STATUS_CONNECTED, client->context);
        client->callback(KARABINER_STATUS_KEYBOARD_READY, client->context);
    }
}

void karabiner_client_init_keyboard(karabiner_client_t* client) {
    fprintf(stderr, "[karabiner_shim] STUB: init_keyboard\n");
}

void karabiner_send_keyboard_report(karabiner_client_t* client,
                                     uint8_t modifiers,
                                     const uint16_t* keys,
                                     int key_count) {
    fprintf(stderr, "[karabiner_shim] STUB: keyboard_report mods=0x%02x keys=%d\n", modifiers, key_count);
}

void karabiner_send_keyboard_release(karabiner_client_t* client) {
    fprintf(stderr, "[karabiner_shim] STUB: keyboard_release\n");
}

void karabiner_client_destroy(karabiner_client_t* client) {
    fprintf(stderr, "[karabiner_shim] STUB: client_destroy\n");
    delete client;
}

#endif // KARABINER_STUB

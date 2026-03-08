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

#include <filesystem>
#include <memory>
#include <pqrs/dispatcher/extra/timer.hpp>
#include <pqrs/karabiner/driverkit/virtual_hid_device_driver.hpp>
#include <pqrs/karabiner/driverkit/virtual_hid_device_service.hpp>

namespace vhd_driver = pqrs::karabiner::driverkit::virtual_hid_device_driver;
namespace vhd_service = pqrs::karabiner::driverkit::virtual_hid_device_service;

struct karabiner_client {
    std::shared_ptr<vhd_service::client> client;
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

    c->client = std::make_shared<vhd_service::client>();

    c->client->connected.connect([c] {
        if (c->callback) c->callback(KARABINER_STATUS_CONNECTED, c->context);
    });

    c->client->connect_failed.connect([c](const asio::error_code&) {
        if (c->callback) c->callback(KARABINER_STATUS_CONNECT_FAILED, c->context);
    });

    c->client->closed.connect([c] {
        if (c->callback) c->callback(KARABINER_STATUS_CLOSED, c->context);
    });

    c->client->error_occurred.connect([c](const asio::error_code&) {
        if (c->callback) c->callback(KARABINER_STATUS_ERROR, c->context);
    });

    c->client->virtual_hid_keyboard_ready.connect([c](bool ready) {
        if (ready && c->callback) {
            c->callback(KARABINER_STATUS_KEYBOARD_READY, c->context);
        }
    });

    c->client->virtual_hid_pointing_ready.connect([c](bool ready) {
        if (ready && c->callback) {
            c->callback(KARABINER_STATUS_POINTING_READY, c->context);
        }
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
        vhd_service::virtual_hid_keyboard_parameters params;
        client->client->async_virtual_hid_keyboard_initialize(params);
    }
}

void karabiner_send_keyboard_report(karabiner_client_t* client,
                                     uint8_t modifiers,
                                     const uint16_t* keys,
                                     int key_count) {
    if (!client || !client->client) return;

    vhd_driver::hid_report::keyboard_input report;

    // Set modifier bits individually
    for (int bit = 0; bit < 8; bit++) {
        if (modifiers & (1 << bit)) {
            report.modifiers.insert(
                static_cast<vhd_driver::hid_report::modifier>(1 << bit));
        }
    }

    // Set keys (keys.insert takes plain uint16_t)
    for (int i = 0; i < key_count && i < 32; i++) {
        report.keys.insert(keys[i]);
    }

    client->client->async_post_report(report);
}

void karabiner_send_keyboard_release(karabiner_client_t* client) {
    if (!client || !client->client) return;

    vhd_driver::hid_report::keyboard_input report;
    client->client->async_post_report(report);
}

void karabiner_client_init_pointing(karabiner_client_t* client) {
    if (client && client->client) {
        client->client->async_virtual_hid_pointing_initialize();
    }
}

void karabiner_send_pointing_report(karabiner_client_t* client,
                                     uint32_t buttons,
                                     int8_t x, int8_t y,
                                     int8_t vertical_wheel,
                                     int8_t horizontal_wheel) {
    if (!client || !client->client) return;

    vhd_driver::hid_report::pointing_input report;
    report.buttons.insert(buttons);
    report.x = static_cast<uint8_t>(x);
    report.y = static_cast<uint8_t>(y);
    report.vertical_wheel = static_cast<uint8_t>(vertical_wheel);
    report.horizontal_wheel = static_cast<uint8_t>(horizontal_wheel);

    client->client->async_post_report(report);
}

void karabiner_send_pointing_release(karabiner_client_t* client) {
    if (!client || !client->client) return;

    vhd_driver::hid_report::pointing_input report;
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
    // Simulate connected + keyboard ready + pointing ready
    if (client && client->callback) {
        client->callback(KARABINER_STATUS_CONNECTED, client->context);
        client->callback(KARABINER_STATUS_KEYBOARD_READY, client->context);
        client->callback(KARABINER_STATUS_POINTING_READY, client->context);
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

void karabiner_client_init_pointing(karabiner_client_t* client) {
    fprintf(stderr, "[karabiner_shim] STUB: init_pointing\n");
}

void karabiner_send_pointing_report(karabiner_client_t* client,
                                     uint32_t buttons,
                                     int8_t x, int8_t y,
                                     int8_t vertical_wheel,
                                     int8_t horizontal_wheel) {
    fprintf(stderr, "[karabiner_shim] STUB: pointing_report buttons=0x%x x=%d y=%d vw=%d hw=%d\n",
            buttons, x, y, vertical_wheel, horizontal_wheel);
}

void karabiner_send_pointing_release(karabiner_client_t* client) {
    fprintf(stderr, "[karabiner_shim] STUB: pointing_release\n");
}

void karabiner_client_destroy(karabiner_client_t* client) {
    fprintf(stderr, "[karabiner_shim] STUB: client_destroy\n");
    delete client;
}

#endif // KARABINER_STUB

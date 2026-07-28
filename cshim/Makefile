CXX := clang++
CXXFLAGS := -std=c++23 -O2 -Wall -mmacosx-version-min=13.0

# Include paths for Karabiner headers.
KARABINER_INCLUDE ?= ../vendor-headers/karabiner
KARABINER_VENDOR_INCLUDE ?= ../vendor-headers/deps

# Build with stub (no Karabiner headers needed) by default.
# Set STUB=0 to build with real Karabiner headers.
STUB ?= 1

ifeq ($(STUB),1)
CXXFLAGS += -DKARABINER_STUB
else
CXXFLAGS += -I$(KARABINER_INCLUDE) -I$(KARABINER_VENDOR_INCLUDE)
endif

.PHONY: all clean

all: libkarabiner_shim.a

# Nothing in the dependency graph encodes the STUB flag, so `make test` (STUB=1)
# followed by `make` (STUB=0) would otherwise reuse the stub object and link a
# server that accepts keystrokes and types nothing. Record the flag in a stamp
# file and drop stale artifacts when it changes.
#
# This runs at parse time on purpose. As a normal rule it would not work: make
# stats targets before running any recipe, so artifacts deleted by a recipe can
# still be considered up to date. GNU make 3.81 (the /usr/bin/make Apple ships)
# also compares mtimes at whole-second granularity, so a stamp rewritten in the
# same second as the object it guards would not look newer either.
STUB_STAMP := $(shell \
	  cur='STUB=$(STUB)'; \
	  if [ "$$(cat .build-flags 2>/dev/null)" != "$$cur" ]; then \
	    rm -f karabiner_shim.o libkarabiner_shim.a; \
	    echo "$$cur" > .build-flags; \
	  fi)

karabiner_shim.o: karabiner_shim.cpp karabiner_shim.h
	$(CXX) $(CXXFLAGS) -c karabiner_shim.cpp -o karabiner_shim.o

libkarabiner_shim.a: karabiner_shim.o
	ar rcs libkarabiner_shim.a karabiner_shim.o

clean:
	rm -f karabiner_shim.o libkarabiner_shim.a .build-flags

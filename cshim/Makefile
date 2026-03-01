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

karabiner_shim.o: karabiner_shim.cpp karabiner_shim.h
	$(CXX) $(CXXFLAGS) -c karabiner_shim.cpp -o karabiner_shim.o

libkarabiner_shim.a: karabiner_shim.o
	ar rcs libkarabiner_shim.a karabiner_shim.o

clean:
	rm -f karabiner_shim.o libkarabiner_shim.a

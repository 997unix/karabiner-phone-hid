CXX := clang++
CXXFLAGS := -std=c++20 -O2 -Wall

# Set KARABINER_INCLUDE to the path containing pqrs/ headers.
# Default: ../vendor-cpp/include
KARABINER_INCLUDE ?= ../vendor-cpp/include

# Build with stub (no Karabiner headers needed) by default.
# Set STUB=0 to build with real Karabiner headers.
STUB ?= 1

ifeq ($(STUB),1)
CXXFLAGS += -DKARABINER_STUB
else
CXXFLAGS += -I$(KARABINER_INCLUDE)
endif

.PHONY: all clean

all: libkarabiner_shim.a

karabiner_shim.o: karabiner_shim.cpp karabiner_shim.h
	$(CXX) $(CXXFLAGS) -c karabiner_shim.cpp -o karabiner_shim.o

libkarabiner_shim.a: karabiner_shim.o
	ar rcs libkarabiner_shim.a karabiner_shim.o

clean:
	rm -f karabiner_shim.o libkarabiner_shim.a

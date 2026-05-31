APP := ikemen-minimalist
SRC := ./src
BIN_DIR := bin

ifeq ($(OS),Windows_NT)
	BIN := $(BIN_DIR)/$(APP).exe
	# GUI subsystem is the default on Windows. This avoids the console subsystem
	# stealing/fighting focus in some w64devkit/terminal setups and lets SDL own
	# the app window/input path. Add `-log <file>` to capture stdout, or use
	# `make build-console` if you need visible console output.
	DEFAULT_LDFLAGS := -ldflags "-H=windowsgui -s -w"
	CONSOLE_LDFLAGS := -ldflags "-s -w"
else
	BIN := $(BIN_DIR)/$(APP)
	DEFAULT_LDFLAGS :=
	CONSOLE_LDFLAGS :=
endif

.PHONY: all deps build build-gui build-console run clean

all: build

deps:
	go mod tidy

build: deps
	mkdir -p $(BIN_DIR)
	CGO_ENABLED=1 go build $(DEFAULT_LDFLAGS) -o $(BIN) $(SRC)

# Explicit GUI build. On non-Windows this is equivalent to normal build.
build-gui: deps
	mkdir -p $(BIN_DIR)
	CGO_ENABLED=1 go build -trimpath -v -tags "static" $(DEFAULT_LDFLAGS) -o $(BIN) $(SRC)

# Debug build with console subsystem on Windows, useful for fmt.Println logs.
build-console: deps
	mkdir -p $(BIN_DIR)
	CGO_ENABLED=1 go build -trimpath -v -tags "static" $(CONSOLE_LDFLAGS) -o $(BIN) $(SRC)

run: build
	@if [ -z "$(SFF)" ] || [ -z "$(AIR)" ]; then \
		echo "Usage: make run SFF=/path/to/char.sff AIR=/path/to/char.air"; \
		exit 2; \
	fi
	$(BIN) "$(SFF)" "$(AIR)"

clean:
	rm -rf $(BIN_DIR)

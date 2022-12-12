# Basic commands
CLEAN	:= rm -f $(BINARY_OUTPUT)
ifeq ($(OS), Windows_NT)
	CLEAN := del /q/s $(BINARY_OUTPUT) 2>&1 | exit 0
endif

# App info
BINARY_NAME := spotify-server

# Folders
OUTPUT_FOLDER := bin
BINARY_OUTPUT := $(OUTPUT_FOLDER)/$(BINARY_NAME)
ifeq ($(OS), Windows_NT)
	OUTPUT_FOLDER := .\bin
	BINARY_OUTPUT := $(OUTPUT_FOLDER)\$(BINARY_NAME).exe
endif

.PHONY: default
default: clean fmt build run

# App basic commands
clean:
	@go clean -i . && $(CLEAN)

fmt:
	@gofmt -s -w -l .

build:
	@go build -o $(BINARY_OUTPUT) .

run:
	@$(BINARY_OUTPUT)
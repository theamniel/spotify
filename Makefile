# Basic commands
CLEAN	:= rm -f $(BINARY_OUTPUT)
ifeq ($(OS), Windows_NT)
	CLEAN := del /q/s $(BINARY_OUTPUT) 2>&1 | exit 0
endif

# App info
BINARY_NAME 			:= spotify.server
ENTRY_MAIN 				:= ./bin/spotify
BINARY_NAME_GRPC 	:= spotify.grpc
ENTRY_GRPC 				:= ./bin/spotify.grpc
PROTO_FILES 			:= services/grpc/proto

# Folders
OUTPUT_FOLDER 			:= .build
BINARY_OUTPUT 			:= $(OUTPUT_FOLDER)/$(BINARY_NAME)
BINARY_OUTPUT_GRPC 	:= $(OUTPUT_FOLDER)/${BINARY_NAME_GRPC}
ifeq ($(OS), Windows_NT)
	OUTPUT_FOLDER := .\.build
	BINARY_OUTPUT := $(OUTPUT_FOLDER)\$(BINARY_NAME).exe
	BINARY_OUTPUT_GRPC 	:= $(OUTPUT_FOLDER)\${BINARY_NAME_GRPC}.exe
endif

.PHONY: setup ## Install all the build dependencies
setup:
	@echo Updating dependency tree...
	go mod tidy 
	go mod download
	@echo Updated dependency tree successfully.

.PHONY: default
default: clean fmt build run

generate-proto:
	protoc --go_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_out=. \
		--go-grpc_opt=paths=source_relative \
		$(PROTO_FILES)/spotify.proto

build-grpc:
	@echo Building grpc binary...
	@go build -o $(BINARY_OUTPUT_GRPC) $(ENTRY_GRPC)
	@echo Built grpc binary successfully.

build-server:
	@echo Building server binary...
	@go build -o $(BINARY_OUTPUT) $(ENTRY_MAIN)
	@echo Built server binary successfully.


clean:
	@go clean -i . && $(CLEAN)

fmt:
	@gofmt -s -w -l .

build: build-grpc build-server

test:
	@cd .example && npm run dev

run-grpc:
	@$(BINARY_OUTPUT_GRPC)

run-server:
	@$(BINARY_OUTPUT)

run:
	$(BINARY_OUTPUT)
	$(BINARY_NAME_GRPC-grpc)
	
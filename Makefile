# App basic
APP_NAME	:= spotify-server
BUILD_DIR	:= bin
ENTRY			:= .
OUTPUT		:= ./$(BUILD_DIR)/$(APP_NAME)

# App commands
CLEAN	:= rm -f $(OUTPUT)

ifeq ($(OS), Windows_NT) 
	OUTPUT 	:= .\$(BUILD_DIR)\$(APP_NAME).exe
	CLEAN		:= del /q/s $(OUTPUT) 2>&1 | exit 0
endif

.PHONY: default
default: clean fmt build run

# App basic commands
clean:
	@$(CLEAN)

fmt:
	@gofmt -s -w -l .

build:
	@go build -o $(OUTPUT) $(ENTRY)

install:
	@go install $(ENTRY)

run:
	@$(OUTPUT)
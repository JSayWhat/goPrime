# Go parameters
GOCMD = go
GORUN = $(GOCMD) run
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean
GOTEST = $(GOCMD) test
GOGET = $(GOCMD) get

# OS-specific binary extension
OS = $(shell uname -s)
ifeq ($(OS),Windows_NT)
    EXT = .exe
else
    EXT =
endif

# Build target
BINARY_NAME = goPrime$(EXT)
BUILD_PATH = cmd/$(BINARY_NAME)

all: test build run

run:
	@echo "Running the application..." 
	./$(BUILD_PATH)

build:
	@echo "Building the application..."
	$(GOBUILD) -o $(BUILD_PATH) -v 

test:
	@echo "Testing..."
	$(GOTEST) -v ./...

clean:
	@echo "Cleaning up..."
	$(GOCLEAN)
	rm -f $(BUILD_PATH)

.PHONY: all build test clean run
# Go parameters
GOCMD = go
GORUN = $(GOCMD) run
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean
GOTEST = $(GOCMD) test
GOGET = $(GOCMD) get

# Build target
BINARY_NAME = goPrime

all: test build run

run:
	@echo "Running the application..." 
	./$(BINARY_NAME)

build:
	@echo "Building the application..."
	$(GOBUILD) -o $(BINARY_NAME) -v

test:
	@echo "Testing..."
	$(GOTEST) -v ./...

clean:
	@echo "Cleaning up..."
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

.PHONY: all build test clean run
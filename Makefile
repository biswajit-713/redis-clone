# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=dicedb
BINARY_UNIX=$(BINARY_NAME)_unix
BINARY_ARM=$(BINARY_NAME)_arm

# Build the project
build:
	$(GOBUILD) -o $(BINARY_NAME) -v ./...

# Run the project
run: build
	./$(BINARY_NAME)

# Clean the project
clean:
	$(GOCLEAN)
	rm -rf $(BINARY_NAME)
	rm -rf $(BINARY_UNIX)

# Test the project
test:
	$(GOTEST) -v ./...

# Install dependencies
deps:
	$(GOGET) -u ./...

# Cross compilation for ARM
build-arm:
	GOOS=linux GOARCH=arm $(GOBUILD) -o $(BINARY_ARM) -v ./...

.PHONY: build clean run test deps build-arm

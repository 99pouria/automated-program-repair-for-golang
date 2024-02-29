# Define the name of the binary
BINARY_NAME=apr

# Default target: build the executable
all: build

# Rule to build the target executable
build:
	go build -o $(BINARY_NAME) cmd/go_apr/main.go

# Clean target: remove the target executable
clean:
	go clean
	rm -f $(BINARY_NAME)

# Run target: build and run the target executable
run: build
	./$(BINARY_NAME) -p $(path) -f $(func) -t $(tests)
	$(MAKE) clean

# Test target: run Go tests for the project
test:
	go test ./...

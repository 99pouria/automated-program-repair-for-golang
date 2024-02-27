# Define the name of the binary
BINARY_NAME=apr

# PATH=""
# FUNC=""

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
	echo $(PATH)
	echo $(FUNC)
	./$(BINARY_NAME) -p $(PATH) -f $(FUNC)

# Test target: run Go tests for the project
test:
	go test ./...

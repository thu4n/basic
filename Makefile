# Name of the binary
BINARY_NAME=basic

# Build the binary in the local directory
build:
	go build -o $(BINARY_NAME) .

# Build and install the binary to $GOPATH/bin (usually ~/go/bin)
install:
	go install .

# Clean up local binary
clean:
	rm -f $(BINARY_NAME)

# Update version (alias for install)
update: install
	@echo "Successfully updated $(BINARY_NAME) in your PATH!"

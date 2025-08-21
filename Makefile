.PHONY: build test clean install uninstall

# Build the binary
build:
	go build -o ccsl ./cmd/ccsl

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -f ccsl
	go clean

# Install locally
install: build
	./scripts/install.sh

# Uninstall
uninstall:
	./scripts/uninstall.sh

# Test with sample data
demo: build
	echo '{"model":{"display_name":"Demo Model"},"workspace":{"current_dir":"'$(shell pwd)'"}}' | ./ccsl

# Show help
help:
	@echo "Available targets:"
	@echo "  build      - Build the ccsl binary"
	@echo "  test       - Run all tests"  
	@echo "  clean      - Clean build artifacts"
	@echo "  install    - Install ccsl locally"
	@echo "  uninstall  - Uninstall ccsl"
	@echo "  demo       - Run demo with sample data"
	@echo "  help       - Show this help message"
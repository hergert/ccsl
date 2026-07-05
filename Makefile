.PHONY: build test lint demo install uninstall clean help

build:
	go build -o ccsl ./cmd/ccsl

test:
	go test ./...

lint:
	gofmt -l . && go vet ./...

demo: build
	./ccsl doctor

install:
	go build -o $(HOME)/.local/bin/ccsl ./cmd/ccsl

uninstall:
	./scripts/uninstall.sh

clean:
	rm -f ccsl
	go clean

help:
	@echo "  build     - Build the ccsl binary here"
	@echo "  test      - Run all tests"
	@echo "  lint      - gofmt + go vet"
	@echo "  demo      - Build and run ccsl doctor with sample data"
	@echo "  install   - Build and install to ~/.local/bin (dev flow)"
	@echo "  uninstall - Remove binary and settings entry"
	@echo "  clean     - Clean build artifacts"

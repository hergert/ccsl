.PHONY: build test clean install uninstall demo help

build:
	go build -o ccsl ./cmd/ccsl

test:
	go test -v ./...

clean:
	rm -f ccsl
	go clean

install: build
	./scripts/install.sh

uninstall:
	./scripts/uninstall.sh

demo: build
	echo '{"model":{"display_name":"Demo"},"agent":{"name":"task"},"workspace":{"current_dir":"'$(shell pwd)'"},"context_window":{"used_percentage":92},"exceeds_200k_tokens":true,"cost":{"total_cost_usd":0.05,"total_duration_ms":498000,"total_lines_added":156,"total_lines_removed":23}}' | ./ccsl

help:
	@echo "Available targets:"
	@echo "  build     - Build the ccsl binary"
	@echo "  test      - Run all tests"
	@echo "  clean     - Clean build artifacts"
	@echo "  install   - Install ccsl locally"
	@echo "  uninstall - Uninstall ccsl"
	@echo "  demo      - Run demo with sample data"
	@echo "  help      - Show this help"

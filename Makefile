.PHONY: build clean test vet fmt lint help

help:
	@echo "Airborne (ab) - MVP developer CLI"
	@echo ""
	@echo "Targets:"
	@echo "  build    - Build the ab binary"
	@echo "  test     - Run tests"
	@echo "  vet      - Run go vet"
	@echo "  fmt      - Format source code"
	@echo "  clean    - Remove build artifacts"

build:
	go build -o ab ./cmd/ab

test:
	go test ./...

vet:
	go vet ./...

fmt:
	go fmt ./...

clean:
	rm -f ab

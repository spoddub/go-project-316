APP_NAME := hexlet-go-crawler
CMD_PATH := ./cmd/hexlet-go-crawler
BIN_DIR := bin
BIN_PATH := $(BIN_DIR)/$(APP_NAME)

tidy:
	go mod tidy
	go fmt ./...
	go vet ./...

test:
	go test ./...

lint:
	golangci-lint run

lint-fix:
	golangci-lint run --fix

build:
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN_PATH) $(CMD_PATH)

run:
	@if [ -z "$(URL)" ]; then \
		echo "Error: URL is required."; \
		echo "Usage: make run URL=https://example.com"; \
		exit 1; \
	fi
	@go run $(CMD_PATH) "$(URL)"
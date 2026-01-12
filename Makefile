build:
	mkdir -p bin
	go build -o bin/hexlet-go-crawler ./cmd/hexlet-go-crawler

test:
	go test ./...
run:
	@if [ -z "$(URL)" ]; then \
		go run ./cmd/hexlet-go-crawler --help; \
	else \
		go run ./cmd/hexlet-go-crawler "$(URL)"; \
	fi
tidy:
	go mod tidy
	go fmt ./...
	go vet ./...
lint:
	golangci-lint run ./...
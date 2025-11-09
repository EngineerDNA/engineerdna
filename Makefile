.PHONY: build clean test run install frontend backend plugins build-plugins

# Version from git tags or CHANGELOG.md
VERSION ?= $(shell ./scripts/get-version.sh)

# Build everything (frontend + backend)
build: frontend backend
	@echo "Build complete"

# Build frontend
frontend:
	@echo "Building frontend..."
	cd frontend && npm install && npm run build

# Build backend
backend:
	@echo "Building backend..."
	@echo "Version: $(VERSION)"
	go build -ldflags="-s -w -X main.Version=$(VERSION)" -o bin/engineerdna .

# Build all plugins
plugins:
	@echo "Building all plugins..."
	@echo "Building source plugins..."
	cd plugins/github && go build -ldflags="-s -w" -o github .
	cd plugins/aws-costs && go build -ldflags="-s -w" -o aws-costs .
	cd plugins/csv-import && go build -ldflags="-s -w" -o csv-import .
	@echo "Building processor plugins..."
	cd plugins/claude-insights && go build -ldflags="-s -w" -o claude-insights .
	cd plugins/openai-insights && go build -ldflags="-s -w" -o openai-insights .
	cd plugins/ollama-insights && go build -ldflags="-s -w" -o ollama-insights .
	@echo "Building destination plugins..."
	cd plugins/google-sheets-export && go build -ldflags="-s -w" -o google-sheets-export .
	@echo "All plugins built successfully"

# Legacy alias
build-plugins: plugins

# Clean build artifacts
clean:
	rm -rf bin/
	rm -rf frontend/dist/
	rm -rf frontend/node_modules/
	rm -f plugins/github/github
	rm -f plugins/aws-costs/aws-costs
	rm -f plugins/csv-import/csv-import
	rm -f plugins/claude-insights/claude-insights
	rm -f plugins/openai-insights/openai-insights
	rm -f plugins/ollama-insights/ollama-insights
	rm -f plugins/google-sheets-export/google-sheets-export
	find plugins -type f -name "*.exe" -delete

# Run tests
test:
	go test ./...

# Run the server
run: build
	./bin/engineerdna serve

# Install to system
install: build
	cp bin/engineerdna /usr/local/bin/

# Initialize the application
init: build
	./bin/engineerdna init

# Cross-compile for different platforms
cross-compile:
	@echo "Cross-compiling for all platforms..."
	@echo "Version: $(VERSION)"
	@mkdir -p bin
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X main.Version=$(VERSION)" -o bin/engineerdna-linux-amd64 .
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w -X main.Version=$(VERSION)" -o bin/engineerdna-linux-arm64 .
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w -X main.Version=$(VERSION)" -o bin/engineerdna-darwin-amd64 .
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w -X main.Version=$(VERSION)" -o bin/engineerdna-darwin-arm64 .
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w -X main.Version=$(VERSION)" -o bin/engineerdna-windows-amd64.exe .
	CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build -ldflags="-s -w -X main.Version=$(VERSION)" -o bin/engineerdna-windows-arm64.exe .
	@echo "Cross-compilation complete. Binaries in bin/"

# Development helpers
dev-frontend:
	cd frontend && npm run dev

dev-backend: backend
	./bin/engineerdna serve

dev: build
	./bin/engineerdna init
	./bin/engineerdna serve

fmt:
	go fmt ./...

vet:
	go vet ./...

lint: fmt vet

# Help
help:
	@echo "EngineerDNA Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  build                - Build frontend and backend"
	@echo "  frontend             - Build frontend only"
	@echo "  backend              - Build backend only"
	@echo "  plugins              - Build all plugins"
	@echo "  clean                - Remove build artifacts"
	@echo "  test                 - Run tests"
	@echo "  run                  - Build and run the server"
	@echo "  install              - Install to /usr/local/bin"
	@echo "  init                 - Initialize the application"
	@echo "  cross-compile        - Build for all platforms"
	@echo "  dev                  - Initialize and start server"
	@echo "  dev-frontend         - Run frontend dev server"
	@echo "  dev-backend          - Run backend dev server"
	@echo "  fmt                  - Format code"
	@echo "  vet                  - Run go vet"
	@echo "  lint                 - Format and vet code"

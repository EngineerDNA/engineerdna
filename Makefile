.PHONY: build clean test run install frontend backend

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
build-plugins:
	@echo "Building plugins..."
	@mkdir -p plugins/csv-import
	@mkdir -p plugins/github
	@mkdir -p plugins/ai-insights
	@mkdir -p plugins/google-sheets-export
	@mkdir -p plugins/pdf-export
	@mkdir -p plugins/markdown-export

# Build PDF export plugin
build-pdf-plugin:
	@echo "Building PDF export plugin..."
	cd plugins/pdf-export && go build -o pdf-export main.go
	@echo "PDF export plugin built successfully"

# Build Markdown export plugin
build-markdown-plugin:
	@echo "Building Markdown export plugin..."
	cd plugins/markdown-export && go build -o markdown-export main.go
	@echo "Markdown export plugin built successfully"

# Clean build artifacts
clean:
	rm -rf bin/
	rm -rf frontend/dist/
	rm -rf frontend/node_modules/
	find plugins -type f -name "main" -delete
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
	@echo "  build-plugins        - Build all plugins"
	@echo "  build-pdf-plugin     - Build PDF export plugin"
	@echo "  build-markdown-plugin - Build Markdown export plugin"
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

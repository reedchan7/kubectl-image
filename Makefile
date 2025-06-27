.PHONY: build clean test install
.DEFAULT_GOAL := help

# Variables
BINARY_NAME=kubectl-image
CMD_PATH=./src/cmd
BIN_DIR=./bin

# Version information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dirty")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build flags
LDFLAGS = -w -s \
	-X 'main.version=$(VERSION)' \
	-X 'main.commit=$(COMMIT)' \
	-X 'main.date=$(DATE)'

help: ## Print help message
	@printf "\nUsage: make <command>\n"
	@grep -F -h "##" $(MAKEFILE_LIST) | grep -F -v grep -F | sed -e 's/\\$$//' | awk 'BEGIN {FS = ":*[[:space:]]*##"}; \
	{ \
		if($$2 == "") \
			pass; \
		else if($$0 ~ /^#/) \
			printf "\n%s\n", $$2; \
		else if($$1 == "") \
			printf "     %-28s%s\n", "", $$2; \
		else \
			printf "    \033[34m%-28s\033[0m %s\n", $$1, $$2; \
	}'

build: ## Build the Go application
	@echo "🔨 Building $(BINARY_NAME) plugin..."
	@echo "Version: $(VERSION)"
	@echo "Commit: $(COMMIT)"
	@echo "Date: $(DATE)"
	@CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(BINARY_NAME) $(CMD_PATH)
	@echo "✅ Build completed: $(BIN_DIR)/$(BINARY_NAME)"

clean: ## Clean the project
	@echo "🧹 Cleaning up..."
	@rm -f $(BIN_DIR)/$(BINARY_NAME)
	@echo "✅ Clean completed"

test: ## Run tests
	@echo "🧪 Running tests..."
	@go test -v ./...
	@echo "✅ Tests completed"

install: build ## Install the plugin to /usr/local/bin
	@echo "🚀 Installing $(BINARY_NAME) to /usr/local/bin..."
	@sudo cp $(BIN_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "✅ Installation successful."
	@echo "👉 To verify, run: 'kubectl plugin list | grep image'"

uninstall: ## Uninstall the plugin from /usr/local/bin
	@echo "🗑️ Uninstalling $(BINARY_NAME) from /usr/local/bin..."
	@sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "✅ Uninstallation successful."

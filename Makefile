VERSION := $(shell git describe --tags --abbrev=0 2>/dev/null)
LDFLAGS := -X github.com/fsuhrau/automationhub/hub.version=$(VERSION)
SRC_DIR=./proto
DST_DIR=./out

.PHONY: help
help:
	@echo "Build:"
	@echo "  make frontend         Build the web UI into endpoints/web/data (embedded by the master)"
	@echo "  make buildMac         Frontend + master/node and cli binaries for macOS (bin/macos)"
	@echo "  make buildLinux       Frontend + binaries for linux/amd64 (bin/linux)"
	@echo "  make buildWindows     Frontend + binaries for windows/amd64 (bin/windows)"
	@echo "  make proto            Generate the node protobuf code"
	@echo "Run:"
	@echo "  make server           Run the master (automationhub.yaml)"
	@echo "  make node             Run a node (automationhub_slave.yaml)"
	@echo "  make run              Run the web UI dev server (proxies /api to :8002)"
	@echo "  make lint             Lint and fix the frontend"
	@echo "DevContainer:"
	@echo "  make dev-bootstrap    Fetch/install devcontainer base layer (first time setup)"
	@echo "  make dev-help         Show all devcontainer commands"

.PHONY: proto
proto:
	mkdir -p $(DST_DIR)/csharp $(DST_DIR)/cpp
	protoc --proto_path=$(SRC_DIR)/node -I=$(SRC_DIR)/node --go_out=. node.proto
	#protoc --proto_path=$(SRC_DIR)/dut -I=$(SRC_DIR)/dut --go_out=. --cpp_out=$(DST_DIR)/cpp --csharp_out=$(DST_DIR)/csharp action.proto

# run the frontend locally
.PHONY: run
run:
	cd frontend && yarn start

# run linter (ESLint)
.PHONY: lint
lint:
	cd frontend && yarn lint --fix

# run frontend build
.PHONY: frontend
frontend:
	cd frontend && BUILD_PATH=../endpoints/web/data/ yarn build

# build frontend and backend; GOOS/GOARCH go on the go build commands, so
# the Linux and Windows targets really cross-compile
.PHONY: buildMac
buildMac: frontend
	go build -trimpath -ldflags '$(LDFLAGS)' -o bin/macos/automationhub .
	go build -trimpath -ldflags '$(LDFLAGS)' -o bin/macos/cli ./cli

.PHONY: buildWindows
buildWindows: frontend
	GOOS=windows GOARCH=amd64 go build -trimpath -ldflags '$(LDFLAGS)' -o bin/windows/automationhub.exe .
	GOOS=windows GOARCH=amd64 go build -trimpath -ldflags '$(LDFLAGS)' -o bin/windows/cli.exe ./cli

.PHONY: buildLinux
buildLinux: frontend
	GOOS=linux GOARCH=amd64 go build -trimpath -ldflags '$(LDFLAGS)' -o bin/linux/automationhub .
	GOOS=linux GOARCH=amd64 go build -trimpath -ldflags '$(LDFLAGS)' -o bin/linux/cli ./cli

.PHONY: server
server:
	go run main.go master --config automationhub.yaml

.PHONY: node
node:
	go run main.go node --config automationhub_slave.yaml

# Bootstrap target — always available even before base/ exists
.PHONY: dev-bootstrap
dev-bootstrap:
	@bash .devcontainer/innoclaude ensure

# dev-help stub: shows bootstrap instructions when base/ is not yet installed.
# Once base/ exists, the real dev-help from dev-container.mk takes over via -include.
ifneq ($(wildcard .devcontainer/base/mk/dev-container.mk),)
else
.PHONY: dev-help
dev-help:
	@echo ""
	@echo "  The base layer is not installed yet."
	@echo ""
	@echo "  make dev-bootstrap    Fetch and install the base layer"
	@echo ""
endif

# DevContainer targets (wildcard picks up all base .mk files)
-include .devcontainer/base/mk/dev-*.mk
-include .devcontainer/base/mk/features/dev-*.mk

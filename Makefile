VERSION := $(shell git describe --tags --abbrev=0)
SRC_DIR=./proto
DST_DIR=./out

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
	cd frontend && BUILD_PATH=../endpoints/web/data/ yarn build && cd ../

# run build for frontend and backend
.PHONY: buildMac
buildMac:
	cd frontend && BUILD_PATH=../endpoints/web/data/ yarn build && cd ../ && go build -trimpath -ldflags '-X github.com/fsuhrau/automationhub/hub.version=$(VERSION)' -o bin/macos/automationhub && cd cli && go build -trimpath -ldflags '-X github.com/fsuhrau/automationhub/hub.version=$(VERSION)' -o ../bin/macos/cli

# run build for frontend and backend for Windows
.PHONY: buildWindows
buildWindows:
	GOOS=windows GOARCH=amd64 cd frontend && BUILD_PATH=../endpoints/web/data/ yarn build && cd ../ && go build -trimpath -ldflags '-X github.com/fsuhrau/automationhub/hub.version=$(VERSION)' -o bin/windows/automationhub.exe && cd cli && go build -trimpath -ldflags '-X github.com/fsuhrau/automationhub/hub.version=$(VERSION)' -o ../bin/windows/cli.exe

# run build for frontend and backend for linux
.PHONY: buildLinux
buildLinux:
	GOOS=linux GOARCH=amd64 cd frontend && BUILD_PATH=../endpoints/web/data/ yarn build && cd ../ && go build -trimpath -ldflags '-X github.com/fsuhrau/automationhub/hub.version=$(VERSION)' -o bin/linux/automationhub && cd cli && go build -trimpath -ldflags '-X github.com/fsuhrau/automationhub/hub.version=$(VERSION)' -o ../bin/linux/cli

.PHONY: server
server:
	go run main.go master --config automationhub.yaml

.PHONY: node
node:
	go run main.go node --config automationhub_slave.yaml

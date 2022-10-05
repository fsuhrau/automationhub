VERSION := $(shell git describe --tags --abbrev=0)

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
.PHONY: build
build:
	cd frontend && BUILD_PATH=../endpoints/web/data/ yarn build && cd ../ && go build -trimpath -ldflags '-X github.com/fsuhrau/automationhub/hub.version=$(VERSION)' -o bin/automationhub && cd cli && go build -trimpath -ldflags '-X github.com/fsuhrau/automationhub/hub.version=$(VERSION)' -o ../bin/cli

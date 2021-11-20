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
	cd frontend && BUILD_PATH=../endpoints/web/data/ yarn build && cd ../ && go build -o bin/automationhub && cd cli && go build -o ../bin/cli

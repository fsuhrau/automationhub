# run the frontend locally
.PHONY: run
run:
	cd frontend && yarn start

# run linter (ESLint)
.PHONY: lint
lint:
	cd frontend && yarn lint --fix

# run build
.PHONY: build
build:
	cd frontend && BUILD_PATH=../endpoints/web/data/ yarn build && cd ../ && go build -o bin/automationhub && cd cli && go build -o ../bin/cli

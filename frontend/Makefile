# run the frontend locally
.PHONY: run
run:
	yarn start

# run linter (ESLint)
.PHONY: lint
lint:
	yarn lint

# run build
.PHONY: build
build:
	mkdir -p ../endpoints/web/ && yarn build && cp -rf build/* ../endpoints/web/

.PHONY: help

help: ## This help.
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help

GCFLAGS=
LDFLAGS_EXTRA=

clean: ## Clean workspace
	rm coverage.out && true

lint: ## Lint code
	go-critic check ./... || true

debug-test: ## Debug test with dlv
	dlv test -- -test.v

connect: ## Connect to debug
	dlv connect 127.0.0.1:36666

test: ## Run tests
	RUN_INTEGRATION=true \
	go run gotest.tools/gotestsum@latest \
	-- \
	-failfast \
	-race ./... -v

test-watch: ## Watch tests
	RUN_INTEGRATION=true \
	go run gotest.tools/gotestsum@latest \
	--format standard-verbose \
	--watch \
	-- \
	-failfast \
	./...

cover: ## Run covarage
	go test -coverprofile=coverage.out
	go tool cover -func=coverage.out
	@rm coverage.out

.PHONY: run

GOCMD=go
GOTEST=$(GOCMD) test

GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
WHITE  := $(shell tput -Txterm setaf 7)
CYAN   := $(shell tput -Txterm setaf 6)
RESET  := $(shell tput -Txterm sgr0)

.PHONY: all test build vendor

all: help

## Test:
test: ## Run all the tests
	$(GOCMD) clean -testcache
	$(GOTEST) -race -timeout 60s ./...

# test-fuzz: ## Run fuzzing tests
# 	$(GOCMD) clean -testcache
# 	$(GOTEST) -v -fuzz=Fuzz -timeout 30s -fuzztime=5s ./...

test-unit: ## Runs only fast tests
	$(GOCMD) clean -testcache
	$(GOTEST) -short -timeout 20s ./...

ci-test-e2e: ## Run end to end tests, only suited for the CI
	$(GOCMD) build -o graviola ./cmd/...
	./graviola --config test/e2e/config.yaml &
	@sleep 3
	$(GOCMD) run ./test/e2e/...

coverage: ## Run the tests of the project and export the coverage
	$(GOCMD) clean -testcache
	$(GOTEST) -timeout 30s -cover -covermode=count -coverprofile=profile.cov ./...
	$(GOCMD) tool cover -func profile.cov

## Lint:
lint: ## Run all available linters
	$(GOCMD) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run

lint-experimental: ## Linters that we are still testings
	@$(GOCMD) install github.com/alexkohler/nakedret/cmd/nakedret@latest
	@$(GOCMD) install github.com/alexkohler/prealloc@latest
	@$(GOCMD) install github.com/ashanbrown/makezero@latest
	@$(GOCMD) install golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@latest
	@$(GOCMD) install github.com/timakin/bodyclose@latest

	@fieldalignment -fix ./...
	@nakedret ./...
	@prealloc -set_exit_status ./...
	@makezero -set_exit_status ./...
	@$(GOCMD) vet -vettool=$(which bodyclose) ./...

## Fmt:
fmt: ## Fixes deprecated APIs and formats the code
	$(GOCMD) tool fix .
	$(GOCMD) fmt ./...

## Security:
vuln:
	$(GOCMD) install golang.org/x/vuln/cmd/govulncheck@latest
	govulncheck ./...

update-deps:
	$(GOCMD) get -u ./...
	$(GOCMD) mod tidy

## Help:
help: ## Show this help.
	@echo ''
	@echo 'Usage:'
	@echo '  ${YELLOW}make${RESET} ${GREEN}<target>${RESET}'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} { \
		if (/^[a-zA-Z_-]+:.*?##.*$$/) {printf "    ${YELLOW}%-20s${GREEN}%s${RESET}\n", $$1, $$2} \
		else if (/^## .*$$/) {printf "  ${CYAN}%s${RESET}\n", substr($$1,4)} \
		}' $(MAKEFILE_LIST)

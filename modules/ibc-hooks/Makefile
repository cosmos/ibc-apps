#!/usr/bin/make -f

# Eveyrhing related to tests
simulate:
	@go test -v -run=TestFullAppSimulation ./simapp -NumBlocks 200 -BlockSize 50 -Commit -Enabled -Period 1

test-unit:
	@VERSION=$(VERSION) go test -mod=readonly -tags='ledger test_ledger_mock' ./tests/...
	
.PHONY: simulate test-unit

# Eveyrhing related to building the smart contracts
optimize-workspace: 
	docker run --rm -v "$(shell pwd)/tests/unit/testdata":/code \
		--mount type=volume,source="$(shell basename $(shell pwd))_cache",target=/code/target \
		--mount type=volume,source=registry_cache,target=/usr/local/cargo/registry \
		cosmwasm/workspace-optimizer:0.12.13

.PHONY: optimize-workspace


###############################################################################
###                                Linting                                  ###
###############################################################################

golangci_lint_cmd=golangci-lint
golangci_version=v1.52.2

lint:
	@echo "--> Running linter"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(golangci_version)
	@./scripts/go-lint-all.bash --timeout=15m

lint-fix:
	@echo "--> Running linter"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(golangci_version)
	@./scripts/go-lint-all.bash --fix
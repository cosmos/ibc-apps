#!/usr/bin/make -f

simulate:
	@go test -v -run=TestFullAppSimulation ./simapp -NumBlocks 200 -BlockSize 50 -Commit -Enabled -Period 1

test-unit:
	@VERSION=$(VERSION) go test -mod=readonly -tags='ledger test_ledger_mock' ./tests/...

integration-test: clean-integration-test-data install
	@echo "Initializing both blockchains..."
	./scripts/tests/start.sh
	@echo "Create relayer..."
	./scripts/tests/relayer/rly-init.sh
	@echo "Transfer coin from chain test-1 to test-2..."
	./scripts/tests/feather/transfer.sh
	@echo "Validate the creation of alliance throught feather..."
	./scripts/tests/feather/validate-alliance.sh
	@echo "Killing feather-cored and removing previous data"
	-@rm -rf ./.test-data
	-@killall feather-cored 2>/dev/null
	-@killall rly 2>/dev/null

clean-integration-test-data:
	@echo "Killing feather-cored and removing previous data"
	-@rm -rf ./.test-data
	-@killall feather-cored 2>/dev/null
	-@killall rly 2>/dev/null
	
.PHONY: simulate test-unit
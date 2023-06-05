#!/usr/bin/make -f

simulate:
	@go test -v -run=TestFullAppSimulation ./simapp -NumBlocks 200 -BlockSize 50 -Commit -Enabled -Period 1

test-unit:
	@VERSION=$(VERSION) go test -mod=readonly -tags='ledger test_ledger_mock' ./tests/...
	
.PHONY: simulate test-unit
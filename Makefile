## mod-tidy: go mod tidy spawn, simapp, and interchaintest with proper go.mod suffixes
mod-tidy:
	cd middleware/packet-forward-middleware && go mod tidy
	cd middleware/packet-forward-middleware/e2e && go mod tidy

	cd modules/async-icq && go mod tidy
	cd modules/async-icq/e2e && go mod tidy

	cd modules/ibc-hooks && go mod tidy
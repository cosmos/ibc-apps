go 1.21

toolchain go1.23.0

module github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7

require (
	cosmossdk.io/api v0.3.1
	cosmossdk.io/errors v1.0.1
	cosmossdk.io/log v1.3.1
	cosmossdk.io/tools/rosetta v0.2.1
	github.com/armon/go-metrics v0.4.1
	github.com/cometbft/cometbft v0.37.5
	github.com/cometbft/cometbft-db v0.8.0
	github.com/cosmos/cosmos-sdk v0.47.13
	github.com/cosmos/gogoproto v1.4.10
	github.com/cosmos/ibc-go/v7 v7.7.0
	github.com/gorilla/mux v1.8.0
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/iancoleman/orderedmap v0.3.0
	github.com/rakyll/statik v0.1.7
	github.com/spf13/cast v1.6.0
	github.com/spf13/cobra v1.8.0
	github.com/stretchr/testify v1.9.0
	go.uber.org/mock v0.4.0
)

replace (
	// cosmos keyring
	github.com/99designs/keyring => github.com/cosmos/keyring v1.2.0

	// support concurrency for iavl
	github.com/cosmos/iavl => github.com/cosmos/iavl v0.20.1

	// dgrijalva/jwt-go is deprecated and doesn't receive security updates.
	// TODO: remove it: https://github.com/cosmos/cosmos-sdk/issues/13134
	github.com/dgrijalva/jwt-go => github.com/golang-jwt/jwt/v4 v4.4.2

	// Fix upstream GHSA-h395-qcrw-5vmq vulnerability.
	// TODO Remove it: https://github.com/cosmos/cosmos-sdk/issues/10409
	github.com/gin-gonic/gin => github.com/gin-gonic/gin v1.9.0

	// https://github.com/cosmos/cosmos-sdk/issues/14949
	// pin the version of goleveldb to v1.0.1-0.20210819022825-2ae1ddf74ef7 required by SDK v47 upgrade guide.
	github.com/syndtr/goleveldb => github.com/syndtr/goleveldb v1.0.1-0.20210819022825-2ae1ddf74ef7

	golang.org/x/exp => golang.org/x/exp v0.0.0-20230711153332-06a737ee72cb
)

// see https://github.com/cosmos/ibc-apps/pull/141
retract v7.1.0

module github.com/quasar-finance/interchain-query-demo

go 1.16

require (
	github.com/cosmos/cosmos-sdk v0.45.8
	github.com/cosmos/ibc-go/v3 v3.0.1
	github.com/gogo/protobuf v1.3.3
	github.com/golang/protobuf v1.5.2
	github.com/gorilla/mux v1.8.0
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/spf13/cast v1.5.0
	github.com/spf13/cobra v1.5.0
	github.com/stretchr/testify v1.8.0
	github.com/tendermint/starport v0.19.3
	github.com/tendermint/tendermint v0.34.21
	github.com/tendermint/tm-db v0.6.6
	golang.org/x/net v0.1.0 // indirect
	google.golang.org/genproto v0.0.0-20221027153422-115e99e71e1c
	google.golang.org/grpc v1.50.1
	gopkg.in/yaml.v2 v2.4.0
)

replace (
	// used latest commit of branch v3.3.0-icq
	github.com/cosmos/ibc-go/v3 => github.com/strangelove-ventures/ibc-go/v3 v3.0.0-20221014082552-99c8caa484af
	github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
	github.com/keybase/go-keychain => github.com/99designs/go-keychain v0.0.0-20191008050251-8e49817e8af4
	google.golang.org/grpc => google.golang.org/grpc v1.33.2
)

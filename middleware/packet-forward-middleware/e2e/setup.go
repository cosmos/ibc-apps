package e2e

import (
	"fmt"
	"os"
	"strings"

	testutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	connectiontypes "github.com/cosmos/ibc-go/v7/modules/core/03-connection/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	ibclocalhost "github.com/cosmos/ibc-go/v7/modules/light-clients/09-localhost"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
)

var (
	pfmRepo, pfmVersion = GetDockerImageInfo()
	PFMImage            = ibc.DockerImage{
		Repository: pfmRepo,
		Version:    pfmVersion,
		UidGid:     "1025:1025",
	}

	Denom               = "token"
	GenesisWalletAmount = int64(10_000_000)

	DefaultConfig = ibc.ChainConfig{
		Type:                   "cosmos",
		Name:                   "pfm",
		ChainID:                "pfm-1",
		Images:                 []ibc.DockerImage{PFMImage},
		Bin:                    "simd",
		Bech32Prefix:           "cosmos",
		Denom:                  Denom,
		CoinType:               "118",
		GasPrices:              fmt.Sprintf("0%s", Denom),
		GasAdjustment:          2.0,
		TrustingPeriod:         "112h",
		NoHostMount:            false,
		ConfigFileOverrides:    nil,
		EncodingConfig:         encoding(),
		UsingNewGenesisCommand: true,
	}
)

func encoding() *testutil.TestEncodingConfig {
	cfg := cosmos.DefaultEncoding()

	// TODO: Getting this error
	// Error: error creating clients: failed to create client on src chain{chain-c}: failed to find a matching client for the new client state: no concrete type registered for type URL /ibc.lightclients.localhost.v2.ClientState against interface *exported.ClientState

	// register custom types
	ibclocalhost.RegisterInterfaces(cfg.InterfaceRegistry)
	transfertypes.RegisterInterfaces(cfg.InterfaceRegistry)
	clienttypes.RegisterInterfaces(cfg.InterfaceRegistry)
	connectiontypes.RegisterInterfaces(cfg.InterfaceRegistry)
	channeltypes.RegisterInterfaces(cfg.InterfaceRegistry)

	return &cfg
}

// GetDockerImageInfo returns the appropriate repo and branch version string for integration with the CI pipeline.
// The remote runner sets the BRANCH_CI env var. If present, interchaintest will use the docker image pushed up to the repo.
// If testing locally, user should run `make local-image` and interchaintest will use the local image.
func GetDockerImageInfo() (repo, version string) {
	branchVersion, found := os.LookupEnv("BRANCH_CI")
	repo = "strangelove-ventures/pfm"

	// github action
	if !found {
		repo = "pfm"
		branchVersion = "local"
	}

	// github converts / to - for pushed docker images
	branchVersion = strings.ReplaceAll(branchVersion, "/", "-")
	return repo, branchVersion
}

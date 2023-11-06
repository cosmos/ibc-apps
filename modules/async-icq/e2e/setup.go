package e2e

import (
	"fmt"
	"os"
	"strings"

	testutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
)

var (
	icqRepo, icqVersion = GetDockerImageInfo()
	ICQImage            = ibc.DockerImage{
		Repository: icqRepo,
		Version:    icqVersion,
		UidGid:     "1025:1025",
	}

	Denom         = "utoken"
	DefaultConfig = ibc.ChainConfig{
		Type:           "cosmos",
		Name:           "icq",
		ChainID:        "icq-1",
		Images:         []ibc.DockerImage{ICQImage},
		Bin:            "simd",
		Bech32Prefix:   "cosmos",
		Denom:          Denom,
		CoinType:       "118",
		GasPrices:      fmt.Sprintf("0.0%s", Denom),
		GasAdjustment:  2.0,
		TrustingPeriod: "336h",
		NoHostMount:    false,
		EncodingConfig: encoding(),
	}

	DefaultRelayer = ibc.DockerImage{
		Repository: "ghcr.io/cosmos/relayer",
		Version:    "v2.4.2",
		UidGid:     "1025:1025",
	}
)

func encoding() *testutil.TestEncodingConfig {
	cfg := cosmos.DefaultEncoding()

	return &cfg
}

// GetDockerImageInfo returns the appropriate repo and branch version string for integration with the CI pipeline.
// The remote runner sets the BRANCH_CI env var. If present, interchaintest will use the docker image pushed up to the repo.
// If testing locally, user should run `make local-image` and interchaintest will use the local image.
func GetDockerImageInfo() (repo, version string) {
	branchVersion, found := os.LookupEnv("BRANCH_CI")
	repo = "ibc-apps/async-icq"

	// github action
	if !found {
		repo = "icq-host"
		branchVersion = "local"
	}

	// github converts / to - for pushed docker images
	branchVersion = strings.ReplaceAll(branchVersion, "/", "-")
	return repo, branchVersion
}

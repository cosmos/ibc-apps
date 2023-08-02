package simapp

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/baseapp"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	simcli "github.com/cosmos/cosmos-sdk/x/simulation/client/cli"
)

// Hardcoded chainID for simulation.
const (
	simulationAppChainID = "simulation-app"
	simulationDirPrefix  = "leveldb-app-sim"
	simulationDbName     = "Simulation"
)

func init() {
	simcli.GetSimulatorFlags()
}

// Running as a go test:
//
// go test -v -run=TestFullAppSimulation ./app -NumBlocks 200 -BlockSize 50 -Commit -Enabled -Period 1 -Seed 40
func TestFullAppSimulation(t *testing.T) {
	config := simcli.NewConfigFromFlags()
	config.ChainID = simulationAppChainID

	if !simcli.FlagEnabledValue {
		t.Skip("skipping application simulation")
	}

	db, dir, logger, _, err := simtestutil.SetupSimulation(
		config,
		simulationDirPrefix,
		simulationDbName,
		simcli.FlagVerboseValue,
		true, // Don't use this as it is confusing
	)
	require.NoError(t, err, "simulation setup failed")

	defer func() {
		require.NoError(t, db.Close())
		require.NoError(t, os.RemoveAll(dir))
	}()

	app := NewSimApp(logger,
		db,
		nil,
		true,
		map[int64]bool{},
		DefaultNodeHome,
		simcli.FlagPeriodValue,
		MakeEncodingConfig(),
		simtestutil.EmptyAppOptions{},
		baseapp.SetChainID(simulationAppChainID),
	)
	require.Equal(t, AppName, app.Name())

	// run randomized simulation
	_, simParams, simErr := simulation.SimulateFromSeed(
		t,
		os.Stdout,
		app.BaseApp,
		simtestutil.AppStateFn(app.AppCodec(), app.SimulationManager(), app.DefaultGenesis()),
		simtypes.RandomAccounts,
		simtestutil.SimulationOperations(app, app.AppCodec(), config),
		app.BankKeeper.GetBlockedAddresses(),
		config,
		app.AppCodec(),
	)

	// export state and simParams before the simulatino error is checked
	err = simtestutil.CheckExportSimulation(app, config, simParams)
	require.NoError(t, err)
	require.NoError(t, simErr)

	if config.Commit {
		simtestutil.PrintStats(db)
	}
}

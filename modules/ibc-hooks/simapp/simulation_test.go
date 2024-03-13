package simapp

import (
	"os"
	"testing"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
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
	config, db, _, app := setupSimulationApp(t, "skipping application simulation")
	// run randomized simulation
	_, simParams, simErr := simulation.SimulateFromSeed(
		t,
		os.Stdout,
		app.BaseApp,
		simtestutil.AppStateFn(app.AppCodec(), app.SimulationManager(), app.DefaultGenesis()),
		simtypes.RandomAccounts, // Replace with own random account function if using keys other than secp256k1
		simtestutil.SimulationOperations(app, app.AppCodec(), config),
		BlockedAddresses(),
		config,
		app.AppCodec(),
	)

	// export state and simParams before the simulation error is checked
	err := simtestutil.CheckExportSimulation(app, config, simParams)
	require.NoError(t, err)
	require.NoError(t, simErr)

	if config.Commit {
		simtestutil.PrintStats(db)
	}
}

func setupSimulationApp(t *testing.T, msg string) (simtypes.Config, dbm.DB, simtestutil.AppOptionsMap, *App) {
	t.Helper()
	config := simcli.NewConfigFromFlags()
	config.ChainID = simulationAppChainID

	db, dir, logger, skip, err := simtestutil.SetupSimulation(config, "leveldb-app-sim", "Simulation", simcli.FlagVerboseValue, simcli.FlagEnabledValue)
	if skip {
		t.Skip(msg)
	}
	require.NoError(t, err, "simulation setup failed")

	t.Cleanup(func() {
		require.NoError(t, db.Close())
		require.NoError(t, os.RemoveAll(dir))
	})

	appOptions := make(simtestutil.AppOptionsMap, 0)
	appOptions[flags.FlagHome] = dir // ensure a unique folder
	appOptions[server.FlagInvCheckPeriod] = simcli.FlagPeriodValue

	app := NewSimApp(logger, db, nil, true, appOptions, baseapp.SetChainID(simulationAppChainID))
	require.Equal(t, "SimApp", app.Name())
	return config, db, appOptions, app
}

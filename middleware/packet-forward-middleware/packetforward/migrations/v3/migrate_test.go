package v3_test

import (
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	v3 "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v8/packetforward/migrations/v3"
	"github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v8/test/mock"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestMigrate(t *testing.T) {
	ctrl := gomock.NewController(t)

	channelKeeper := mock.NewMockChannelKeeper(ctrl)
	bankKeeper := mock.NewMockBankKeeper(ctrl)
	transferKeeper := mock.NewMockTransferKeeper(ctrl)

	logger := log.NewTestLogger(t)
	logger.Debug("initializing test setup")

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())

	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, logger)

	// Expected mock calls
	// 1. Iterate over all IBC transfer channels
	// 2. For each channel, get the escrow address and corresponding bank balance
	// 3. Update the escrow amount in transfer keeper state

	// Test addresses
	escrowAddresses := []string{
		transfertypes.GetEscrowAddress("transfer", "channel-0").String(),
		transfertypes.GetEscrowAddress("transfer", "channel-1").String(),
		transfertypes.GetEscrowAddress("transfer", "channel-2").String(),
	}

	// Test various scenarios of ibc transfer module escrow state that match
	// or do not match the bank balances. The provided bank balances are both
	// the values are used to correct the escrow state in the migration and also
	// to verify the correctness of the new escrow state afterwards.
	tests := []struct {
		name                    string
		giveTransferEscrowState map[string]sdk.Coin  // denom -> escrow amount
		bankBalances            map[string]sdk.Coins // escrow address -> bank balance
	}{
		{
			name:                    "empty channels",
			giveTransferEscrowState: map[string]sdk.Coin{},
			bankBalances:            map[string]sdk.Coins{},
		},
		{
			name: "balanced - 1 channel",
			giveTransferEscrowState: map[string]sdk.Coin{
				"meow": sdk.NewInt64Coin("meow", 100),
			},
			bankBalances: map[string]sdk.Coins{
				escrowAddresses[0]: sdk.NewCoins(sdk.NewInt64Coin("meow", 100)),
			},
		},
		{
			name: "balanced - multi channel",
			giveTransferEscrowState: map[string]sdk.Coin{
				"meow": sdk.NewInt64Coin("meow", 100),
				"woof": sdk.NewInt64Coin("woof", 200),
			},
			bankBalances: map[string]sdk.Coins{
				escrowAddresses[0]: sdk.NewCoins(sdk.NewInt64Coin("meow", 100)),
				escrowAddresses[1]: sdk.NewCoins(sdk.NewInt64Coin("woof", 200)),
			},
		},
		{
			name: "underbalanced escrow state - single channel",
			giveTransferEscrowState: map[string]sdk.Coin{
				"meow": sdk.NewInt64Coin("meow", 80),
			},
			bankBalances: map[string]sdk.Coins{
				escrowAddresses[0]: sdk.NewCoins(sdk.NewInt64Coin("meow", 100)),
			},
		},
		{
			name: "underbalanced escrow state - multi channel",
			giveTransferEscrowState: map[string]sdk.Coin{
				"meow": sdk.NewInt64Coin("meow", 80),
				"woof": sdk.NewInt64Coin("woof", 180),
			},
			bankBalances: map[string]sdk.Coins{
				escrowAddresses[0]: sdk.NewCoins(sdk.NewInt64Coin("meow", 100)),
				escrowAddresses[1]: sdk.NewCoins(sdk.NewInt64Coin("woof", 200)),
			},
		},
		{
			name: "overbalanced escrow state - single channel",
			giveTransferEscrowState: map[string]sdk.Coin{
				"meow": sdk.NewInt64Coin("meow", 120),
			},
			bankBalances: map[string]sdk.Coins{
				escrowAddresses[0]: sdk.NewCoins(sdk.NewInt64Coin("meow", 100)),
			},
		},
		{
			name: "overbalanced escrow state - multi channel",
			giveTransferEscrowState: map[string]sdk.Coin{
				"meow": sdk.NewInt64Coin("meow", 120),
				"woof": sdk.NewInt64Coin("woof", 230),
			},
			bankBalances: map[string]sdk.Coins{
				escrowAddresses[0]: sdk.NewCoins(sdk.NewInt64Coin("meow", 100)),
				escrowAddresses[1]: sdk.NewCoins(sdk.NewInt64Coin("woof", 200)),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectedTotalEscrowed := sdk.NewCoins()

			// Initialize mock calls
			// 1. Fetch each escrow bank balances for the EXPECTED state values.
			for _, escrowAddress := range escrowAddresses {
				expectedBalance := tt.bankBalances[escrowAddress]

				bankKeeper.EXPECT().
					GetAllBalances(ctx, sdk.AccAddress(escrowAddress)).
					Return(expectedBalance).
					Times(1)

				// Aggregate escrow balances
				expectedTotalEscrowed = expectedTotalEscrowed.Add(expectedBalance...)
			}

			// 2. Update the escrow state in the transfer keeper, for each denom
			// from the aggregated escrow balances.
			for _, escrowCoin := range expectedTotalEscrowed {
				transferKeeper.EXPECT().
					SetTotalEscrowForDenom(
						ctx,
						escrowCoin,
					).
					Times(1)
			}

			err := v3.Migrate(ctx, bankKeeper, channelKeeper, transferKeeper)
			require.NoError(t, err)
		})
	}
}

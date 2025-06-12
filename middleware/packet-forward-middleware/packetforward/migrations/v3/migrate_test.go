package v3_test

import (
	"github.com/golang/mock/gomock"
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	v3 "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v10/packetforward/migrations/v3"
	"github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v10/test/mock"
	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
	"github.com/stretchr/testify/require"

	tmproto "github.com/cometbft/cometbft/api/cometbft/types/v2"
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

	// Test addresses
	escrowChannels := []string{
		"channel-0",
		"channel-1",
		"channel-2",
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
				escrowChannels[0]: sdk.NewCoins(sdk.NewInt64Coin("meow", 100)),
			},
		},
		{
			name: "balanced - multi channel",
			giveTransferEscrowState: map[string]sdk.Coin{
				"meow": sdk.NewInt64Coin("meow", 100),
				"woof": sdk.NewInt64Coin("woof", 200),
			},
			bankBalances: map[string]sdk.Coins{
				escrowChannels[0]: sdk.NewCoins(sdk.NewInt64Coin("meow", 100)),
				escrowChannels[1]: sdk.NewCoins(sdk.NewInt64Coin("woof", 200)),
			},
		},
		{
			name: "underbalanced escrow state - single channel",
			giveTransferEscrowState: map[string]sdk.Coin{
				"meow": sdk.NewInt64Coin("meow", 80),
			},
			bankBalances: map[string]sdk.Coins{
				escrowChannels[0]: sdk.NewCoins(sdk.NewInt64Coin("meow", 100)),
			},
		},
		{
			name: "underbalanced escrow state - multi channel",
			giveTransferEscrowState: map[string]sdk.Coin{
				"meow": sdk.NewInt64Coin("meow", 80),
				"woof": sdk.NewInt64Coin("woof", 180),
			},
			bankBalances: map[string]sdk.Coins{
				escrowChannels[0]: sdk.NewCoins(sdk.NewInt64Coin("meow", 100)),
				escrowChannels[1]: sdk.NewCoins(sdk.NewInt64Coin("woof", 200)),
			},
		},
		// Escrow module state shouldn't be overbalanced, but we test it here
		// to ensure the migration correctly updates the escrow state even in
		// this case.
		{
			name: "overbalanced escrow state - single channel",
			giveTransferEscrowState: map[string]sdk.Coin{
				"meow": sdk.NewInt64Coin("meow", 120),
			},
			bankBalances: map[string]sdk.Coins{
				escrowChannels[0]: sdk.NewCoins(sdk.NewInt64Coin("meow", 100)),
			},
		},
		{
			name: "overbalanced escrow state - multi channel",
			giveTransferEscrowState: map[string]sdk.Coin{
				"meow": sdk.NewInt64Coin("meow", 120),
				"woof": sdk.NewInt64Coin("woof", 230),
			},
			bankBalances: map[string]sdk.Coins{
				escrowChannels[0]: sdk.NewCoins(sdk.NewInt64Coin("meow", 100)),
				escrowChannels[1]: sdk.NewCoins(sdk.NewInt64Coin("woof", 200)),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectedTotalEscrowed := sdk.NewCoins()

			// Initialize mock calls
			// 1. Fetch each escrow bank balances for the EXPECTED state values.

			portID := "transfer"
			channels := []channeltypes.IdentifiedChannel{}
			for _, channel := range escrowChannels {
				channels = append(channels, channeltypes.NewIdentifiedChannel(portID, channel, channeltypes.Channel{}))
			}

			transferKeeper.EXPECT().GetPort(ctx).Return(portID).Times(1)
			channelKeeper.EXPECT().
				GetAllChannelsWithPortPrefix(ctx, portID).
				Return(channels).
				Times(1)

			// 2. Aggregate the bank balances to calculate the expected total escrowed
			for _, escrowChannel := range escrowChannels {
				expectedBalance := tt.bankBalances[escrowChannel]

				escrowAddress := transfertypes.GetEscrowAddress(portID, escrowChannel)

				bankKeeper.EXPECT().
					GetAllBalances(ctx, escrowAddress).
					Return(expectedBalance).
					Times(1)

				// Aggregate escrow balances
				expectedTotalEscrowed = expectedTotalEscrowed.Add(expectedBalance...)
			}

			// 3. Update the escrow state in the transfer keeper, for each denom
			// from the aggregated escrow balances.
			for _, escrowCoin := range expectedTotalEscrowed {
				transferKeeper.EXPECT().
					GetTotalEscrowForDenom(ctx, escrowCoin.Denom).
					Return(sdk.Coin{}).
					Times(1)

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

package test

import (
	"testing"
	"time"

	"github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v4/packetforward"
	"github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v4/packetforward/keeper"
	"github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v4/packetforward/types"
	"github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v4/test/mock"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmdb "github.com/tendermint/tm-db"
	"go.uber.org/mock/gomock"

	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	porttypes "github.com/cosmos/ibc-go/v4/modules/core/05-port/types"
)

func NewTestSetup(t *testing.T, ctl *gomock.Controller) *Setup {
	t.Helper()

	initializer := newInitializer()

	transferKeeperMock := mock.NewMockTransferKeeper(ctl)
	channelKeeperMock := mock.NewMockChannelKeeper(ctl)
	bankKeeperMock := mock.NewMockBankKeeper(ctl)
	ibcModuleMock := mock.NewMockIBCModule(ctl)
	ics4WrapperMock := mock.NewMockICS4Wrapper(ctl)

	packetforwardKeeper := initializer.packetforwardKeeper(transferKeeperMock, channelKeeperMock, bankKeeperMock, ics4WrapperMock)

	require.NoError(t, initializer.StateStore.LoadLatestVersion())

<<<<<<< HEAD
	packetforwardKeeper.SetParams(initializer.Ctx, types.DefaultParams())

=======
>>>>>>> 26d8080 (refactor: remove the ability to take a fee for each forwarded packet (#202))
	return &Setup{
		Initializer: initializer,

		Keepers: &testKeepers{
			PacketForwardKeeper: packetforwardKeeper,
		},

		Mocks: &testMocks{
<<<<<<< HEAD
			TransferKeeperMock:     transferKeeperMock,
			ChannelKeeperMock:      channelKeeperMock,
			DistributionKeeperMock: distributionKeeperMock,
			IBCModuleMock:          ibcModuleMock,
			ICS4WrapperMock:        ics4WrapperMock,
=======
			TransferKeeperMock: transferKeeperMock,
			IBCModuleMock:      ibcModuleMock,
			ICS4WrapperMock:    ics4WrapperMock,
>>>>>>> 26d8080 (refactor: remove the ability to take a fee for each forwarded packet (#202))
		},

		ForwardMiddleware: initializer.forwardMiddleware(ibcModuleMock, packetforwardKeeper, 0, keeper.DefaultForwardTransferPacketTimeoutTimestamp, keeper.DefaultRefundTransferPacketTimeoutTimestamp),
	}
}

type Setup struct {
	Initializer initializer

	Keepers *testKeepers
	Mocks   *testMocks

	ForwardMiddleware packetforward.IBCMiddleware
}

type testKeepers struct {
	PacketForwardKeeper *keeper.Keeper
}

type testMocks struct {
<<<<<<< HEAD
	TransferKeeperMock     *mock.MockTransferKeeper
	ChannelKeeperMock      *mock.MockChannelKeeper
	DistributionKeeperMock *mock.MockDistributionKeeper
	IBCModuleMock          *mock.MockIBCModule
	ICS4WrapperMock        *mock.MockICS4Wrapper
=======
	TransferKeeperMock *mock.MockTransferKeeper
	IBCModuleMock      *mock.MockIBCModule
	ICS4WrapperMock    *mock.MockICS4Wrapper
>>>>>>> 26d8080 (refactor: remove the ability to take a fee for each forwarded packet (#202))
}

type initializer struct {
	DB         *tmdb.MemDB
	StateStore store.CommitMultiStore
	Ctx        sdk.Context
	Marshaler  codec.Codec
	Amino      *codec.LegacyAmino
}

// Create an initializer with in memory database and default codecs
func newInitializer() initializer {
	logger := log.TestingLogger()
	logger.Debug("initializing test setup")

	db := tmdb.NewMemDB()
	stateStore := store.NewCommitMultiStore(db)

	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, logger)
	interfaceRegistry := cdctypes.NewInterfaceRegistry()
	amino := codec.NewLegacyAmino()
	marshaler := codec.NewProtoCodec(interfaceRegistry)

	return initializer{
		DB:         db,
		StateStore: stateStore,
		Ctx:        ctx,
		Marshaler:  marshaler,
		Amino:      amino,
	}
}

<<<<<<< HEAD
func (i initializer) paramsKeeper() paramskeeper.Keeper {
	storeKey := sdk.NewKVStoreKey(paramstypes.StoreKey)
	transientStoreKey := sdk.NewTransientStoreKey(paramstypes.TStoreKey)
	i.StateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, i.DB)
	i.StateStore.MountStoreWithDB(transientStoreKey, storetypes.StoreTypeTransient, i.DB)

	paramsKeeper := paramskeeper.NewKeeper(i.Marshaler, i.Amino, storeKey, transientStoreKey)

	return paramsKeeper
}

=======
>>>>>>> 26d8080 (refactor: remove the ability to take a fee for each forwarded packet (#202))
func (i initializer) packetforwardKeeper(
	transferKeeper types.TransferKeeper,
	channelKeeper types.ChannelKeeper,
	bankKeeper types.BankKeeper,
	ics4Wrapper porttypes.ICS4Wrapper,
) *keeper.Keeper {
	storeKey := sdk.NewKVStoreKey(types.StoreKey)
	i.StateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, i.DB)

	subspace := paramsKeeper.Subspace(types.ModuleName)
	packetforwardKeeper := keeper.NewKeeper(
		i.Marshaler,
		storeKey,
		subspace,
		transferKeeper,
		channelKeeper,
		bankKeeper,
		ics4Wrapper,
	)

	return packetforwardKeeper
}

func (i initializer) forwardMiddleware(app porttypes.IBCModule, k *keeper.Keeper, retriesOnTimeout uint8, forwardTimeout time.Duration, refundTimeout time.Duration) packetforward.IBCMiddleware {
	return packetforward.NewIBCMiddleware(app, k, retriesOnTimeout, forwardTimeout, refundTimeout)
}

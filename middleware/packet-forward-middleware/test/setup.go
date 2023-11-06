package test

import (
	"testing"
	"time"

	"github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/packetforward"
	"github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/packetforward/keeper"
	"github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/packetforward/types"
	"github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/test/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"

	tmdb "github.com/cometbft/cometbft-db"
	"github.com/cometbft/cometbft/libs/log"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	porttypes "github.com/cosmos/ibc-go/v7/modules/core/05-port/types"
)

func NewTestSetup(t *testing.T, ctl *gomock.Controller) *Setup {
	t.Helper()
	initializer := newInitializer()

	transferKeeperMock := mock.NewMockTransferKeeper(ctl)
	channelKeeperMock := mock.NewMockChannelKeeper(ctl)
	distributionKeeperMock := mock.NewMockDistributionKeeper(ctl)
	bankKeeperMock := mock.NewMockBankKeeper(ctl)
	ibcModuleMock := mock.NewMockIBCModule(ctl)
	ics4WrapperMock := mock.NewMockICS4Wrapper(ctl)

	paramsKeeper := initializer.paramsKeeper()
	packetforwardKeeper := initializer.packetforwardKeeper(paramsKeeper, transferKeeperMock, channelKeeperMock, distributionKeeperMock, bankKeeperMock, ics4WrapperMock)

	require.NoError(t, initializer.StateStore.LoadLatestVersion())

	if err := packetforwardKeeper.SetParams(initializer.Ctx, types.DefaultParams()); err != nil {
		t.Fatal(err)
	}

	return &Setup{
		Initializer: initializer,

		Keepers: &testKeepers{
			ParamsKeeper:        &paramsKeeper,
			PacketForwardKeeper: packetforwardKeeper,
		},

		Mocks: &testMocks{
			TransferKeeperMock:     transferKeeperMock,
			DistributionKeeperMock: distributionKeeperMock,
			IBCModuleMock:          ibcModuleMock,
			ICS4WrapperMock:        ics4WrapperMock,
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
	ParamsKeeper        *paramskeeper.Keeper
	PacketForwardKeeper *keeper.Keeper
}

type testMocks struct {
	TransferKeeperMock     *mock.MockTransferKeeper
	DistributionKeeperMock *mock.MockDistributionKeeper
	IBCModuleMock          *mock.MockIBCModule
	ICS4WrapperMock        *mock.MockICS4Wrapper
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

func (i initializer) paramsKeeper() paramskeeper.Keeper {
	storeKey := sdk.NewKVStoreKey(paramstypes.StoreKey)
	transientStoreKey := sdk.NewTransientStoreKey(paramstypes.TStoreKey)
	i.StateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, i.DB)
	i.StateStore.MountStoreWithDB(transientStoreKey, storetypes.StoreTypeTransient, i.DB)

	paramsKeeper := paramskeeper.NewKeeper(i.Marshaler, i.Amino, storeKey, transientStoreKey)

	return paramsKeeper
}

func (i initializer) packetforwardKeeper(
	paramsKeeper paramskeeper.Keeper,
	transferKeeper types.TransferKeeper,
	channelKeeper types.ChannelKeeper,
	distributionKeeper types.DistributionKeeper,
	bankKeeper types.BankKeeper,
	ics4Wrapper porttypes.ICS4Wrapper,
) *keeper.Keeper {
	storeKey := sdk.NewKVStoreKey(types.StoreKey)
	i.StateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, i.DB)

	govModuleAddress := "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn"

	packetforwardKeeper := keeper.NewKeeper(
		i.Marshaler,
		storeKey,
		transferKeeper,
		channelKeeper,
		distributionKeeper,
		bankKeeper,
		ics4Wrapper,
		govModuleAddress,
	)

	return packetforwardKeeper
}

func (i initializer) forwardMiddleware(app porttypes.IBCModule, k *keeper.Keeper, retriesOnTimeout uint8, forwardTimeout time.Duration, refundTimeout time.Duration) packetforward.IBCMiddleware {
	return packetforward.NewIBCMiddleware(app, k, retriesOnTimeout, forwardTimeout, refundTimeout)
}

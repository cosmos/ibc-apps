package test

import (
	"testing"
	"time"

	"github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/router"
	"github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/router/keeper"
	"github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/router/types"
	"github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/test/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

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
	authKeeperMock := mock.NewMockAuthKeeper(ctl)

	paramsKeeper := initializer.paramsKeeper()
	routerKeeper := initializer.routerKeeper(paramsKeeper, transferKeeperMock, channelKeeperMock, distributionKeeperMock, bankKeeperMock, ics4WrapperMock, authKeeperMock)
	// routerModule := initializer.routerModule(routerKeeper)

	require.NoError(t, initializer.StateStore.LoadLatestVersion())

	routerKeeper.SetParams(initializer.Ctx, types.DefaultParams())

	return &Setup{
		Initializer: initializer,

		Keepers: &testKeepers{
			ParamsKeeper: &paramsKeeper,
			RouterKeeper: routerKeeper,
		},

		Mocks: &testMocks{
			TransferKeeperMock:     transferKeeperMock,
			DistributionKeeperMock: distributionKeeperMock,
			IBCModuleMock:          ibcModuleMock,
			ICS4WrapperMock:        ics4WrapperMock,
			AuthKeeperMock:         authKeeperMock,
		},

		ForwardMiddleware: initializer.forwardMiddleware(ibcModuleMock, routerKeeper, 0, keeper.DefaultForwardTransferPacketTimeoutTimestamp, keeper.DefaultRefundTransferPacketTimeoutTimestamp),
	}
}

type Setup struct {
	Initializer initializer

	Keepers *testKeepers
	Mocks   *testMocks

	ForwardMiddleware router.IBCMiddleware
}

type testKeepers struct {
	ParamsKeeper *paramskeeper.Keeper
	RouterKeeper *keeper.Keeper
}

type testMocks struct {
	TransferKeeperMock     *mock.MockTransferKeeper
	DistributionKeeperMock *mock.MockDistributionKeeper
	IBCModuleMock          *mock.MockIBCModule
	ICS4WrapperMock        *mock.MockICS4Wrapper
	AuthKeeperMock         *mock.MockAuthKeeper
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

func (i initializer) routerKeeper(
	paramsKeeper paramskeeper.Keeper,
	transferKeeper types.TransferKeeper,
	channelKeeper types.ChannelKeeper,
	distributionKeeper types.DistributionKeeper,
	bankKeeper types.BankKeeper,
	ics4Wrapper porttypes.ICS4Wrapper,
	authKeeper types.AuthKeeper,
) *keeper.Keeper {
	storeKey := sdk.NewKVStoreKey(types.StoreKey)
	i.StateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, i.DB)

	subspace := paramsKeeper.Subspace(types.ModuleName)
	routerKeeper := keeper.NewKeeper(
		i.Marshaler,
		storeKey,
		subspace,
		transferKeeper,
		channelKeeper,
		distributionKeeper,
		bankKeeper,
		ics4Wrapper,
		authKeeper,
	)

	return routerKeeper
}

func (i initializer) forwardMiddleware(app porttypes.IBCModule, k *keeper.Keeper, retriesOnTimeout uint8, forwardTimeout time.Duration, refundTimeout time.Duration) router.IBCMiddleware {
	return router.NewIBCMiddleware(app, k, retriesOnTimeout, forwardTimeout, refundTimeout)
}

# Integration
This document provides instructions on integrating and configuring the Packet Forward Middleware (PFM) within your
existing chain implementation. This document is _NOT_ a guide on developing with the Cosmos SDK or ibc-go and makes
the assumption that you have some existing codebase for your chain with IBC already enabled.

## Example integration of the Packet Forward Middleware

```go
// app.go

// Import the packet forward middleware
import (
    packetforward "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/router"
    packetforwardkeeper "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/router/keeper"
    packetforwardtypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/router/types"
)

...

// Register the AppModule for the packet forward middleware module
ModuleBasics = module.NewBasicManager(
    ...
    packetforward.AppModuleBasic{},
    ...
)

...

// Add packet forward middleware Keeper
type App struct {
	...
	PacketForwardKeeper *packetforwardkeeper.Keeper
	...
}

...

// Create store keys 
keys := sdk.NewKVStoreKeys(
    ...
    packetforwardtypes.StoreKey,
    ...
)

...

// Initialize the packet forward middleware Keeper
// It's important to note that the PFM Keeper must be initialized before the Transfer Keeper
app.PacketForwardKeeper = packetforwardkeeper.NewKeeper(
    appCodec,
    keys[packetforwardtypes.StoreKey],
    app.GetSubspace(packetforwardtypes.ModuleName),
    app.TransferKeeper, // will be zero-value here, reference is set later on with SetTransferKeeper.
    app.IBCKeeper.ChannelKeeper,
    appKeepers.DistrKeeper,
    app.BankKeeper,
    app.IBCKeeper.ChannelKeeper,
)

// Initialize the transfer module Keeper
app.TransferKeeper = ibctransferkeeper.NewKeeper(
    appCodec,
    keys[ibctransfertypes.StoreKey],
    app.GetSubspace(ibctransfertypes.ModuleName),
    app.PacketForwardKeeper,
    app.IBCKeeper.ChannelKeeper,
    &app.IBCKeeper.PortKeeper,
    app.AccountKeeper,
    app.BankKeeper,
    scopedTransferKeeper,
)

app.PacketForwardKeeper.SetTransferKeeper(app.TransferKeeper)

// See the section below for configuring an application stack with the packet forward middleware 

...

// Register packet forward middleware AppModule
app.moduleManager = module.NewManager(
    ...
    packetforward.NewAppModule(app.PacketForwardKeeper),
)

...

// Add packet forward middleware to begin blocker logic
app.moduleManager.SetOrderBeginBlockers(
    ...
    packetforwardtypes.ModuleName,
    ...
)

// Add packet forward middleware to end blocker logic
app.moduleManager.SetOrderEndBlockers(
    ...
    packetforwardtypes.ModuleName,
    ...
)

// Add packet forward middleware to init genesis logic
app.moduleManager.SetOrderInitGenesis(
    ...
    ibcfeetypes.ModuleName,
    ...
)

// Add packet forward middleware to init params keeper
func initParamsKeeper(appCodec codec.BinaryCodec, legacyAmino *codec.LegacyAmino, key, tkey storetypes.StoreKey) paramskeeper.Keeper {
    ...
    paramsKeeper.Subspace(packetforwardtypes.ModuleName).WithKeyTable(packetforwardtypes.ParamKeyTable())
    ...
}
```

## Configuring the transfer application stack with Packet Forward Middleware

Here is an example of how to create an application stack using `transfer` and `packet-forward-middleware`. 
The following `transferStack` is configured in `app/app.go` and added to the IBC `Router`. 
The in-line comments describe the execution flow of packets between the application stack and IBC core.

For more information on configuring an IBC application stack see the ibc-go docs [here](https://github.com/cosmos/ibc-go/blob/main/docs/middleware/ics29-fee/integration.md#configuring-an-application-stack-with-fee-middleware).


```go
// Create Transfer Stack
// SendPacket, since it is originating from the application to core IBC:
// transferKeeper.SendPacket -> packetforward.SendPacket -> channel.SendPacket

// RecvPacket, message that originates from core IBC and goes down to app, the flow is the other way
// channel.RecvPacket -> packetforward.OnRecvPacket -> transfer.OnRecvPacket

// transfer stack contains (from top to bottom):
// - Packet Forward Middleware
// - Transfer
var transferStack ibcporttypes.IBCModule
transferStack = transfer.NewIBCModule(app.TransferKeeper)
transferStack = packetforward.NewIBCMiddleware(
	transferStack,
	app.PacketForwardKeeper,
	0, // retries on timeout
	packetforwardkeeper.DefaultForwardTransferPacketTimeoutTimestamp, // forward timeout
	packetforwardkeeper.DefaultRefundTransferPacketTimeoutTimestamp, // refund timeout
)

// Add transfer stack to IBC Router
ibcRouter.AddRoute(ibctransfertypes.ModuleName, transferStack)
```

## Configurable options in the Packet Forward Middleware

Provide description of these configurable params

- Retries
- Timeouts
- Fee distribution to community pool
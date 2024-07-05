# IBC-hooks

> This module is forked from https://github.com/osmosis-labs/osmosis/tree/main/x/ibc-hooks

## IBC hooks

The IBC hooks module is an IBC middleware that enables ICS-20 token transfers to initiate contract calls.
This functionality allows cross-chain contract calls that involve token movement.
IBC hooks are useful for a variety of use cases, including cross-chain swaps, which are an extremely powerful primitive.

## How do IBC hooks work?

IBC hooks are made possible through the `memo` field included in every ICS-20 transfer packet, as introduced in [IBC v3.4.0](https://medium.com/the-interchain-foundation/moving-beyond-simple-token-transfers-d42b2b1dc29b).

The IBC hooks IBC middleware parses an ICS20 transfer, and if the `memo` field is of a particular form, executes a Wasm contract call.

The following sections detail the `memo` format for Wasm contract calls and the execution guarantees provided.

### CosmWasm Contract Execution Format

Before diving into the IBC metadata format, it's important to understand the CosmWasm execute message format to get a sense of the specific fields that need to be set. Provided below is the CosmWasm `MsgExecuteContract` format as defined in the [Wasm module](https://github.com/CosmWasm/wasmd/blob/4fe2fbc8f322efdaf187e2e5c99ce32fd1df06f0/x/wasm/types/tx.pb.go#L340-L349).

```go
type MsgExecuteContract struct {
	// Sender is the that actor that signed the messages
	Sender string
	// Contract is the address of the smart contract
	Contract string
	// Msg json encoded message to be passed to the contract
	Msg RawContractMessage
	// Funds coins that are transferred to the contract on execution
	Funds sdk.Coins
}
```

For use with IBC hooks, the message fields above can be derived from the following:

- `Sender`: IBC packet senders cannot be explicitly trusted, as they can be deceitful. Chains cannot risk the sender being confused with a particular local user or module address. To prevent this, the `sender` is replaced with an account that represents the sender prefixed by the channel and a Wasm module prefix. This is done by setting the sender to `Bech32(Hash("ibc-wasm-hook-intermediary" || channelID || sender))`, where the `channelId` is the channel id on the local chain.
- `Contract`: This field should be directly obtained from the ICS-20 packet metadata
- `Msg`: This field should be directly obtained from the ICS-20 packet metadata.
- `Funds`: This field is set to the amount of funds being sent over in the ICS-20 packet. The denom in the packet must be specified as the counterparty chain's representation of the denom.

<!-- markdown-link-check-disable-next-line -->
> **_WARNING:_**  Due to a [bug](https://twitter.com/SCVSecurity/status/1682329758020022272) in the packet forward middleware, we cannot trust the sender from chains that use PFM. Until that is fixed, we recommend chains to not trust the sender on contracts executed via IBC hooks.


The fully constructed execute message will look like the following:

```go
msg := MsgExecuteContract{
	// Sender is the that actor that signed the messages
	Sender: "osmo1-hash-of-channel-and-sender",
	// Contract is the address of the smart contract
	Contract: packet.data.memo["wasm"]["ContractAddress"],
	// Msg json encoded message to be passed to the contract
	Msg: packet.data.memo["wasm"]["Msg"],
	// Funds coins that are transferred to the contract on execution
	Funds: sdk.NewCoin{Denom: ibc.ConvertSenderDenomToLocalDenom(packet.data.Denom), Amount: packet.data.Amount}
}
```

### ICS20 packet structure

Given the details above, you can propagate the implied ICS-20 packet data structure.
ICS20 is JSON native, so you can use JSON for the memo format.

```json
{
    //... other ibc fields that we don't care about
    "data":{
    	"denom": "denom on counterparty chain (e.g. uatom)",  // will be transformed to the local denom (ibc/...)
        "amount": "1000",
        "sender": "addr on counterparty chain", // will be transformed
        "receiver": "contract addr or blank",
    	"memo": {
           "wasm": {
              "contract": "osmo1contractAddr",
              "msg": {
                "raw_message_fields": "raw_message_data",
              }
            }
        }
    }
}
```

An ICS-20 packet is formatted correctly for IBC hooks if all of the following are true:

- The `memo` is not blank.
- The`memo` is valid JSON.
- The `memo` has at least one key, with the value `"wasm"`.
- The `memo["wasm"]` has exactly two entries, `"contract"` and `"msg"`.
- The `memo["wasm"]["msg"]` is a valid JSON object.
- The `receiver == "" || receiver == memo["wasm"]["contract"]`.

An ICS-20 packet is directed toward IBC hooks if all of the following are true:

- The `memo` is not blank.
- The `memo` is valid JSON.
- The `memo` has at least one key, with the name `"wasm"`.

If an ICS-20 packet is not directed towards IBC hooks, IBC hooks doesn't do anything.
If an ICS-20 packet is directed towards IBC hooks, and is formatted incorrectly, then IBC hooks returns an error.

### Execution flow

1. Pre-IBC hooks:

- Ensure the incoming IBC packet is cryptographically valid.
- Ensure the incoming IBC packet is not timed out.

2. In IBC hooks, pre-packet execution:

- Ensure the packet is correctly formatted (as defined above).
- Edit the receiver to be the hardcoded IBC module account.

3. In IBC hooks, post packet execution:

- Construct a Wasm message as defined before.
- Execute the Wasm message.
- If the Wasm message has an error, return `ErrAck`.
- Otherwise, continue through middleware.

## Ack callbacks

A contract that sends an IBC transfer may need to listen for the `ack` from that packet. `Ack` callbacks allow
contracts to listen in on the `ack` of specific packets.

### Design

The sender of an IBC transfer packet may specify a callback for when the `Ack` of that packet is received in the memo
field of the transfer packet.

Crucially, **only the IBC packet sender can set the callback**.

### Use case

The cross-chain swaps implementation sends an IBC transfer. If the transfer were to fail, the sender should be able to retrieve their funds which would otherwise be stuck in the contract. To do this, users should be allowed to retrieve the funds after the timeout has passed. However, without the `Ack` information, one cannot guarantee that the send hasn't failed (i.e.: returned an error ack notifying that the receiving chain didn't accept it).

### Implementation

#### Callback information in memo

For the callback to be processed, the transfer packet's `memo` should contain the following in its JSON:

```json
{"ibc_callback": "osmo1contractAddr"}
```

The IBC hooks will keep the mapping from the packet's channel and sequence to the contract in storage. When an `Ack` is
received, it will notify the specified contract via a sudo message.

#### Interface for receiving the Acks and Timeouts

The contract that awaits the callback should implement the following interface for a sudo message:

```rust
#[cw_serde]
pub enum IBCLifecycleComplete {
    #[serde(rename = "ibc_ack")]
    IBCAck {
        /// The source channel (osmosis side) of the IBC packet
        channel: String,
        /// The sequence number that the packet was sent with
        sequence: u64,
        /// String encoded version of the `Ack` as seen by OnAcknowledgementPacket(..)
        ack: String,
        /// Weather an `Ack` is a success of failure according to the transfer spec
        success: bool,
    },
    #[serde(rename = "ibc_timeout")]
    IBCTimeout {
        /// The source channel (osmosis side) of the IBC packet
        channel: String,
        /// The sequence number that the packet was sent with
        sequence: u64,
    },
}

/// Message type for `sudo` entry_point
#[cw_serde]
pub enum SudoMsg {
    #[serde(rename = "ibc_lifecycle_complete")]
    IBCLifecycleComplete(IBCLifecycleComplete),
}
```

## Installation

Follow these steps to install the IBC hooks module. The following lines are all added to `app.go`

1. Import the following packages.

```go
// import (
    ...
    ibchooks "github.com/cosmos/ibc-apps/modules/ibc-hooks/v8"
	ibchookskeeper "github.com/cosmos/ibc-apps/modules/ibc-hooks/v8/keeper"
	ibchookstypes "github.com/cosmos/ibc-apps/modules/ibc-hooks/v8/types"
    ...
// )
```

2. Add the module.

```go
// var (
	//DefaultNodeHome string
	//ModuleBasics = module.NewBasicManager(
        ...
        ibchooks.AppModuleBasic{},
        ...
    // )
```

3. Add the IBC hooks keeper.

```go

// type App struct {
    ...
    IBCHooksKeeper   ibchookskeeper.Keeper
    ...
// }

```
4. Initiate the keepers and helpers.

```go
...

	// 'ibc-hooks' module - depends on
	// 1. 'auth'
	// 2. 'bank'
	// 3. 'distr'
	app.keys[ibchookstypes.StoreKey] = storetypes.NewKVStoreKey(ibchookstypes.StoreKey)
	app.IBCHooksKeeper = ibchookskeeper.NewKeeper(
		app.keys[ibchookstypes.StoreKey],
	)
	app.Ics20WasmHooks = ibchooks.NewWasmHooks(&app.IBCHooksKeeper, nil, AccountAddressPrefix) // The contract keeper needs to be set later

	// initialize the wasm keeper with
	// wasmKeeper := wasm.NewKeeper( ... )
	app.WasmKeeper = &wasmKeeper

	// Pass the contract keeper to all the structs (generally ICS4Wrappers for ibc middlewares) that need it
	app.ContractKeeper = wasmkeeper.NewDefaultPermissionKeeper(app.WasmKeeper)
	app.Ics20WasmHooks.ContractKeeper = app.ContractKeeper
	app.HooksICS4Wrapper = ibchooks.NewICS4Middleware(
		app.IBCKeeper.ChannelKeeper,
		app.Ics20WasmHooks,
	)
	// Hooks Middleware
	transferIBCModule := ibctransfer.NewIBCModule(app.TransferKeeper)
	app.TransferStack = ibchooks.NewIBCMiddleware(&transferIBCModule, &app.HooksICS4Wrapper)

...
```

5. Register the genesis, begin blocker and end block hooks.

```go
...

	app.ModuleManager.SetOrderBeginBlockers(
		// upgrades should be run first
		upgradetypes.ModuleName,
		capabilitytypes.ModuleName,
		minttypes.ModuleName,
		consensustypes.ModuleName,
		distrtypes.ModuleName,
		slashingtypes.ModuleName,
		evidencetypes.ModuleName,
		stakingtypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		govtypes.ModuleName,
		crisistypes.ModuleName,
		ibctransfertypes.ModuleName,
		ibcexported.ModuleName,
		icatypes.ModuleName,
		ibcfeetypes.ModuleName,
		genutiltypes.ModuleName,
		authz.ModuleName,
		feegrant.ModuleName,
		group.ModuleName,
		paramstypes.ModuleName,
		vestingtypes.ModuleName,
		nft.ModuleName,
		ibchookstypes.ModuleName,
		wasm.ModuleName,
		// this line is used by starport scaffolding # stargate/app/beginBlockers
	)

	app.ModuleManager.SetOrderEndBlockers(
		crisistypes.ModuleName,
		govtypes.ModuleName,
		stakingtypes.ModuleName,
		consensustypes.ModuleName,
		ibctransfertypes.ModuleName,
		ibcexported.ModuleName,
		icatypes.ModuleName,
		capabilitytypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		distrtypes.ModuleName,
		slashingtypes.ModuleName,
		minttypes.ModuleName,
		genutiltypes.ModuleName,
		evidencetypes.ModuleName,
		authz.ModuleName,
		feegrant.ModuleName,
		group.ModuleName,
		paramstypes.ModuleName,
		upgradetypes.ModuleName,
		vestingtypes.ModuleName,
		nft.ModuleName,
		ibcfeetypes.ModuleName,
		ibchookstypes.ModuleName,
		wasm.ModuleName,
		// this line is used by starport scaffolding # stargate/app/endBlockers
	)

	// NOTE: The genutils module must occur after staking so that pools are
	// properly initialized with tokens from genesis accounts.
	// NOTE: Capability module must occur first so that it can initialize any capabilities
	// so that other modules that want to create or claim capabilities afterwards in InitChain
	// can do so safely.
	app.ModuleManager.SetOrderInitGenesis(
		capabilitytypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		distrtypes.ModuleName,
		stakingtypes.ModuleName,
		slashingtypes.ModuleName,
		consensustypes.ModuleName,
		govtypes.ModuleName,
		minttypes.ModuleName,
		crisistypes.ModuleName,
		genutiltypes.ModuleName,
		ibctransfertypes.ModuleName,
		ibcexported.ModuleName,
		icatypes.ModuleName,
		evidencetypes.ModuleName,
		authz.ModuleName,
		feegrant.ModuleName,
		group.ModuleName,
		paramstypes.ModuleName,
		upgradetypes.ModuleName,
		vestingtypes.ModuleName,
		nft.ModuleName,
		ibcfeetypes.ModuleName,
		ibchookstypes.ModuleName,
		wasm.ModuleName,
		// this line is used by starport scaffolding # stargate/app/initGenesis
	)

...
```
## Tests


Tests are included in the [tests folder](./tests/unit/testdata/counter/README.md).

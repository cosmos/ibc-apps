# IBC-hooks

> This module is forked from https://github.com/osmosis-labs/osmosis/tree/main/x/ibc-hooks

## Wasm Hooks

Wasm hooks are an IBC middleware that enables ICS-20 token transfers to initiate contract calls.
This functionality allows cross-chain contract calls that involve token movement. 
Wasm hooks are useful for a variety of use cases, including cross-chain swaps, which are an extremely powerful primitive.

## How do Wasm Hooks work?

Wasm hooks are made possible through the `memo` field included in every ICS-20 transfer packet, as introduced in [IBC v3.4.0](https://medium.com/the-interchain-foundation/moving-beyond-simple-token-transfers-d42b2b1dc29b).

The Wasm hooks IBC middleware parses an ICS20 transfer, and if the `memo` field is of a particular form, executes a Wasm contract call. 

The following sections detail the `memo` format for Wasm contract calls and the execution guarantees provided.

### Cosmwasm Contract Execution Format

Before diving into the IBC metadata format, it's important to understand the Cosmwasm execute message format to get a sense of the specific fields that need to be set. Provided below is the CosmWasm `MsgExecuteContract` format as defined in the [Wasm module](https://github.com/CosmWasm/wasmd/blob/4fe2fbc8f322efdaf187e2e5c99ce32fd1df06f0/x/wasm/types/tx.pb.go#L340-L349).

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

For use with Wasm hooks, the message fields above can be derived from the following: 

- `Sender`: IBC packet senders cannot be explicitly trusted, as they can be deceitful. Chains cannot risk the sender being confused with a particular local user or module address. To prevent this, the `sender` is replaced with an account that represents the sender prefixed by the channel and a Wasm module prefix. This is done by setting the sender to `Bech32(Hash("ibc-wasm-hook-intermediary" || channelID || sender))`, where the `channelId` is the channel id on the local chain. 
- `Contract`: This field should be directly obtained from the ICS-20 packet metadata
- `Msg`: This field should be directly obtained from the ICS-20 packet metadata.
- `Funds`: This field is set to the amount of funds being sent over in the ICS-20 packet. The denom in the packet must be specified as the counterparty chain's representation of the denom.


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

An ICS-20 packet is formatted correctly for Wasm hooks if all of the following are true:

- The `memo` is not blank.
- The`memo` is valid JSON. 
- The `memo` has at least one key, with the value `"wasm"`.
- The `memo["wasm"]` has exactly two entries, `"contract"` and `"msg"`. 
- The `memo["wasm"]["msg"]` is a valid JSON object.
- The `receiver == "" || receiver == memo["wasm"]["contract"]`. 

An ICS-20 packet is directed toward Wasm hooks if all of the following are true:

- The `memo` is not blank. 
- The `memo` is valid JSON.
- The `memo` has at least one key, with the name `"wasm"`.

If an ICS-20 packet is not directed towards Wasm hooks, Wasm hooks doesn't do anything.
If an ICS-20 packet is directed towards Wasm hooks, and is formatted incorrectly, then Wasm hooks returns an error.

### Execution flow

1. Pre-Wasm hooks:

- Ensure the incoming IBC packet is cryptographically valid. 
- Ensure the incoming IBC packet is not timed out.

2. In Wasm hooks, pre-packet execution:

- Ensure the packet is correctly formatted (as defined above).
- Edit the receiver to be the hardcoded IBC module account.

3. In wasm hooks, post packet execution:

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

The Wasm hooks will keep the mapping from the packet's channel and sequence to the contract in storage. When an `Ack` is
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

## Tests

Tests are included in the [tests folder](./tests/testdata/counter/README.md). 
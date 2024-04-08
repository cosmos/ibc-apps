# ibc-apps

![IBC-APPS Header](ibc-apps.png)

IBC applications and middleware for Cosmos SDK blockchains

🌌 Why have an ibc-apps repo?
================================

Early IBC work started in the [ibc-go](https://github.com/cosmos/ibc-go) repo. As the repo grew, the need arose to parallelize the work among many teams.

The ibc-apps repo is meant to be an easily discoverable, navigable, central place for modules and middleware.

🌌🌌 Who's it for?
===================

IBC-Apps is for:
- _Core **ibc-go** contributors_; it frees them from having to maintain IBC Apps,

- _Publishers of **ibc apps**_, so their apps can be easily found, and

- _Everyone who uses IBC_ and wants to benefit from the full range of its capabilities.


🌌🌌🌌 What is it?
==================

### What is IBC?

The Inter-Blockchain Communication Protocol (IBC) is a protocol to handle the authentication and transport of data between two blockchains. IBC requires a minimal set of functions, specified in the [Interchain Standards](https://github.com/cosmos/ibc/tree/main/spec/ics-001-ics-standard) (ICS). These specifications do not limit the network topology or consensus algorithm, so IBC can be used with a wide range of blockchains or state machines. The IBC protocol provides a permissionless way for relaying data packets between blockchains, unlike most trusted bridging technologies. The security of IBC reduces to the security of the participating chains.

IBC solves a widespread problem: cross-chain communication. This problem exists on public blockchains when exchanges wish to perform swaps. The problem arises early in the case of application-specific blockchains, where every asset is likely to emerge from its own purpose-built chain. Cross-chain communication is also a challenge in the world of private blockchains, in cases where communication with a public chain or other private chains is desirable.

Cross-chain communication between application-specific blockchains in Cosmos creates the potential for high horizontal scalability with transaction finality. These design features provide convincing solutions to well-known problems that plague other platforms, such as transaction costs, network capacity, and transaction confirmation finality.


### What is an IBC App?

IBC apps can be split into two categories - modules and middleware.

IBC Modules are self-contained applications that enable packets to be sent to and received from other IBC-enabled chains.  IBC application developers do not need to concern themselves with the low-level details of clients, connections, and proof verification.

IBC Middleware are self-contained modules that sit between core IBC and an underlying IBC application.  This allows developers to customize lower-level packet handling.  Multiple middleware modules can be chained together.


🌌🌌🌌🌌 How to Use this repo
=============================

If you'd like to include software in this repo, please see [contributing](CONTRIBUTING.md).

🌌🌌🌌🌌🌌 Bonus Content
=============================

## Maintained Branches

|                          **Branch Name**                         | **IBC-Go** |
|:----------------------------------------------------------------:|:----------:|
|            [main](https://github.com/cosmos/ibc-apps)            |     v8     |
| [release/v7](https://github.com/cosmos/ibc-apps/tree/release/v7) |     v7     |
| [release/v6](https://github.com/cosmos/ibc-apps/tree/release/v6) |     v6     |
| [release/v5](https://github.com/cosmos/ibc-apps/tree/release/v5) |     v5     |
| [release/v4](https://github.com/cosmos/ibc-apps/tree/release/v4) |     v4     |

## List of Apps

| Name | Type | Example | Stakeholders | Description |
| ---- | ---- | ------- | ------------ | ----------- |
| [Async Interchain Query](./modules/async-icq/) | Module | Link | [Strangelove](https://github.com/strangelove-ventures/) | Interchain Queries enable blockchains to query the state of an account on another chain without the need for ICA auth. |
| [IBC Hooks](./modules/ibc-hooks/) | Module | [Link](./modules/ibc-hooks/simapp/app.go) | [Osmosis](https://github.com/osmosis-labs) | The IBC hooks module is an IBC middleware that enables ICS-20 token transfers to initiate contract calls. |
| [Packet Forward Middleware](./middleware/packet-forward-middleware) | Middleware | Link | [Strangelove](https://github.com/strangelove-ventures/) | Middleware for forwarding IBC packets. |

## Ecosystem Apps

Modules and middleware developed by other awesome teams in the ecosystem:

| Name | Type | Stakeholders | Description |
| ---- | ---- | ------------ | ----------- |
| [Interchain KV Queries](https://github.com/ingenuity-build/interchain-queries) | Module | [Ingenuity](https://github.com/ingenuity-build) | An application that enables on chain querying of another IBC enabled chains state without the need for the chain being queried to implement the application. |
| [query](https://github.com/defund-labs/defund/tree/main/x/query) | Module | [Defund Labs](https://github.com/defund-labs) | An application that enables on chain querying of another IBC enabled chains state without the need for the chain being queried to implement the application. Similar to the interchain-queries application in the row above but without callbacks. |
| [NFT Transfer (ICS 721)](https://github.com/bianjieai/nft-transfer) | Module | [Bianjieai](https://github.com/bianjieai) | An application that enables cross chain NFT transfer. |
| <!-- markdown-link-check-disable-line --> [CosmWasm NFT Transfer (ICS 721)](https://github.com/public-awesome/cw-ics721) | WASM Contract | [Public Awesome (Stargaze)](https://github.com/public-awesome), [Ark Protocol](https://x.com/ArkProtocol) | An application that enables cross chain NFT transfer. CosmWasm implementation of the above, written in Rust. |
| [recovery](https://github.com/evmos/evmos/tree/v15.0.0/x/recovery) | Middleware | [Evmos](https://github.com/evmos) | Middleware enabling the recovery of tokens sent to unsupported addresses. |
| [ibc-rate-limit](https://github.com/osmosis-labs/osmosis/tree/main/x/ibc-rate-limit) | Middleware | [Osmosis Labs](https://github.com/osmosis-labs) | Middleware that limits the in or out flow of an asset in a certain time period to minimise the risks of cross chain token transfers. This is implemented as a middleware wrapping ICS20 with the rate limiting logic implemented by cosmwasm contracts. |
| [Interchain Atomic Swap](https://github.com/sideprotocol/mesh-liquidity-wasm/tree/v0.1.0/contracts/ics100) | WASM Contract | [Side Labs](https://github.com/sideprotocol) | An application that facilitates inter-blockchain peer-to-peer asset swaps. |
| [Interchain Liquidity](https://github.com/sideprotocol/ibcswap-wasm/tree/main/contracts/ics101) | WASM Contract | [Side Labs](https://github.com/sideprotocol) | An application that splits the state of a weighted liquidity pool between two chains, enabling inter-blockchain automated asset swaps. |

# packet-forward-middleware
Middleware for forwarding IBC packets.

Asynchronous acknowledgements are utilized for atomic multi-hop packet flows. The acknowledgement will only be written on the chain where the user initiated the packet flow after the forward/multi-hop sequence has completed (success or failure). This means that a user (i.e. an IBC application) only needs to monitor the chain where the initial transfer was sent for the response of the entire process.

## About

The packet-forward-middleware is an IBC middleware module built for Cosmos blockchains utilizing the IBC protocol. A chain which incorporates the 
packet-forward-middleware is able to route incoming IBC packets from a source chain to a destination chain. As the Cosmos SDK/IBC become commonplace in the 
blockchain space more and more zones will come online, these new zones joining are noticing a problem: they need to maintain a large amount of infrastructure 
(archive nodes and relayers for each counterparty chain) to connect with all the chains in the ecosystem, a number that is continuing to increase quickly. Luckly 
this problem has been anticipated and IBC has been architected to accomodate multi-hop transactions. However, a packet forwarding/routing feature was not in the 
initial IBC release. 

## Sequence diagrams

### Multi-hop A->B->C->D success
```
        channel-0 channel-1         channel-2 channel-3        channel-4 channel-5
┌───────┐       ibc        ┌───────┐        ibc       ┌───────┐        ibc       ┌───────┐
│Chain A│◄────────────────►│Chain B│◄────────────────►│Chain C│◄────────────────►│Chain D│
└───────┘                  └───────┘                  └───────┘                  └───────┘
     1. transfer 2. recv_packet  3. forward 4. recv_packet  5. forward 6. recv_packet
         ─────────────────► packet  ─────────────────► packet  ─────────────────►
     9. ack                 forward   8. ack           forward   7. ack  
         ◄───────────────── middleware◄─────────────── middleware◄───────────────
```

### Multi-hop A->B->C->D, C->D `recv_packet` error, refund back to A
```
        channel-0 channel-1         channel-2 channel-3        channel-4 channel-5
┌───────┐       ibc        ┌───────┐        ibc       ┌───────┐        ibc       ┌───────┐
│Chain A│◄────────────────►│Chain B│◄────────────────►│Chain C│◄────────────────►│Chain D│
└───────┘                  └───────┘                  └───────┘                  └───────┘
     1. transfer 2. recv_packet  3. forward 4. recv_packet  5. forward 6. recv_packet ERR
         ─────────────────► packet  ─────────────────► packet  ─────────────────►
         9. ack ERR         forward   8. ack ERR       forward   7. ack ERR
         ◄───────────────── middleware◄─────────────── middleware◄───────────────
```

### Forward A->B->C with 1 retry, max timeouts occurs, refund back to A
```
        channel-0 channel-1         channel-2 channel-3
┌───────┐       ibc        ┌───────┐        ibc       ┌───────┐
│Chain A│◄────────────────►│Chain B│◄────────────────►│Chain C│
└───────┘                  └───────┘                  └───────┘
     1. transfer 2. recv_packet     3. forward
         ─────────────────► packet  ─────────────────►
                            forward   4. timeout
                            middleware◄───────────────
                                    5. forward retry
                                    ─────────────────►
         7. ack ERR                 6. timeout
         ◄─────────────────         ◄─────────────────
```

## Examples

Utilizing the packet `memo` field, instructions can be encoded as JSON for multi-hop sequences.

### Minimal Example - Chain forward A->B->C

- The packet-forward-middleware integrated on Chain B.
- The packet data `receiver` for the `MsgTransfer` on Chain A is set to `"pfm"` or some other invalid bech32 string.*
- The packet `memo` is included in `MsgTransfer` by user on Chain A.

memo:
```
{
  "forward": {
    "receiver": "chain-c-bech32-address",
    "port": "transfer",
    "channel": "channel-123"
  }
}
```

### Full Example - Chain forward A->B->C->D with retry on timeout

- The packet-forward-middleware integrated on Chain B and Chain C.
- The packet data `receiver` for the `MsgTransfer` on Chain A is set to `"pfm"` or some other invalid bech32 string.*
- The forward metadata `receiver` for the hop from Chain B to Chain C is set to `"pfm"` or some other invalid bech32 string.*
- The packet `memo` is included in `MsgTransfer` by user on Chain A.
- A packet timeout of 10 minutes and 2 retries is set for both forwards. 

In the case of a timeout after 10 minutes for either forward, the packet would be retried up to 2 times, at which case an error ack would be written to issue a refund on the prior chain.

`next` is the `memo` to pass for the next transfer hop. Per `memo` intended usage of a JSON string, it should be either JSON which will be Marshaled retaining key order, or an escaped JSON string which will be passed directly.

`next` as JSON
```
{
  "forward": {
    "receiver": "pfm", // purposely using invalid bech32 here*
    "port": "transfer",
    "channel": "channel-123",
    "timeout": "10m",
    "retries": 2,
    "next": {
      "forward": {
        "receiver": "chain-d-bech32-address",
        "port": "transfer",
        "channel":"channel-234",
        "timeout":"10m",
        "retries": 2
      }
    }
  }
}
```

`next` as escaped JSON string
```
{
  "forward": {
    "receiver": "pfm", // purposely using invalid bech32 here*
    "port": "transfer",
    "channel": "channel-123",
    "timeout": "10m",
    "retries": 2,
    "next": "{\"forward\":{\"receiver\":\"chain-d-bech32-address\",\"port\":\"transfer\",\"channel\":\"channel-234\",\"timeout\":\"10m\",\"retries\":2}}"
  }
}
```

## Intermediate Receivers*

PFM does not need the packet data `receiver` address to be valid, as it will create a hash of the sender and channel to derive a receiver address on the intermediate chains. This is done for security purposes to ensure that users cannot move funds through arbitrary accounts on intermediate chains.

To prevent accidentally sending funds to a chain which does not have PFM, it is recommended to use an invalid bech32 string (such as `"pfm"`) for the `receiver` on intermediate chains.  By using an invalid bech32 string, a transfer that is accidentally sent to a chain that does not have PFM would fail to be received, and properly refunded to the user on the source chain, rather than having funds get stuck on the intermediate chain.

The examples above show the intended usage of the `receiver` field for one or multiple intermediate PFM chains.

## References

- https://www.mintscan.io/cosmos/proposals/56
- https://github.com/cosmos/ibc-go/pull/373
- https://github.com/strangelove-ventures/governance/blob/master/proposals/2021-09-hub-ibc-router/README.md

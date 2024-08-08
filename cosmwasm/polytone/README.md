# Polytone

Polytone is a protocol that gives every smart contract an account on every IBC-connected blockchain. Polytone has a CosmWasm implementation so CosmWasm chains can deploy Polytone Interchain Accounts and Queries today!

More detail information on Polytone can be found in the [wiki](https://github.com/cosmos/ibc-apps/wiki):
- [How Polytone Works](https://github.com/cosmos/ibc-apps/wiki/How-Polytone-Works)
- [How to use Polytone](https://github.com/cosmos/ibc-apps/wiki/How-to-use-Polytone)
- [How Polytone Handles Channel Closure](https://github.com/cosmos/ibc-apps/wiki/How-Polytone-Handles-Channel-Closure)

## Overview

Polytone is made of three modules: note, voice, and proxy.

![image](https://user-images.githubusercontent.com/30676292/232922218-9348e5cc-3ffc-443e-bcd1-da3cc06755d1.png)

The note says what to say, and the voice (via the sender’s proxy) says it.

![image](https://user-images.githubusercontent.com/30676292/232922253-95444eb3-0c89-4f83-889b-08744febc83a.png)

### Connecting Note & Voice

Different blockchains have different encodings and message types, in Polytone we call these “extensions”.  For a note to connect to a voice, the voice must support all extensions the note does; The voice must be able to say everything the note can speak.

![image](https://user-images.githubusercontent.com/30676292/232922296-dbc7fe87-e50f-4090-afbf-45ffb481b859.png)

Once a note has connected to a voice, it is “paired”. Once paired, the note will only ever connect with that voice, even if the first channel to connect them closes. Pairing simplifies Polytone’s API, as message senders don’t need to specify the channel to send on.

![image](https://user-images.githubusercontent.com/30676292/232922352-63e1ea54-d9fb-41c5-a707-441c61ecfb7e.png)

More information on connections and using Polytone as a developer is available in the [wiki](https://github.com/cosmos/ibc-apps/wiki/How-to-use-Polytone).

### Executing Messages

To execute messages, the messages are sent to the note, which relays them to its voice, which relays them to the sender’s proxy. If the sender has no proxy, a new one is created before relaying the messages.

![image](https://user-images.githubusercontent.com/30676292/232922379-59565f93-1765-426d-8d7c-790b4216cc46.png)

If one of the executed messages fails, all of the messages are rolled back.

![image](https://user-images.githubusercontent.com/30676292/232922425-92028a48-f6f0-40bd-a83d-9aeb631caca0.png)

Executing queries has the same semantics as executing messages. If a single query fails, all queries are canceled.

## Audit

[Polytone has been audited by Oak Security](https://github.com/oak-security/audit-reports/blob/master/Polytone/2023-06-05%20Audit%20Report%20-%20Polytone%20v1.0.pdf).

## Acknowledgements

Thanks to Shane, humanalgorithm, and the Stargaze team for encouraging this work and helping ideate on the idea of an outpost specific interchain accounts. Thanks to Belsy, Ethan Frey, and larry0x for the design feedback and great technical discussion in our CW-ICA Telegram chat. Thank you to Jake and Noah who helped in the design, ideation, and relaying for Polytone. Thank you to art3mix and benskey who worked on the implementation. And finally, thank you to the Juno Community DAO and Stargaze for funding this work!

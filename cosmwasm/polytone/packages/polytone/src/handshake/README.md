## The Polytone Handshake

## Background

Polytone is a protocol that gives every smart contract an account on
every IBC blockchain. It has two modules, note and voice. Note
modules receive instructions on local chains, and tell voice
modules on remote chains to execute them.

There are two things to be negotiated during the handshake.

1. What is the message encoding? For example: JSON or Protobuf.
2. What type of messages are being sent? For example: Cosmos SDK with
   the Osmosis extension.

## Requirements

We'd like the following to be true:

1. A note should only connect to a voice of the same polytone version.
2. A note should only connect to a voice who's supported extensions
   and encodings are a superset of their own.
3. A channel can be opened from either the note or voice side of the
   connection.

The first requrement ensures that a note never connects to a different
IBC protocol or another note module. The second, that every message
sent could be executed. The third is good relayer UX.

## Design

There are four steps to the IBC handshake, init, try, ack, and
confirm. These alternate between the two chains, for example, for a
connection between A and B, init would be called on A, try on B, ack
on A, and confirm on B.

During a handshake, the modules involved can pass around a version
datastructure. For the init phase, this is set by the initializer of
the handshake, afterwards it is set by the modules.

### Init

```
def init(version):
    if version == "polytone":
        if note:
            return "polytone-note-1"
        if voice:
            return "polytone-voice-1"
    return error
```

During initialization, it is checked that the initiator means to
create a Polytone channel (satisfying requirement (1)), and the
version of the module is sent to the counterparty.

### Try

First, we defined `supported` as the set of encodings and extensions
supported by a module. For example, a CosmWasm voice module for
Osmosis might have:

```
supported = ["JSON-CosmosMsg", "JSON-OsmosisExtension"]
```

The logic for try then verifies that the counterparty module satisfies
requirement (2) and returns the supported set.

```
def try(version):
    if note && version == "polytone-voice-1" ||
       voice && version == "polytone-note-1":
        return supported
    return error
```

### ACK

```
def ack(version):
    if note && supported.subseteq(version) ||
       voice && version.subseteq(supported):
        return selection(version)
```

The `selection` function selects from any mutually exclusive
options. For example, for a version like:

```
["proto3-CosmosMsg", "JSON-CosmosMsg"]
```

A selection method that preferred JSON would return:

```
["JSON-CosmosMsg"]
```

## Confirm

```
def confirm(version):
    return version
```

Once the selection is done, there is nothing to do but accept it on
the confirm step. Implementors will likely add additional logic here
for checking the final version, and configuring the state machine for
the selected version.

use cosmwasm_schema::{cw_serde, QueryResponses};
use cosmwasm_std::{CosmosMsg, Empty, QueryRequest, Uint64};

use polytone::callbacks::CallbackRequest;

#[cw_serde]
pub struct InstantiateMsg {
    /// This contract pairs with the first voice module that a relayer
    /// connects it with, or the pair specified here. Once it has a
    /// pair, it will never handshake with a different voice module,
    /// even after channel closure. This ensures that there will only
    /// ever be one voice for every note.
    pub pair: Option<Pair>,

    /// The max gas allowed in a transaction. When returning callbacks
    /// the module will use this to calculate the amount of gas to
    /// save for handling a callback error. This protects from
    /// callbacks that run out of gas preventing ACKs or timeouts from
    /// being returned.
    ///
    /// The contract admin can update with `MigrateMsg::WithUpdate`.
    pub block_max_gas: Uint64,
}

#[cw_serde]
#[cfg_attr(feature = "interface", derive(cw_orch::ExecuteFns))] // cw-orch automatic
pub enum ExecuteMsg {
    /// Performs the requested queries on the voice chain and returns
    /// a callback of Vec<QuerierResult>, or ACK-FAIL if unmarshalling
    /// any of the query requests fails.
    #[cfg_attr(feature = "interface", fn_name("ibc_query"))]
    Query {
        msgs: Vec<QueryRequest<Empty>>,
        callback: CallbackRequest,
        timeout_seconds: Uint64,
    },
    /// Executes the requested messages on the voice chain on behalf
    /// of the note chain sender. Message receivers can return data in
    /// their callbacks by calling `set_data` on their `Response`
    /// object. Optionally, returns a callback of `Vec<Callback>` where
    /// index `i` corresponds to the callback for `msgs[i]`.
    ///
    /// Accounts are created on the voice chain after the first call
    /// to execute by the local address. To create an account, but
    /// perform no additional actions, pass an empty list to
    /// `msgs`. Accounts are queryable via the `RemoteAddress {
    /// local_address }` query after they have been created.
    #[cfg_attr(feature = "interface", fn_name("ibc_execute"))]
    Execute {
        msgs: Vec<CosmosMsg<Empty>>,
        callback: Option<CallbackRequest>,
        timeout_seconds: Uint64,
    },
}

#[cw_serde]
#[derive(QueryResponses)]
#[cfg_attr(feature = "interface", derive(cw_orch::QueryFns))] // cw-orch automatic
pub enum QueryMsg {
    /// This channel this note is currently connected to, or none if
    /// no channel is connected.
    #[returns(Option<String>)]
    ActiveChannel,
    /// The contract's corresponding voice on a remote chain.
    #[returns(Option<Pair>)]
    Pair,
    /// Returns the remote address for the provided local address. If
    /// no account exists, returns `None`. An account can be created
    /// by calling `ExecuteMsg::Execute` with the sender being
    /// `local_address`.
    #[returns(Option<String>)]
    RemoteAddress { local_address: String },
    /// Currently set gas limit
    #[returns(Uint64)]
    BlockMaxGas,
}

/// This contract's voice. There is one voice per note, and many notes
/// per voice.
#[cw_serde]
pub struct Pair {
    pub connection_id: String,
    pub remote_port: String,
}

#[cw_serde]
pub enum MigrateMsg {
    /// Updates the contract's configuration. To update the config
    /// without updating the code, migrate to the same code ID.
    WithUpdate { block_max_gas: Uint64 },
}

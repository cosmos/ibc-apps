use cosmwasm_schema::cw_serde;
use cosmwasm_std::{CosmosMsg, Empty, QueryRequest};

pub const VERSION: &str = "polytone";

#[cw_serde]
pub struct Packet {
    /// Message sender on the note chain.
    pub sender: String,
    /// Message to execute on voice chain.
    pub msg: Msg,
}

#[cw_serde]
pub enum Msg {
    /// Performs the requested queries on the voice chain and returns a
    /// callback of Vec<QuerierResult>, or ACK-FAIL if unmarshalling
    /// any of the query requests fails.
    Query { msgs: Vec<QueryRequest<Empty>> },
    /// Executes the requested messages on the voice chain on behalf of
    /// the note chain sender. Message receivers can return data
    /// in their callbacks by calling `set_data` on their `Response`
    /// object. Returns a callback of `Vec<Callback>` where index `i`
    /// corresponds to the callback for `msgs[i]`.
    Execute { msgs: Vec<CosmosMsg<Empty>> },
}

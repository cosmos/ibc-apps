use cosmwasm_std::Addr;
use cw_storage_plus::{Item, Map};
use polytone::callbacks::CallbackMessage;

/// The note that can call this contract.
pub(crate) const NOTE: Item<Addr> = Item::new("note");

/// (initiator, initiator_msg) -> callback
pub(crate) const RESULTS: Map<(String, String), CallbackMessage> = Map::new("results");

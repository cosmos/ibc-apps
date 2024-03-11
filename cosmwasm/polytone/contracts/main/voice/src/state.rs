use cosmwasm_std::Addr;
use cw_storage_plus::{Item, Map};

/// (connection_id, remote_port, remote_sender) -> proxy
pub(crate) const SENDER_TO_PROXY: Map<(String, String, String), Addr> = Map::new("c2p");

/// (channel_id) -> connection_id
pub(crate) const CHANNEL_TO_CONNECTION: Map<String, String> = Map::new("c2c");

/// Code ID of the proxy contract being used.
pub(crate) const PROXY_CODE_ID: Item<u64> = Item::new("pci");

/// Max gas usable in a single block.
pub(crate) const BLOCK_MAX_GAS: Item<u64> = Item::new("bmg");

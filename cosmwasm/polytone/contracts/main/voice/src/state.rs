use cosmwasm_schema::cw_serde;
use cosmwasm_std::Addr;
use cw_storage_plus::{Item, Map};

/// (connection_id, remote_port, remote_sender) -> proxy
pub(crate) const SENDER_TO_PROXY: Map<(String, String, String), Addr> = Map::new("c2p");

/// proxy -> { connection_id, remote_port, remote_sender }
pub(crate) const PROXY_TO_SENDER: Map<Addr, SenderInfo> = Map::new("p2c");

/// (channel_id) -> connection_id
pub(crate) const CHANNEL_TO_CONNECTION: Map<String, String> = Map::new("c2c");

/// Code ID of the proxy contract being used.
pub(crate) const PROXY_CODE_ID: Item<u64> = Item::new("pci");

/// Max gas usable in a single block.
pub(crate) const BLOCK_MAX_GAS: Item<u64> = Item::new("bmg");

/// Contract address length used by the chain.
pub(crate) const CONTRACT_ADDR_LEN: Item<u8> = Item::new("cal");

#[cw_serde]
pub struct SenderInfo {
    pub connection_id: String,
    pub remote_port: String,
    pub remote_sender: String,
}
use cosmwasm_std::Storage;
use cw_storage_plus::{Item, Map};

use crate::error::ContractError;

/// (Connection-ID, Remote port) of this contract's pair.
pub const CONNECTION_REMOTE_PORT: Item<(String, String)> = Item::new("a");

/// Channel-ID of the channel currently connected. Holds no value when
/// no channel is active.
pub const CHANNEL: Item<String> = Item::new("b");

/// Max gas usable in a single block.
pub const BLOCK_MAX_GAS: Item<u64> = Item::new("bmg");

/// (channel_id) -> sequence number. `u64` is the type used in the
/// Cosmos SDK for sequence numbers:
///
/// <https://github.com/cosmos/ibc-go/blob/a25f0d421c32b3a2b7e8168c9f030849797ff2e8/modules/core/02-client/keeper/keeper.go#L116-L125>
const SEQUENCE_NUMBER: Map<String, u64> = Map::new("sn");

/// Increments and returns the next sequence number.
pub(crate) fn increment_sequence_number(
    storage: &mut dyn Storage,
    channel_id: String,
) -> Result<u64, ContractError> {
    let seq = SEQUENCE_NUMBER
        .may_load(storage, channel_id.clone())?
        .unwrap_or_default()
        .checked_add(1)
        .ok_or(ContractError::SequenceOverflow)?;
    SEQUENCE_NUMBER.save(storage, channel_id, &seq)?;
    Ok(seq)
}

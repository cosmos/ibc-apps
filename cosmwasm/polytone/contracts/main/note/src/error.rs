use cosmwasm_std::StdError;
use thiserror::Error;

#[derive(Error, Debug, PartialEq)]
pub enum ContractError {
    #[error(transparent)]
    Std(#[from] StdError),

    #[error(transparent)]
    Handshake(#[from] polytone::handshake::error::HandshakeError),

    #[error("contract is already paired with port ({pair_port}) on connection ({pair_connection}), got port ({suggested_port}) on connection ({suggested_connection})")]
    AlreadyPaired {
        suggested_connection: String,
        suggested_port: String,
        pair_connection: String,
        pair_port: String,
    },

    #[error("contract has no pair, establish a channel with a voice module to create one")]
    NoPair,

    #[error("ERR_GAS_NEEDED can't be higher then BLOCK_MAX_GAS")]
    GasLimitsMismatch,

    #[error("channel sequence number overflow, to fix: the contract admin may migrate to close and reopen the channel")]
    SequenceOverflow,
}

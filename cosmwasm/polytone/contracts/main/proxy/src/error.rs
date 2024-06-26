use cosmwasm_std::StdError;
use cw_utils::ParseReplyError;
use thiserror::Error;

#[derive(Error, Debug)]
pub enum ContractError {
    #[error(transparent)]
    Std(#[from] StdError),

    #[error(transparent)]
    Parse(#[from] ParseReplyError),

    #[error("caller must be the contract instantiator")]
    NotInstantiator,

    #[error("executing message {index}: {error}")]
    MsgError { index: u64, error: String },
}

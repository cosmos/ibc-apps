use cosmwasm_schema::{cw_serde, QueryResponses};
use polytone::callbacks::CallbackMessage;

#[cw_serde]
pub struct InstantiateMsg {
    /// The polytone note contract that can call this contract.
    pub note: String,
}

#[cw_serde]
pub enum ExecuteMsg {
    /// Stores the callback in state and makes it queryable.
    Callback(CallbackMessage),
}

#[cw_serde]
#[derive(QueryResponses)]
pub enum QueryMsg {
    /// Gets note that can call this contract.
    #[returns(String)]
    Note {},
    /// Gets callback result.
    #[returns(ResultResponse)]
    Result {
        initiator: String,
        initiator_msg: String,
    },
}

#[cw_serde]
pub struct ResultResponse {
    pub callback: CallbackMessage,
}

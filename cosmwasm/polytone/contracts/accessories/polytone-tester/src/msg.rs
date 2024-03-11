use cosmwasm_schema::{cw_serde, QueryResponses};
use cosmwasm_std::Binary;

#[cw_serde]
pub struct InstantiateMsg {}

#[cw_serde]
pub enum ExecuteMsg {
    /// Calls `set_data(data)` if `data` is not None.
    Hello { data: Option<Binary> },
    /// Stores the callback in state and makes it queryable
    Callback(polytone::callbacks::CallbackMessage),
    /// Runs out of gas.
    RunOutOfGas {},
}

#[cw_serde]
#[derive(QueryResponses)]
pub enum QueryMsg {
    /// Gets callback history.
    #[returns(CallbackHistoryResponse)]
    History {},
    /// Gets the history of addresses' that have called the `hello {
    /// data }` method.
    #[returns(HelloHistoryResponse)]
    HelloHistory {},
}

#[cw_serde]
pub struct CallbackHistoryResponse {
    pub history: Vec<polytone::callbacks::CallbackMessage>,
}

#[cw_serde]
pub struct HelloHistoryResponse {
    /// History of callers of the `hello { data }` method.
    pub history: Vec<String>,
}

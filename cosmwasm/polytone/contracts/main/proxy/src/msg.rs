use cosmwasm_schema::{cw_serde, QueryResponses};
use cosmwasm_std::CosmosMsg;

#[cw_serde]
pub struct InstantiateMsg {}

#[cw_serde]
#[cfg_attr(feature = "interface", derive(cw_orch::ExecuteFns))] // cw-orch automatic
pub enum ExecuteMsg {
    Proxy { msgs: Vec<CosmosMsg> },
}

#[cw_serde]
#[derive(QueryResponses)]
#[cfg_attr(feature = "interface", derive(cw_orch::QueryFns))] // cw-orch automatic
pub enum QueryMsg {
    #[returns(cosmwasm_std::Addr)]
    Instantiator {},
}

use cosmwasm_schema::{cw_serde, QueryResponses};
use cosmwasm_std::{Binary, Uint64};

use crate::state::SenderInfo;

#[cw_serde]
pub struct InstantiateMsg {
    /// Code ID to use for instantiating proxy contracts.
    pub proxy_code_id: Uint64,
    /// The max gas allowed in a single block.
    pub block_max_gas: Uint64,
    /// The contract address length used by the chain. Defaults to 32. Some
    /// chains use other lengths, such as Injective which uses 20.
    pub contract_addr_len: Option<u8>,
}

#[cw_serde]
#[cfg_attr(feature = "interface", derive(cw_orch::ExecuteFns))] // cw-orch automatic
pub enum ExecuteMsg {
    /// Receives and handles an incoming packet.
    Rx {
        /// The local connection id the packet arrived on.
        connection_id: String,
        /// The port of the counterparty module.
        counterparty_port: String,
        /// The packet data.
        data: Binary,
    },
}

#[cw_serde]
#[derive(QueryResponses)]
#[cfg_attr(feature = "interface", derive(cw_orch::QueryFns))] // cw-orch automatic
pub enum QueryMsg {
    /// Queries the configured block max gas. Serialized as
    /// `"block_max_gas"`.
    #[returns(Uint64)]
    BlockMaxGas,
    /// Queries the configured proxy code ID. Serialized as
    /// `"proxy_code_id"`.
    #[returns(Uint64)]
    ProxyCodeId,
    /// Queries the configured contract address length.
    #[returns(u8)]
    ContractAddrLen,
    /// Queries the sender information for a given proxy.
    #[returns(SenderInfo)]
    SenderInfoForProxy { proxy: String },
}

#[cw_serde]
pub enum MigrateMsg {
    /// Updates the module's configuration.
    WithUpdate {
        /// Code ID to use for instantiating proxy contracts.
        proxy_code_id: Uint64,
        /// The max gas allowed in a single block.
        block_max_gas: Uint64,
        /// The contract address length used by the chain.
        contract_addr_len: u8,
    },
}

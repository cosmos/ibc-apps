#[cfg(not(feature = "library"))]
use cosmwasm_std::entry_point;
use cosmwasm_std::{
    from_binary, instantiate2_address, to_binary, to_vec, Binary, CodeInfoResponse, ContractResult,
    Deps, DepsMut, Env, MessageInfo, Response, StdResult, SubMsg, SystemResult, Uint64, WasmMsg,
};
use cw2::set_contract_version;

use polytone::ack::{ack_query_fail, ack_query_success};
use polytone::ibc::{Msg, Packet};

use crate::error::ContractError;
use crate::ibc::{ACK_GAS_NEEDED, REPLY_FORWARD_DATA};
use crate::msg::{ExecuteMsg, InstantiateMsg, MigrateMsg, QueryMsg};
use crate::state::{BLOCK_MAX_GAS, PROXY_CODE_ID, SENDER_TO_PROXY};

const CONTRACT_NAME: &str = "crates.io:polytone-voice";
const CONTRACT_VERSION: &str = env!("CARGO_PKG_VERSION");

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn instantiate(
    deps: DepsMut,
    _env: Env,
    _info: MessageInfo,
    msg: InstantiateMsg,
) -> Result<Response, ContractError> {
    set_contract_version(deps.storage, CONTRACT_NAME, CONTRACT_VERSION)?;

    if msg.proxy_code_id.is_zero() {
        return Err(ContractError::CodeIdCantBeZero);
    }

    if msg.block_max_gas.u64() <= ACK_GAS_NEEDED {
        return Err(ContractError::GasLimitsMismatch);
    }

    PROXY_CODE_ID.save(deps.storage, &msg.proxy_code_id.u64())?;
    BLOCK_MAX_GAS.save(deps.storage, &msg.block_max_gas.u64())?;

    Ok(Response::default()
        .add_attribute("method", "instantiate")
        .add_attribute("proxy_code_id", msg.proxy_code_id)
        .add_attribute("block_max_gas", msg.block_max_gas))
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn execute(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    msg: ExecuteMsg,
) -> Result<Response, ContractError> {
    match msg {
        ExecuteMsg::Rx {
            connection_id,
            counterparty_port,
            data,
        } => {
            if info.sender != env.contract.address {
                Err(ContractError::NotSelf)
            } else {
                let Packet { sender, msg } = from_binary(&data)?;
                match msg {
                    Msg::Query { msgs } => {
                        let mut results = Vec::with_capacity(msgs.len());
                        for msg in msgs {
                            let query_result = deps.querier.raw_query(&to_vec(&msg)?);
                            let error = match query_result {
                                SystemResult::Ok(ContractResult::Err(error)) => {
                                    format!("contract: {error}")
                                }
                                SystemResult::Err(error) => format!("system: {error}"),
                                SystemResult::Ok(ContractResult::Ok(res)) => {
                                    results.push(res);
                                    continue;
                                }
                            };
                            return Ok(Response::default()
                                .add_attribute("method", "rx_query_fail")
                                .add_attribute("query_index", results.len().to_string())
                                .add_attribute("query_error", error.as_str())
                                .set_data(ack_query_fail(
                                    Uint64::new(results.len() as u64),
                                    error,
                                )));
                        }
                        Ok(Response::default()
                            .add_attribute("method", "rx_query_success")
                            .add_attribute("queries_executed", results.len().to_string())
                            .set_data(ack_query_success(results)))
                    }
                    Msg::Execute { msgs } => {
                        let (instantiate, proxy) = if let Some(proxy) = SENDER_TO_PROXY.may_load(
                            deps.storage,
                            (
                                connection_id.clone(),
                                counterparty_port.clone(),
                                sender.clone(),
                            ),
                        )? {
                            (None, proxy)
                        } else {
                            let contract =
                                deps.api.addr_canonicalize(env.contract.address.as_str())?;
                            let code_id = PROXY_CODE_ID.load(deps.storage)?;
                            let CodeInfoResponse { checksum, .. } =
                                deps.querier.query_wasm_code_info(code_id)?;
                            let salt = salt(&connection_id, &counterparty_port, &sender);
                            let proxy = deps.api.addr_humanize(&instantiate2_address(
                                &checksum, &contract, &salt,
                            )?)?;
                            SENDER_TO_PROXY.save(
                                deps.storage,
                                (connection_id, counterparty_port, sender.clone()),
                                &proxy,
                            )?;
                            (
                                Some(WasmMsg::Instantiate2 {
                                    admin: None,
                                    code_id,
                                    label: format!("polytone-proxy {sender}"),
                                    msg: to_binary(&polytone_proxy::msg::InstantiateMsg {})?,
                                    funds: vec![],
                                    salt,
                                }),
                                proxy,
                            )
                        };
                        Ok(Response::default()
                            .add_attribute("method", "rx_execute")
                            .add_messages(instantiate)
                            .add_submessage(SubMsg::reply_always(
                                WasmMsg::Execute {
                                    contract_addr: proxy.into_string(),
                                    msg: to_binary(&polytone_proxy::msg::ExecuteMsg::Proxy {
                                        msgs,
                                    })?,
                                    funds: vec![],
                                },
                                REPLY_FORWARD_DATA,
                            )))
                    }
                }
            }
        }
    }
}

/// Generates the salt used to generate an address for a user's
/// account.
///
/// `local_channel` is not attacker controlled and protects from
/// collision from an attacker generated duplicate
/// chain. `remote_port` ensures that two different modules on the
/// same chain produce different addresses for the same
/// `remote_sender`.
fn salt(local_connection: &str, counterparty_port: &str, remote_sender: &str) -> Binary {
    use sha2::{Digest, Sha512};
    // the salt can be a max of 64 bytes (512 bits).
    let hash = Sha512::default()
        .chain_update(local_connection.as_bytes())
        .chain_update(counterparty_port.as_bytes())
        .chain_update(remote_sender.as_bytes())
        .finalize();
    Binary::from(hash.as_slice())
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn query(deps: Deps, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {
        QueryMsg::BlockMaxGas => to_binary(&BLOCK_MAX_GAS.load(deps.storage)?),
        QueryMsg::ProxyCodeId => to_binary(&PROXY_CODE_ID.load(deps.storage)?),
    }
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn migrate(deps: DepsMut, _env: Env, msg: MigrateMsg) -> Result<Response, ContractError> {
    match msg {
        MigrateMsg::WithUpdate {
            proxy_code_id,
            block_max_gas,
        } => {
            if proxy_code_id.is_zero() {
                return Err(ContractError::CodeIdCantBeZero);
            }

            if block_max_gas.u64() <= ACK_GAS_NEEDED {
                return Err(ContractError::GasLimitsMismatch);
            }

            // update the proxy code ID and block max gas
            PROXY_CODE_ID.save(deps.storage, &proxy_code_id.u64())?;
            BLOCK_MAX_GAS.save(deps.storage, &block_max_gas.u64())?;

            Ok(Response::default()
                .add_attribute("method", "migrate_with_update")
                .add_attribute("proxy_code_id", proxy_code_id)
                .add_attribute("block_max_gas", block_max_gas))
        }
    }
}

#[cfg(test)]
mod tests {
    use cosmwasm_std::{instantiate2_address, CanonicalAddr, HexBinary};

    use super::salt;

    fn gen_address(
        local_connection: &str,
        counterparty_port: &str,
        remote_sender: &str,
    ) -> CanonicalAddr {
        let checksum =
            HexBinary::from_hex("13a1fc994cc6d1c81b746ee0c0ff6f90043875e0bf1d9be6b7d779fc978dc2a5")
                .unwrap();
        let creator = CanonicalAddr::from((0..90).map(|_| 9).collect::<Vec<u8>>().as_slice());

        let salt = salt(local_connection, counterparty_port, remote_sender);
        assert!(salt.len() <= 64);
        instantiate2_address(checksum.as_slice(), &creator, &salt).unwrap()
    }

    /// Addresses can be generated, and changing inputs changes
    /// output.
    #[test]
    fn test_address_generation() {
        let one = gen_address("c1", "c1", "c1");
        let two = gen_address("c2", "c1", "c1");
        let three = gen_address("c1", "c2", "c1");
        let four = gen_address("c1", "c1", "c2");
        assert!(one != two && two != three && three != four)
    }
}

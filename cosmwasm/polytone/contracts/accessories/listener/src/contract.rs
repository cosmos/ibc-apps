#[cfg(not(feature = "library"))]
use cosmwasm_std::entry_point;
use cosmwasm_std::{to_json_binary, Binary, Deps, DepsMut, Env, MessageInfo, Response, StdResult};
use cw2::set_contract_version;

use crate::error::ContractError;
use crate::msg::{ExecuteMsg, InstantiateMsg, QueryMsg, ResultResponse};
use crate::state::{NOTE, RESULTS};

const CONTRACT_NAME: &str = "crates.io:polytone-listener";
const CONTRACT_VERSION: &str = env!("CARGO_PKG_VERSION");

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn instantiate(
    deps: DepsMut,
    _env: Env,
    _info: MessageInfo,
    msg: InstantiateMsg,
) -> Result<Response, ContractError> {
    set_contract_version(deps.storage, CONTRACT_NAME, CONTRACT_VERSION)?;

    let note = deps.api.addr_validate(&msg.note)?;
    NOTE.save(deps.storage, &note)?;

    Ok(Response::default()
        .add_attribute("method", "instantiate")
        .add_attribute("note", msg.note))
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn execute(
    deps: DepsMut,
    _env: Env,
    info: MessageInfo,
    msg: ExecuteMsg,
) -> Result<Response, ContractError> {
    match msg {
        ExecuteMsg::Callback(callback) => {
            // Only the note can execute the callback on this contract.
            if info.sender != NOTE.load(deps.storage)? {
                return Err(ContractError::Unauthorized {});
            }

            RESULTS.save(
                deps.storage,
                (
                    callback.initiator.to_string(),
                    callback.initiator_msg.to_string(),
                ),
                &callback,
            )?;
            Ok(Response::default()
                .add_attribute("method", "callback")
                .add_attribute("initiator", callback.initiator.to_string())
                .add_attribute("initiator_msg", callback.initiator_msg.to_string()))
        }
    }
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn query(deps: Deps, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {
        QueryMsg::Note {} => to_json_binary(&NOTE.load(deps.storage)?),
        QueryMsg::Result {
            initiator,
            initiator_msg,
        } => to_json_binary(&ResultResponse {
            callback: RESULTS.load(deps.storage, (initiator, initiator_msg))?,
        }),
    }
}

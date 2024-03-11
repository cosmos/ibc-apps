#[cfg(not(feature = "library"))]
use cosmwasm_std::entry_point;
use cosmwasm_std::{
    to_binary, Binary, Deps, DepsMut, Env, IbcMsg, IbcTimeout, MessageInfo, Response, StdResult,
};
use cw2::set_contract_version;
use polytone::callbacks::CallbackRequestType;
use polytone::{accounts, callbacks, ibc};

use crate::error::ContractError;

use crate::ibc::ERR_GAS_NEEDED;
use crate::msg::{ExecuteMsg, InstantiateMsg, MigrateMsg, Pair, QueryMsg};
use crate::state::{increment_sequence_number, BLOCK_MAX_GAS, CHANNEL, CONNECTION_REMOTE_PORT};

const CONTRACT_NAME: &str = "crates.io:polytone-note";
const CONTRACT_VERSION: &str = env!("CARGO_PKG_VERSION");

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn instantiate(
    deps: DepsMut,
    _env: Env,
    _info: MessageInfo,
    msg: InstantiateMsg,
) -> Result<Response, ContractError> {
    set_contract_version(deps.storage, CONTRACT_NAME, CONTRACT_VERSION)?;

    if msg.block_max_gas.u64() <= ERR_GAS_NEEDED {
        return Err(ContractError::GasLimitsMismatch);
    }

    BLOCK_MAX_GAS.save(deps.storage, &msg.block_max_gas.u64())?;

    let mut response = Response::default()
        .add_attribute("method", "instantiate")
        .add_attribute("block_max_gas", msg.block_max_gas);

    if let Some(Pair {
        connection_id,
        remote_port,
    }) = msg.pair
    {
        response = response
            .add_attribute("pair_connection", connection_id.to_string())
            .add_attribute("pair_port", remote_port.to_string());
        CONNECTION_REMOTE_PORT.save(deps.storage, &(connection_id, remote_port))?;
    };

    Ok(response)
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn execute(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    msg: ExecuteMsg,
) -> Result<Response, ContractError> {
    let (msg, callback, timeout_seconds, request_type) = match msg {
        ExecuteMsg::Execute {
            msgs,
            callback,
            timeout_seconds,
        } => (
            ibc::Msg::Execute { msgs },
            callback,
            timeout_seconds,
            CallbackRequestType::Execute,
        ),
        ExecuteMsg::Query {
            msgs,
            callback,
            timeout_seconds,
        } => (
            ibc::Msg::Query { msgs },
            Some(callback),
            timeout_seconds,
            CallbackRequestType::Query,
        ),
    };

    let channel_id = CHANNEL
        .may_load(deps.storage)?
        .ok_or(ContractError::NoPair)?;

    let sequence_number = increment_sequence_number(deps.storage, channel_id.clone())?;

    callbacks::request_callback(
        deps.storage,
        deps.api,
        channel_id.clone(),
        sequence_number,
        info.sender.clone(),
        callback,
        request_type,
    )?;

    accounts::on_send_packet(
        deps.storage,
        channel_id.clone(),
        sequence_number,
        &info.sender,
    )?;

    Ok(Response::default()
        .add_attribute("method", "execute")
        .add_message(IbcMsg::SendPacket {
            channel_id,
            data: to_binary(&ibc::Packet {
                sender: info.sender.into_string(),
                msg,
            })
            .expect("msgs are known to be serializable"),
            timeout: IbcTimeout::with_timestamp(env.block.time.plus_seconds(timeout_seconds.u64())),
        }))
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn query(deps: Deps, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {
        QueryMsg::ActiveChannel => to_binary(&CHANNEL.may_load(deps.storage)?),
        QueryMsg::Pair => to_binary(&CONNECTION_REMOTE_PORT.may_load(deps.storage)?.map(
            |(connection_id, remote_port)| Pair {
                connection_id,
                remote_port,
            },
        )),
        QueryMsg::RemoteAddress { local_address } => to_binary(&accounts::query_account(
            deps.storage,
            deps.api.addr_validate(&local_address)?,
        )?),
        QueryMsg::BlockMaxGas => to_binary(&BLOCK_MAX_GAS.load(deps.storage)?),
    }
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn migrate(deps: DepsMut, _env: Env, msg: MigrateMsg) -> Result<Response, ContractError> {
    match msg {
        MigrateMsg::WithUpdate { block_max_gas } => {
            if block_max_gas.u64() <= ERR_GAS_NEEDED {
                return Err(ContractError::GasLimitsMismatch);
            }

            BLOCK_MAX_GAS.save(deps.storage, &block_max_gas.u64())?;
            Ok(Response::default()
                .add_attribute("method", "migrate_with_update")
                .add_attribute("block_max_gas", block_max_gas))
        }
    }
}

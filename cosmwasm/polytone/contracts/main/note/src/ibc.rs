#[cfg(not(feature = "library"))]
use cosmwasm_std::entry_point;
use cosmwasm_std::{
    DepsMut, Env, IbcBasicResponse, IbcChannelCloseMsg, IbcChannelConnectMsg, IbcChannelOpenMsg,
    IbcChannelOpenResponse, IbcPacketAckMsg, IbcPacketReceiveMsg, IbcPacketTimeoutMsg,
    IbcReceiveResponse, Never, Reply, Response, SubMsg,
};
use polytone::{accounts, callbacks, handshake::note};

use crate::{
    error::ContractError,
    state::{BLOCK_MAX_GAS, CHANNEL, CONNECTION_REMOTE_PORT},
};

/// The amount of gas that needs to be reserved for handling a
/// callback error in the reply method. See `TestNoteOutOfGas` in the
/// simulation tests for a test that can be used to tune thise.
pub(crate) const ERR_GAS_NEEDED: u64 = 101_000;

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn ibc_channel_open(
    deps: DepsMut,
    _env: Env,
    msg: IbcChannelOpenMsg,
) -> Result<IbcChannelOpenResponse, ContractError> {
    let response = note::open(&msg, &["JSON-CosmosMsg"])?;
    match CONNECTION_REMOTE_PORT.may_load(deps.storage)? {
        Some((conn, port)) => {
            if msg.channel().counterparty_endpoint.port_id != port
                || msg.channel().connection_id != conn
            {
                Err(ContractError::AlreadyPaired {
                    suggested_connection: msg.channel().connection_id.clone(),
                    suggested_port: msg.channel().counterparty_endpoint.port_id.clone(),
                    pair_connection: conn,
                    pair_port: port,
                })
            } else {
                Ok(response)
            }
        }
        None => Ok(response),
    }
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn ibc_channel_connect(
    deps: DepsMut,
    _env: Env,
    msg: IbcChannelConnectMsg,
) -> Result<IbcBasicResponse, ContractError> {
    note::connect(&msg, &["JSON-CosmosMsg"])?;
    CONNECTION_REMOTE_PORT.save(
        deps.storage,
        &(
            msg.channel().connection_id.clone(),
            msg.channel().counterparty_endpoint.port_id.clone(),
        ),
    )?;
    CHANNEL.save(deps.storage, &msg.channel().endpoint.channel_id)?;
    Ok(IbcBasicResponse::new()
        .add_attribute("method", "ibc_channel_connect")
        .add_attribute("channel_id", &msg.channel().endpoint.channel_id))
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn ibc_channel_close(
    deps: DepsMut,
    _env: Env,
    msg: IbcChannelCloseMsg,
) -> Result<IbcBasicResponse, ContractError> {
    CHANNEL.remove(deps.storage);
    Ok(IbcBasicResponse::default()
        .add_attribute("method", "ibc_channel_close")
        .add_attribute("connection_id", msg.channel().connection_id.clone())
        .add_attribute(
            "counterparty_port_id",
            msg.channel().counterparty_endpoint.port_id.clone(),
        ))
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn ibc_packet_receive(
    _deps: DepsMut,
    _env: Env,
    _msg: IbcPacketReceiveMsg,
) -> Result<IbcReceiveResponse, Never> {
    unreachable!("voice should never send a packet")
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn ibc_packet_ack(
    deps: DepsMut,
    _env: Env,
    ack: IbcPacketAckMsg,
) -> Result<IbcBasicResponse, ContractError> {
    let (callback, executed_by) = callbacks::on_ack(deps.storage, &ack);
    let callback = callback.map(|callback| {
        SubMsg::reply_on_error(callback, ack.original_packet.sequence).with_gas_limit(
            BLOCK_MAX_GAS
                .load(deps.storage)
                .expect("set during instantiation")
                - ERR_GAS_NEEDED,
        )
    });

    accounts::on_ack(
        deps.storage,
        ack.original_packet.src.channel_id.clone(),
        ack.original_packet.sequence,
        executed_by,
    );

    Ok(IbcBasicResponse::default()
        .add_attribute("method", "ibc_packet_ack")
        .add_attribute("sequence_number", ack.original_packet.sequence.to_string())
        .add_submessages(callback))
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn ibc_packet_timeout(
    deps: DepsMut,
    _env: Env,
    msg: IbcPacketTimeoutMsg,
) -> Result<IbcBasicResponse, ContractError> {
    let callback = callbacks::on_timeout(deps.storage, &msg).map(|cosmos_msg| {
        SubMsg::reply_on_error(cosmos_msg, msg.packet.sequence).with_gas_limit(
            BLOCK_MAX_GAS
                .load(deps.storage)
                .expect("set during instantiation")
                - ERR_GAS_NEEDED,
        )
    });

    accounts::on_timeout(deps.storage, msg.packet.src.channel_id, msg.packet.sequence);

    Ok(IbcBasicResponse::default()
        .add_attribute("method", "ibc_packet_timeout")
        .add_attribute("sequence_number", msg.packet.sequence.to_string())
        .add_submessages(callback))
}

#[cfg_attr(not(feature = "library"), entry_point)]
pub fn reply(_deps: DepsMut, _env: Env, msg: Reply) -> Result<Response, ContractError> {
    let sequence = msg.id;
    Ok(Response::default()
        .add_attribute("method", "reply_callback_error")
        .add_attribute("packet_sequence", sequence.to_string())
        .add_attribute("callback_error", msg.result.unwrap_err()))
}

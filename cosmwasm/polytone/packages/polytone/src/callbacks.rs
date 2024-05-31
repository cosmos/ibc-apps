use cosmwasm_schema::cw_serde;
use cosmwasm_std::{
    to_json_binary, Addr, Api, Binary, CosmosMsg, IbcPacketAckMsg, IbcPacketTimeoutMsg, StdResult,
    Storage, SubMsgResponse, Uint64, WasmMsg,
};
use cw_storage_plus::Map;

use crate::ack::unmarshal_ack;

/// Executed on the callback receiver upon message completion. When
/// being executed, the message will be tagged with "callback":
///
/// ```json
/// {"callback": {
///       "initiator": ...,
///       "initiator_msg": ...,
///       "result": ...,
/// }}
/// ```
#[cw_serde]
pub struct CallbackMessage {
    /// Initaitor on the note chain.
    pub initiator: Addr,
    /// Message sent by the initaitor. This _must_ be base64 encoded
    /// or execution will fail.
    pub initiator_msg: Binary,
    /// Data from the host chain.
    pub result: Callback,
}

#[cw_serde]
pub enum Callback {
    /// Result of executing the requested query, or an error.
    ///
    /// result[i] corresponds to the i'th query and contains the
    /// base64 encoded query response.
    Query(Result<Vec<Binary>, ErrorResponse>),

    /// Result of executing the requested messages, or an error.
    ///
    /// 14/04/23: if a submessage errors the reply handler can see
    /// `codespace: wasm, code: 5`, but not the actual error. as a
    /// result, we can't return good errors for Execution and this
    /// error string will only tell you the error's codespace. for
    /// example, an out-of-gas error is code 11 and looks like
    /// `codespace: sdk, code: 11`.
    Execute(Result<ExecutionResponse, String>),

    /// An error occured that could not be recovered from. The only
    /// known way that this can occur is message handling running out
    /// of gas, in which case the error will be `codespace: sdk, code:
    /// 11`.
    ///
    /// This error is not named becuase it could also occur due to a
    /// panic or unhandled error during message processing. We don't
    /// expect this to happen and have carefully written the code to
    /// avoid it.
    FatalError(String),
}

#[cw_serde]
pub struct ExecutionResponse {
    /// The address on the remote chain that executed the messages.
    pub executed_by: String,
    /// Index `i` corresponds to the result of executing the `i`th
    /// message.
    pub result: Vec<SubMsgResponse>,
}

#[cw_serde]
pub struct ErrorResponse {
    /// The index of the first message who's execution failed.
    pub message_index: Uint64,
    /// The error that occured executing the message.
    pub error: String,
}

/// A request for a callback.
#[cw_serde]
pub struct CallbackRequest {
    pub receiver: String,
    pub msg: Binary,
}

/// Disembiguates between a callback for remote message execution and
/// queries.
#[cw_serde]
pub enum CallbackRequestType {
    Execute,
    Query,
}

/// Requests that a callback be returned for the IBC message
/// identified by `(channel_id, sequence_number)`.
pub fn request_callback(
    storage: &mut dyn Storage,
    api: &dyn Api,
    channel_id: String,
    sequence_number: u64,
    initiator: Addr,
    request: Option<CallbackRequest>,
    request_type: CallbackRequestType,
) -> StdResult<()> {
    if let Some(request) = request {
        let receiver = api.addr_validate(&request.receiver)?;
        let initiator_msg = request.msg;

        CALLBACKS.save(
            storage,
            (channel_id, sequence_number),
            &PendingCallback {
                initiator,
                initiator_msg,
                receiver,
                request_type,
            },
        )?;
    }

    Ok(())
}

/// Call on every packet ACK. Returns a callback message to execute,
/// if any, and the address that executed the request on the remote
/// chain (the message initiator's remote account), if any.
///
/// (storage, ack) -> (callback, executed_by)
pub fn on_ack(
    storage: &mut dyn Storage,
    IbcPacketAckMsg {
        acknowledgement,
        original_packet,
        ..
    }: &IbcPacketAckMsg,
) -> (Option<CosmosMsg>, Option<String>) {
    let result = unmarshal_ack(acknowledgement);

    let executed_by = match result {
        Callback::Execute(Ok(ExecutionResponse {
            ref executed_by, ..
        })) => Some(executed_by.clone()),
        _ => None,
    };
    let callback_message = dequeue_callback(
        storage,
        original_packet.src.channel_id.clone(),
        original_packet.sequence,
    )
    .map(|request| callback_message(request, result));

    (callback_message, executed_by)
}

/// Call on every packet timeout. Returns a callback message to execute,
/// if any.
pub fn on_timeout(
    storage: &mut dyn Storage,
    IbcPacketTimeoutMsg { packet, .. }: &IbcPacketTimeoutMsg,
) -> Option<CosmosMsg> {
    let request = dequeue_callback(storage, packet.src.channel_id.clone(), packet.sequence)?;
    let timeout = "timeout".to_string();
    let result = match request.request_type {
        CallbackRequestType::Execute => Callback::Execute(Err(timeout)),
        CallbackRequestType::Query => Callback::Query(Err(ErrorResponse {
            message_index: Uint64::zero(),
            error: timeout,
        })),
    };
    Some(callback_message(request, result))
}

fn callback_message(request: PendingCallback, result: Callback) -> CosmosMsg {
    /// Gives the executed message a "callback" tag:
    /// `{ "callback": CallbackMsg }`.
    #[cw_serde]
    enum C {
        Callback(CallbackMessage),
    }
    WasmMsg::Execute {
        contract_addr: request.receiver.into_string(),
        msg: to_json_binary(&C::Callback(CallbackMessage {
            initiator: request.initiator,
            initiator_msg: request.initiator_msg,
            result,
        }))
        .expect("fields are known to be serializable"),
        funds: vec![],
    }
    .into()
}

fn dequeue_callback(
    storage: &mut dyn Storage,
    channel_id: String,
    sequence_number: u64,
) -> Option<PendingCallback> {
    let request = CALLBACKS
        .may_load(storage, (channel_id.clone(), sequence_number))
        .unwrap()?;
    CALLBACKS.remove(storage, (channel_id, sequence_number));
    Some(request)
}

#[cw_serde]
struct PendingCallback {
    initiator: Addr,
    initiator_msg: Binary,
    /// The address that will receive the callback on completion.
    receiver: Addr,
    /// Used to return the appropriate callback type during timeouts.
    request_type: CallbackRequestType,
}

/// (channel_id, sequence_number) -> callback
const CALLBACKS: Map<(String, u64), PendingCallback> = Map::new("polytone-callbacks");

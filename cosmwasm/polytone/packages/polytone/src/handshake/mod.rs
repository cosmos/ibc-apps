use cosmwasm_std::{
    to_json_binary, Ibc3ChannelOpenResponse, IbcChannelOpenMsg, IbcChannelOpenResponse, IbcOrder,
};

use error::HandshakeError;

pub const POLYTONE_VERSION: &str = "polytone-1";

/// The version returned by the note module during the first step of
/// the handshake.
pub fn note_version() -> String {
    format!("{POLYTONE_VERSION}-note")
}

/// The version returned by the voice module during the first step of
/// the handshake.
pub fn voice_version() -> String {
    format!("{POLYTONE_VERSION}-voice")
}

/// Performs the open step of the IBC handshake for Polytone modules.
///
/// # Arguments
///
/// - `version` the version to return, one of `note_version()`, or
///   `voice_version()`.
/// - `extensions` the Polytone extensions supported by the caller.
///   Extensions are explained in detail in the polytone spec.
/// - `msg` the message received to open the channel.
/// - `counterparty_version` the expected version of the counterparty
///   for example, for a note connecting to a voice, this should be
///   `voice_version()`.
fn open(
    msg: &IbcChannelOpenMsg,
    extensions: &[&str],
    version: String,
    counterparty_version: String,
) -> Result<IbcChannelOpenResponse, HandshakeError> {
    match msg {
        IbcChannelOpenMsg::OpenInit { channel } => {
            if channel.version != POLYTONE_VERSION {
                Err(HandshakeError::ProtocolMismatch {
                    actual: channel.version.clone(),
                    expected: POLYTONE_VERSION.to_string(),
                })
            } else if channel.order != IbcOrder::Unordered {
                Err(HandshakeError::ExpectUnordered)
            } else {
                Ok(Some(Ibc3ChannelOpenResponse { version }))
            }
        }
        IbcChannelOpenMsg::OpenTry {
            channel,
            counterparty_version: cv,
        } => {
            if *cv != counterparty_version {
                Err(HandshakeError::WrongCounterparty)
            } else if channel.order != IbcOrder::Unordered {
                Err(HandshakeError::ExpectUnordered)
            } else {
                Ok(Some(Ibc3ChannelOpenResponse {
                    version: to_json_binary(extensions).unwrap().to_base64(),
                }))
            }
        }
    }
}

pub mod error;
pub mod note;
pub mod voice;

#[cfg(test)]
mod tests;

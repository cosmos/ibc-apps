use super::error::HandshakeError;
use super::{note, voice, POLYTONE_VERSION};
use cosmwasm_std::{
    IbcChannel, IbcChannelConnectMsg, IbcChannelOpenMsg, IbcChannelOpenResponse, IbcEndpoint,
    IbcOrder,
};

type OpenFn = fn(&IbcChannelOpenMsg, &[&str]) -> Result<IbcChannelOpenResponse, HandshakeError>;
type ConnectFn = fn(&IbcChannelConnectMsg, &[&str]) -> Result<(), HandshakeError>;

struct MockHandshake {
    pub init: OpenFn,
    pub try_: OpenFn,
    pub ack: ConnectFn,
    pub confirm: ConnectFn,
}

impl MockHandshake {
    pub fn new(init: OpenFn, try_: OpenFn, ack: ConnectFn, confirm: ConnectFn) -> Self {
        Self {
            init,
            try_,
            ack,
            confirm,
        }
    }

    pub fn run(
        &self,
        start_version: &str,
        start_extensions: &[&str],
        end_extensions: &[&str],
    ) -> Result<(), HandshakeError> {
        let connection_start = "connection-start".to_string();
        let connection_end = "connection-end".to_string();
        let endpoint = IbcEndpoint {
            port_id: "port".to_string(),
            channel_id: "channel".to_string(),
        };
        let mut channel = IbcChannel::new(
            endpoint.clone(),
            endpoint,
            IbcOrder::Unordered,
            start_version.to_string(),
            connection_start.clone(),
        );

        let v = (self.init)(
            &IbcChannelOpenMsg::OpenInit {
                channel: channel.clone(),
            },
            start_extensions,
        )?
        .expect("init should propose a new version")
        .version;

        channel.version = v;
        channel.connection_id = connection_end.clone();

        let v = (self.try_)(
            &IbcChannelOpenMsg::OpenTry {
                channel: channel.clone(),
                counterparty_version: channel.version.clone(),
            },
            end_extensions,
        )?
        .expect("try should propose a new version")
        .version;

        channel.version = v;
        channel.connection_id = connection_start;

        (self.ack)(
            &IbcChannelConnectMsg::OpenAck {
                channel: channel.clone(),
                counterparty_version: channel.version.clone(),
            },
            start_extensions,
        )?;

        channel.connection_id = connection_end;

        (self.confirm)(
            &IbcChannelConnectMsg::OpenConfirm { channel },
            end_extensions,
        )?;

        Ok(())
    }
}

#[test]
fn test_note_to_voice() {
    MockHandshake::new(note::open, voice::open, note::connect, voice::connect)
        .run("polytone-1", &["JSON-CosmosMsg"], &["JSON-CosmosMsg"])
        .unwrap();
}

#[test]
fn test_voice_to_note() {
    MockHandshake::new(voice::open, note::open, voice::connect, note::connect)
        .run("polytone-1", &["JSON-CosmosMsg"], &["JSON-CosmosMsg"])
        .unwrap();
}

#[test]
fn test_voice_to_voice() {
    let err = MockHandshake::new(voice::open, voice::open, voice::connect, voice::connect)
        .run("polytone-1", &["JSON-CosmosMsg"], &["JSON-CosmosMsg"])
        .unwrap_err();
    assert_eq!(err, HandshakeError::WrongCounterparty)
}

#[test]
fn test_note_to_note() {
    let err = MockHandshake::new(note::open, note::open, note::connect, note::connect)
        .run("polytone-1", &["JSON-CosmosMsg"], &["JSON-CosmosMsg"])
        .unwrap_err();
    assert_eq!(err, HandshakeError::WrongCounterparty)
}

#[test]
fn test_wrong_init_version() {
    let err = MockHandshake::new(voice::open, note::open, voice::connect, note::connect)
        .run("ics721-1", &["JSON-CosmosMsg"], &["JSON-CosmosMsg"])
        .unwrap_err();
    assert_eq!(
        err,
        HandshakeError::ProtocolMismatch {
            actual: "ics721-1".to_string(),
            expected: POLYTONE_VERSION.to_string()
        }
    )
}

#[test]
fn test_note_init_extension_rules() {
    let handshake = MockHandshake::new(note::open, voice::open, note::connect, voice::connect);

    // Allowed if note is subset of voice.
    handshake.run("polytone-1", &["a"], &["a", "b"]).unwrap();
    handshake
        .run("polytone-1", &["a", "b"], &["a", "b"])
        .unwrap();

    // Not allowed otherwise.
    let err = handshake
        .run("polytone-1", &["c"], &["a", "b"])
        .unwrap_err();

    assert_eq!(err, HandshakeError::Unspeakable("c".to_string()));
}

#[test]
fn test_voice_init_extension_rules() {
    let handshake = MockHandshake::new(voice::open, note::open, voice::connect, note::connect);

    // Allowed if note is subset of voice.
    handshake.run("polytone-1", &["a", "b"], &["a"]).unwrap();
    handshake
        .run("polytone-1", &["a", "b"], &["a", "b"])
        .unwrap();

    // Not allowed otherwise.
    let err = handshake
        .run("polytone-1", &["c", "e"], &["c", "d", "e", "f"])
        .unwrap_err();

    assert_eq!(err, HandshakeError::Unspeakable("d".to_string()));
}

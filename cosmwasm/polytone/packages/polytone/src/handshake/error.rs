use thiserror::Error;

#[derive(Error, Debug, PartialEq, Eq)]
pub enum HandshakeError {
    #[error("protocol missmatch, got {actual}, expected {expected}")]
    ProtocolMismatch { actual: String, expected: String },
    #[error("channel must be unordered")]
    ExpectUnordered,
    #[error("only a note and voice may connect")]
    WrongCounterparty,
    #[error("note can say ({0}), but voice can not speak it")]
    Unspeakable(String),
}

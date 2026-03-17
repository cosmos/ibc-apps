pub mod note;
pub mod proxy;
pub mod voice;
pub use note::PolytoneNote;
pub use proxy::PolytoneProxy;
pub use voice::PolytoneVoice;

pub mod deploy;

#[derive(Clone)]
pub struct Polytone<Chain: cw_orch::prelude::CwEnv> {
    pub note: PolytoneNote<Chain>,
    pub voice: PolytoneVoice<Chain>,
    pub proxy: PolytoneProxy<Chain>, // This contract doesn't have an address, it's only a code id  used for instantiating
}

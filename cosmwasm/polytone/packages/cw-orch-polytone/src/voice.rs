use cw_orch::{interface, prelude::*};

#[interface(
    polytone_voice::msg::InstantiateMsg,
    polytone_voice::msg::ExecuteMsg,
    polytone_voice::msg::QueryMsg,
    polytone_voice::msg::MigrateMsg
)]
pub struct PolytoneVoice<Chain>;

impl<Chain: CwEnv> Uploadable for PolytoneVoice<Chain> {
    fn wrapper(&self) -> <Mock as TxHandler>::ContractSource {
        Box::new(
            ContractWrapper::new(
                polytone_voice::contract::execute,
                polytone_voice::contract::instantiate,
                polytone_voice::contract::query,
            )
            .with_reply(polytone_voice::ibc::reply)
            .with_ibc(
                polytone_voice::ibc::ibc_channel_open,
                polytone_voice::ibc::ibc_channel_connect,
                polytone_voice::ibc::ibc_channel_close,
                polytone_voice::ibc::ibc_packet_receive,
                polytone_voice::ibc::ibc_packet_ack,
                polytone_voice::ibc::ibc_packet_timeout,
            ),
        )
    }
    fn wasm(&self) -> WasmPath {
        artifacts_dir_from_workspace!()
            .find_wasm_path("polytone_voice")
            .unwrap()
    }
}

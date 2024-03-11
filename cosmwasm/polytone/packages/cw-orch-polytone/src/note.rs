use cw_orch::{interface, prelude::*};

#[interface(
    polytone_note::msg::InstantiateMsg,
    polytone_note::msg::ExecuteMsg,
    polytone_note::msg::QueryMsg,
    polytone_note::msg::MigrateMsg
)]
pub struct PolytoneNote<Chain>;

impl<Chain: CwEnv> Uploadable for PolytoneNote<Chain> {
    fn wrapper(&self) -> <Mock as TxHandler>::ContractSource {
        Box::new(
            ContractWrapper::new(
                polytone_note::contract::execute,
                polytone_note::contract::instantiate,
                polytone_note::contract::query,
            )
            .with_reply(polytone_note::ibc::reply)
            .with_ibc(
                polytone_note::ibc::ibc_channel_open,
                polytone_note::ibc::ibc_channel_connect,
                polytone_note::ibc::ibc_channel_close,
                polytone_note::ibc::ibc_packet_receive,
                polytone_note::ibc::ibc_packet_ack,
                polytone_note::ibc::ibc_packet_timeout,
            ),
        )
    }
    fn wasm(&self) -> WasmPath {
        artifacts_dir_from_workspace!()
            .find_wasm_path("polytone_note")
            .unwrap()
    }
}

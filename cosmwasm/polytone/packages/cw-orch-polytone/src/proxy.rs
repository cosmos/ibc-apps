use cw_orch::{interface, prelude::*};

#[interface(
    polytone_proxy::msg::InstantiateMsg,
    polytone_proxy::msg::ExecuteMsg,
    polytone_proxy::msg::QueryMsg,
    Empty
)]
pub struct PolytoneProxy<Chain>;

impl<Chain: CwEnv> Uploadable for PolytoneProxy<Chain> {
    fn wrapper(&self) -> <Mock as TxHandler>::ContractSource {
        Box::new(
            ContractWrapper::new(
                polytone_proxy::contract::execute,
                polytone_proxy::contract::instantiate,
                polytone_proxy::contract::query,
            )
            .with_reply(polytone_proxy::contract::reply),
        )
    }
    fn wasm(&self) -> WasmPath {
        artifacts_dir_from_workspace!()
            .find_wasm_path("polytone_proxy")
            .unwrap()
    }
}

use cosmwasm_std::{
    testing::{mock_dependencies, mock_env, mock_info},
    to_json_binary, Uint64, WasmMsg,
};

use crate::{
    contract::{execute, instantiate},
    error::ContractError,
    msg::InstantiateMsg,
    state::CHANNEL,
};

#[test]
fn simple_note() {
    let mut deps = mock_dependencies();
    let env = mock_env();
    let info = mock_info("sender", &[]);

    instantiate(
        deps.as_mut(),
        env.clone(),
        info.clone(),
        InstantiateMsg { pair: None },
    )
    .unwrap();
    CHANNEL
        .save(deps.as_mut().storage, &"some_channel".to_string())
        .unwrap();

    execute(
        deps.as_mut(),
        env.clone(),
        info.clone(),
        crate::msg::ExecuteMsg::Execute {
            on_behalf_of: None,
            msgs: vec![WasmMsg::Execute {
                contract_addr: "some_addr".to_string(),
                msg: to_json_binary("some_msg").unwrap(),
                funds: vec![],
            }
            .into()],
            callback: None,
            timeout_seconds: Uint64::new(10000),
        },
    )
    .unwrap();
}

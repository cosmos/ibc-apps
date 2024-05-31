use cosmwasm_std::{to_json_binary, Addr, Empty, Uint64};

use cw_multi_test::{App, Contract, ContractWrapper, Executor};
use polytone::callbacks::{Callback, CallbackMessage};

use crate::{
    error::ContractError,
    msg::{ExecuteMsg, InstantiateMsg, QueryMsg, ResultResponse},
};

pub const CREATOR_ADDR: &str = "creator";
pub const INITIATOR_ADDR: &str = "initiator";
pub const INITIATOR_MSG: &str = "initiator_msg";

fn note_contract() -> Box<dyn Contract<Empty>> {
    let contract = ContractWrapper::new(
        polytone_note::contract::execute,
        polytone_note::contract::instantiate,
        polytone_note::contract::query,
    );
    Box::new(contract)
}

fn listener_contract() -> Box<dyn Contract<Empty>> {
    let contract = ContractWrapper::new(
        crate::contract::execute,
        crate::contract::instantiate,
        crate::contract::query,
    );
    Box::new(contract)
}

#[test]
fn test() {
    let mut app = App::default();

    let note_code = app.store_code(note_contract());
    let listener_code = app.store_code(listener_contract());

    let note1 = app
        .instantiate_contract(
            note_code,
            Addr::unchecked(CREATOR_ADDR),
            &polytone_note::msg::InstantiateMsg {
                pair: None,
                block_max_gas: Uint64::new(110_000),
            },
            &[],
            "note1",
            Some(CREATOR_ADDR.to_string()),
        )
        .unwrap();
    let note2 = app
        .instantiate_contract(
            note_code,
            Addr::unchecked(CREATOR_ADDR),
            &polytone_note::msg::InstantiateMsg {
                pair: None,
                block_max_gas: Uint64::new(110_000),
            },
            &[],
            "note2",
            Some(CREATOR_ADDR.to_string()),
        )
        .unwrap();

    let listener = app
        .instantiate_contract(
            listener_code,
            Addr::unchecked(CREATOR_ADDR),
            &InstantiateMsg {
                note: note1.to_string(),
            },
            &[],
            "listener",
            Some(CREATOR_ADDR.to_string()),
        )
        .unwrap();

    // Returns correct note.
    let queried_note: String = app
        .wrap()
        .query_wasm_smart(listener.clone(), &QueryMsg::Note {})
        .unwrap();
    assert_eq!(queried_note, note1.to_string());

    // Allows note to execute callback.
    let callback = CallbackMessage {
        initiator: Addr::unchecked(INITIATOR_ADDR),
        initiator_msg: to_json_binary(INITIATOR_MSG).unwrap(),
        result: Callback::Execute(Result::Err("ERROR".to_string())),
    };
    app.execute_contract(
        note1,
        listener.clone(),
        &ExecuteMsg::Callback(callback.clone()),
        &[],
    )
    .unwrap();

    // Prevents different note from executing callback.
    let err: ContractError = app
        .execute_contract(
            note2,
            listener.clone(),
            &ExecuteMsg::Callback(callback.clone()),
            &[],
        )
        .unwrap_err()
        .downcast()
        .unwrap();
    assert_eq!(err, ContractError::Unauthorized {});

    // Returns the correct callback.
    let response: ResultResponse = app
        .wrap()
        .query_wasm_smart(
            listener,
            &QueryMsg::Result {
                initiator: INITIATOR_ADDR.to_string(),
                initiator_msg: to_json_binary(INITIATOR_MSG).unwrap().to_string(),
            },
        )
        .unwrap();
    assert_eq!(response.callback, callback);
}

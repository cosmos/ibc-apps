use cosmwasm_std::{Addr, Uint64};

use crate::{error::ContractError, suite_tests::suite::CREATOR_ADDR};

use super::suite::SuiteBuilder;

#[test]
fn test_update() {
    let mut suite = SuiteBuilder::default()
        .with_block_max_gas(Uint64::new(111_000))
        .build();

    suite.assert_block_max_gas(111_000);
    suite.assert_proxy_code(9999);

    let proxy_code_new = suite.store_voice_contract();

    suite
        .update(Addr::unchecked(CREATOR_ADDR), proxy_code_new, 111_000, 32)
        .unwrap();

    // assert that both fields updated succesfully
    suite.assert_block_max_gas(111_000);
    suite.assert_proxy_code(proxy_code_new);
}

#[test]
fn test_query_block_max_gas() {
    let mut suite = SuiteBuilder::default().build();

    suite.assert_block_max_gas(110_000);

    suite
        .update(Addr::unchecked(CREATOR_ADDR), suite.voice_code, 111_000, 32)
        .unwrap();

    suite.assert_block_max_gas(111_000);
}

#[test]
fn test_query_proxy_code_id() {
    let mut suite = SuiteBuilder::default().build();

    suite.assert_proxy_code(9999);

    suite
        .update(Addr::unchecked(CREATOR_ADDR), 1, 110_000, 32)
        .unwrap();

    suite.assert_proxy_code(1);
}

#[test]
fn test_query_contract_addr_len() {
    let mut suite = SuiteBuilder::default().build();

    suite.assert_contract_addr_len(32);

    suite
        .update(Addr::unchecked(CREATOR_ADDR), 1, 110_000, 20)
        .unwrap();

    suite.assert_contract_addr_len(20);
}

#[test]
#[should_panic]
fn test_code_id_validation() {
    SuiteBuilder::default()
        .with_proxy_code_id(Uint64::new(0))
        .build();
}

#[test]
#[should_panic]
fn test_gas_validation() {
    SuiteBuilder::default()
        .with_block_max_gas(Uint64::new(0))
        .build();
}

#[test]
#[should_panic]
fn test_contract_addr_len_min_validation() {
    SuiteBuilder::default()
        .with_contract_addr_len(Some(0))
        .build();
}

#[test]
#[should_panic]
fn test_contract_addr_len_max_validation() {
    SuiteBuilder::default()
        .with_contract_addr_len(Some(33))
        .build();
}

#[test]
fn test_migrate_validation() {
    let mut suite = SuiteBuilder::default().build();

    let err = suite
        .update(Addr::unchecked(CREATOR_ADDR), 0, 110_000, 32)
        .unwrap_err()
        .downcast::<ContractError>()
        .unwrap();

    assert_eq!(err, ContractError::CodeIdCantBeZero);

    let err = suite
        .update(Addr::unchecked(CREATOR_ADDR), 1, 0, 32)
        .unwrap_err()
        .downcast::<ContractError>()
        .unwrap();

    assert_eq!(err, ContractError::GasLimitsMismatch);

    let err = suite
        .update(Addr::unchecked(CREATOR_ADDR), 1, 110_000, 0)
        .unwrap_err()
        .downcast::<ContractError>()
        .unwrap();

    assert_eq!(err, ContractError::ContractAddrLenCantBeZero);

    let err = suite
        .update(Addr::unchecked(CREATOR_ADDR), 1, 110_000, 33)
        .unwrap_err()
        .downcast::<ContractError>()
        .unwrap();

    assert_eq!(err, ContractError::ContractAddrLenCantBeGreaterThan32);
}
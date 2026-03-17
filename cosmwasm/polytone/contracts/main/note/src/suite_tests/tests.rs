use cosmwasm_std::{Addr, Uint64};

use crate::{error::ContractError, msg::Pair};

use super::suite::{SuiteBuilder, CREATOR_ADDR};

#[test]
fn test_instantiate_no_pair() {
    let suite = SuiteBuilder::default()
        .with_block_max_gas(Uint64::new(111_000))
        .build();

    suite.assert_block_max_gas(111_000);
    suite.assert_pair(None);
}

#[test]
fn test_instantiate_with_pair() {
    let pair = Pair {
        connection_id: "id".to_string(),
        remote_port: "port".to_string(),
    };
    let suite = SuiteBuilder::default()
        .with_pair(pair.clone())
        .with_block_max_gas(Uint64::new(111_000))
        .build();

    suite.assert_block_max_gas(111_000);
    suite.assert_pair(Some(pair));
}

#[test]
fn test_update() {
    let mut suite = SuiteBuilder::default().build();

    suite.assert_block_max_gas(110_000);

    suite
        .update(Addr::unchecked(CREATOR_ADDR), 111_000)
        .unwrap();

    suite.assert_block_max_gas(111_000);
}

#[test]
fn test_query_block_max_gas() {
    let suite = SuiteBuilder::default()
        .with_block_max_gas(Uint64::new(111_000))
        .build();

    suite.assert_block_max_gas(111_000);
}

#[test]
#[should_panic]
fn test_gas_validation() {
    SuiteBuilder::default()
        .with_block_max_gas(Uint64::new(0))
        .build();
}

#[test]
fn test_migrate_validation() {
    let mut suite = SuiteBuilder::default().build();

    let err = suite
        .update(Addr::unchecked(CREATOR_ADDR), 0)
        .unwrap_err()
        .downcast::<ContractError>()
        .unwrap();

    assert_eq!(err, ContractError::GasLimitsMismatch);
}

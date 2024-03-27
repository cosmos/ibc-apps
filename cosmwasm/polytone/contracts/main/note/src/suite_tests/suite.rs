use cosmwasm_std::{Addr, Empty, Uint64};
use cw_multi_test::{App, AppResponse, Contract, ContractWrapper, Executor};

use crate::msg::QueryMsg::{ActiveChannel, BlockMaxGas, Pair as PairQuery};
use crate::msg::{InstantiateMsg, MigrateMsg, Pair};

pub const CREATOR_ADDR: &str = "creator";

fn note_contract() -> Box<dyn Contract<Empty>> {
    let contract = ContractWrapper::new(
        crate::contract::execute,
        crate::contract::instantiate,
        crate::contract::query,
    )
    .with_migrate(crate::contract::migrate);
    Box::new(contract)
}

pub(crate) struct Suite {
    app: App,
    pub _admin: Addr,
    pub note_address: Addr,
    pub note_code: u64,
}

pub(crate) struct SuiteBuilder {
    pub instantiate: InstantiateMsg,
}

impl Default for SuiteBuilder {
    fn default() -> Self {
        Self {
            instantiate: InstantiateMsg {
                block_max_gas: Uint64::new(110_000),
                pair: None,
            },
        }
    }
}

impl SuiteBuilder {
    pub fn build(self) -> Suite {
        let mut app = App::default();

        let note_code = app.store_code(note_contract());

        let note_address = app
            .instantiate_contract(
                note_code,
                Addr::unchecked(CREATOR_ADDR),
                &self.instantiate,
                &[],
                "note contract",
                Some(CREATOR_ADDR.to_string()),
            )
            .unwrap();

        Suite {
            app,
            _admin: Addr::unchecked(CREATOR_ADDR),
            note_address,
            note_code,
        }
    }

    pub fn with_block_max_gas(mut self, limit: Uint64) -> Self {
        self.instantiate.block_max_gas = limit;
        self
    }

    pub fn with_pair(mut self, pair: Pair) -> Self {
        self.instantiate.pair = Some(pair);
        self
    }
}

// queries
impl Suite {
    pub fn query_block_max_gas(&self) -> u64 {
        self.app
            .wrap()
            .query_wasm_smart(&self.note_address, &BlockMaxGas)
            .unwrap()
    }

    pub fn query_pair(&self) -> Option<Pair> {
        self.app
            .wrap()
            .query_wasm_smart(&self.note_address, &PairQuery)
            .unwrap()
    }

    pub fn _query_active_channel(&self) -> String {
        self.app
            .wrap()
            .query_wasm_smart(&self.note_address, &ActiveChannel)
            .unwrap()
    }
}

// migrate
impl Suite {
    pub fn update(&mut self, sender: Addr, block_max_gas: u64) -> anyhow::Result<AppResponse> {
        self.app.migrate_contract(
            sender,
            self.note_address.clone(),
            &MigrateMsg::WithUpdate {
                block_max_gas: block_max_gas.into(),
            },
            self.note_code,
        )
    }
}

// assertion helpers
impl Suite {
    pub fn assert_block_max_gas(&self, val: u64) {
        let curr = self.query_block_max_gas();
        assert_eq!(curr, val);
    }

    pub fn assert_pair(&self, val: Option<Pair>) {
        let curr = self.query_pair();
        assert_eq!(curr, val);
    }
}

use cosmwasm_std::{Addr, Empty, Uint64};
use cw_multi_test::{App, AppResponse, Contract, ContractWrapper, Executor};

use crate::msg::QueryMsg::{BlockMaxGas, ContractAddrLen, ProxyCodeId};
use crate::msg::{InstantiateMsg, MigrateMsg};

pub const CREATOR_ADDR: &str = "creator";

fn voice_contract() -> Box<dyn Contract<Empty>> {
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
    pub voice_address: Addr,
    pub voice_code: u64,
}

pub(crate) struct SuiteBuilder {
    pub instantiate: InstantiateMsg,
}

impl Default for SuiteBuilder {
    fn default() -> Self {
        Self {
            instantiate: InstantiateMsg {
                proxy_code_id: Uint64::new(9999),
                block_max_gas: Uint64::new(110_000),
                contract_addr_len: None,
            },
        }
    }
}

impl SuiteBuilder {
    pub fn build(self) -> Suite {
        let mut app = App::default();

        let voice_code = app.store_code(voice_contract());

        let voice_address = app
            .instantiate_contract(
                voice_code,
                Addr::unchecked(CREATOR_ADDR),
                &self.instantiate,
                &[],
                "voice contract",
                Some(CREATOR_ADDR.to_string()),
            )
            .unwrap();

        Suite {
            app,
            _admin: Addr::unchecked(CREATOR_ADDR),
            voice_address,
            voice_code,
        }
    }

    pub fn with_block_max_gas(mut self, limit: Uint64) -> Self {
        self.instantiate.block_max_gas = limit;
        self
    }

    pub fn with_proxy_code_id(mut self, code_id: Uint64) -> Self {
        self.instantiate.proxy_code_id = code_id;
        self
    }

    pub fn with_contract_addr_len(mut self, len: Option<u8>) -> Self {
        self.instantiate.contract_addr_len = len;
        self
    }
}

impl Suite {
    pub fn store_voice_contract(&mut self) -> u64 {
        self.app.store_code(voice_contract())
    }
}

// query
impl Suite {
    pub fn query_block_max_gas(&self) -> u64 {
        self.app
            .wrap()
            .query_wasm_smart(&self.voice_address, &BlockMaxGas)
            .unwrap()
    }

    pub fn query_proxy_code_id(&self) -> u64 {
        self.app
            .wrap()
            .query_wasm_smart(&self.voice_address, &ProxyCodeId)
            .unwrap()
    }

    pub fn query_contract_addr_len(&self) -> u8 {
        self.app
            .wrap()
            .query_wasm_smart(&self.voice_address, &ContractAddrLen)
            .unwrap()
    }
}

// migrate
impl Suite {
    pub fn update(
        &mut self,
        sender: Addr,
        contract_code_id: u64,
        block_max_gas: u64,
        contract_addr_len: u8,
    ) -> anyhow::Result<AppResponse> {
        self.app.migrate_contract(
            sender,
            self.voice_address.clone(),
            &MigrateMsg::WithUpdate {
                proxy_code_id: contract_code_id.into(),
                block_max_gas: block_max_gas.into(),
                contract_addr_len,
            },
            self.voice_code,
        )
    }
}

// assertion helpers
impl Suite {
    pub fn assert_block_max_gas(&self, val: u64) {
        let curr = self.query_block_max_gas();
        assert_eq!(curr, val);
    }

    pub fn assert_proxy_code(&self, val: u64) {
        let curr = self.query_proxy_code_id();
        assert_eq!(curr, val);
    }

    pub fn assert_contract_addr_len(&self, val: u8) {
        let curr = self.query_contract_addr_len();
        assert_eq!(curr, val);
    }
}

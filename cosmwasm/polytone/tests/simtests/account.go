package simtests

import (
	"testing"

	"github.com/CosmWasm/wasmd/app"
	"github.com/CosmWasm/wasmd/x/wasm/ibctesting"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

type Account struct {
	PrivKey    cryptotypes.PrivKey
	PubKey     cryptotypes.PubKey
	Address    sdk.AccAddress
	Acc        authtypes.AccountI
	Chain      *ibctesting.TestChain // lfg garbage collection!!
	SuiteChain *Chain
}

func genAccount(t *testing.T, privkey cryptotypes.PrivKey, suiteChain *Chain) Account {
	chain := suiteChain.Chain
	pubkey := privkey.PubKey()
	addr := sdk.AccAddress(pubkey.Address())

	suiteChain.MintBondedDenom(t, addr)

	accountNumber := chain.App.AccountKeeper.GetNextAccountNumber(chain.GetContext())
	baseAcc := authtypes.NewBaseAccount(addr, pubkey, accountNumber, 0)
	chain.App.AccountKeeper.SetAccount(chain.GetContext(), baseAcc)

	return Account{
		PrivKey:    privkey,
		PubKey:     pubkey,
		Address:    addr,
		Acc:        baseAcc,
		Chain:      chain,
		SuiteChain: suiteChain,
	}
}

// Generates a new account on the provided chain with 100_000_000
// tokens of the chain's bonding denom.
func GenAccount(t *testing.T, suiteChain *Chain) Account {
	privkey := secp256k1.GenPrivKey()
	return genAccount(t, privkey, suiteChain)
}

func (a *Account) KeplrChainDropdownSelect(t *testing.T, selection *Chain) Account {
	if acc := selection.Chain.App.AccountKeeper.GetAccount(selection.Chain.GetContext(), a.Address); acc != nil {
		return Account{
			PrivKey:    a.PrivKey,
			PubKey:     a.PubKey,
			Address:    a.Address,
			Acc:        acc,
			Chain:      selection.Chain,
			SuiteChain: selection,
		}
	} else {
		return genAccount(t, a.PrivKey, selection)
	}
}

// Sends some messages from an account.
func (a *Account) Send(t *testing.T, msgs ...sdk.Msg) (*sdk.Result, error) {
	a.Chain.Coordinator.UpdateTime()

	_, r, err := app.SignAndDeliver(
		t,
		a.Chain.TxConfig,
		a.Chain.App.BaseApp,
		a.Chain.GetContext().BlockHeader(),
		msgs,
		a.Chain.ChainID,
		[]uint64{a.Acc.GetAccountNumber()},
		[]uint64{a.Acc.GetSequence()},
		a.PrivKey,
	)
	if err != nil {
		return r, err
	}

	a.Chain.NextBlock()

	// increment sequence for successful transaction execution
	err = a.Acc.SetSequence(a.Acc.GetSequence() + 1)
	if err != nil {
		return nil, err
	}

	a.Chain.Coordinator.IncrementTime()

	a.Chain.CaptureIBCEvents(r)

	return r, nil
}

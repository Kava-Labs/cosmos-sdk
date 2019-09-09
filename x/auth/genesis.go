package auth

import (
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// InitGenesis - Init store state from genesis data
//
// CONTRACT: old coins from the FeeCollectionKeeper need to be transferred through
// a genesis port script to the new fee collector account
func InitGenesis(ctx sdk.Context, ak AccountKeeper, data GenesisState) {
	ak.SetParams(ctx, data.Params)

	// TODO load old nextAccountNumber
	// TODO don't sort and rewrite account numbers below
	// TODO check for account number duplicates, and address duplicates
	// TODO sanitize coins (if needed)

	// load the accounts
	sort.Slice(data.Accounts, func(i, j int) bool {
		return data.Accounts[i].GetAccountNumber() < data.Accounts[j].GetAccountNumber()
	})
	for _, a := range data.Accounts {
		acc := ak.NewAccount(ctx, a) // set account number
		ak.SetAccount(ctx, acc)
	}
}

// ExportGenesis returns a GenesisState for a given context and keeper
func ExportGenesis(ctx sdk.Context, ak AccountKeeper) GenesisState {
	params := ak.GetParams(ctx)

	// get all the accounts, converting to GenesisAccount type
	var accounts []types.GenesisAccount
	ak.IterateAccounts(ctx, func(acc exported.Account) bool {
		genAcc := acc.(types.GenesisAccount) // will panic if an account doesn't implement GenesisAccount
		accounts = append(accounts, genAcc)
		return false
	})
	// TODO sort accounts (and coins?) to create canonical order for export?

	return NewGenesisState(params, accounts)
}

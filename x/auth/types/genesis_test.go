package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestValidateGenesis(t *testing.T) {
	testCases := []struct {
		name       string
		genState   GenesisState
		expectPass bool
	}{
		{
			"normal",
			GenesisState{
				Params: DefaultParams(),
				Accounts: []GenesisAccount{
					&BaseAccount{
						Address:       sdk.AccAddress("test_account_1______"),
						Coins:         sdk.NewCoins(sdk.NewInt64Coin("chf", 312)),
						AccountNumber: 34},
					&BaseAccount{
						Address:       sdk.AccAddress("test_account_2______"),
						Coins:         sdk.NewCoins(sdk.NewInt64Coin("cad", 39)),
						AccountNumber: 34},
				},
			},
			true,
		},
		{
			"duplicate account",
			GenesisState{
				Params: DefaultParams(),
				Accounts: []GenesisAccount{
					&BaseAccount{AccountNumber: 34},
					&BaseAccount{AccountNumber: 34},
				},
			},
			false,
		},
		{
			"invalid account",
			GenesisState{
				Params:   DefaultParams(),
				Accounts: []GenesisAccount{NewDelayedVestingAccount(&BaseAccount{}, 0)},
			},
			false,
		},
		{
			"no accounts",
			GenesisState{Params: DefaultParams()},
			true,
		},
		{
			"invalid params",
			GenesisState{},
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateGenesis(tc.genState)
			if tc.expectPass {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

package auth_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

// returns context and app with params set on account keeper
func createTestApp(isCheckTx bool) (*simapp.SimApp, sdk.Context) {
	app := simapp.Setup(isCheckTx) // this bool indicated whether initChain should be run with the default genesis state
	ctx := app.BaseApp.NewContext(isCheckTx, abci.Header{})

	return app, ctx
}

func TestExportImport(t *testing.T) {
	// Create keeper, populate accounts
	sourceApp, ctx := createTestApp(false) // run InitChain
	accounts := []auth.Account{
		&auth.BaseAccount{
			Address:  sdk.AccAddress("test_address_1______"), // addresses must be 20 bytes long
			Coins:    sdk.NewCoins(sdk.NewInt64Coin("eur", 34)),
			Sequence: 38456398475,
		},
		auth.NewContinuousVestingAccount(&auth.BaseAccount{Address: sdk.AccAddress("test_address_2______"), Coins: sdk.NewCoins(sdk.NewInt64Coin("yen", 123456))}, 50, 100),
		auth.NewDelayedVestingAccount(&auth.BaseAccount{Address: sdk.AccAddress("test_address_3______")}, 100),
		// TODO add more accounts
	}
	for _, acc := range accounts {
		acc = sourceApp.AccountKeeper.NewAccount(ctx, acc)
		sourceApp.AccountKeeper.SetAccount(ctx, acc)
	}
	sourceApp.Commit()

	// Export accounts json, import into new app
	genState := auth.ExportGenesis(ctx, sourceApp.AccountKeeper)
	importApp, importAppCtx := createTestApp(true) // don't run InitChain
	auth.InitGenesis(importAppCtx, importApp.AccountKeeper, genState)
	// TODO replace above code with below once genaccounts disabled
	// appState, _, err := app.ExportAppStateAndValidators(false, nil)
	// require.NoError(t, err)
	// newApp, newCtx := createTestApp(true) // don't run initChain
	// t.Logf("appState: %s", appState)
	// newApp.InitChain(
	// 	abci.RequestInitChain{
	// 		Validators:    []abci.ValidatorUpdate{},
	// 		AppStateBytes: appState,
	// 	},
	// )
	// newApp.Commit()

	// Compare accounts between the keeper of the source app and the keeper of the new app.
	sourceAccounts := sourceApp.AccountKeeper.GetAllAccounts(ctx)
	// sort.Slice(sourceAccounts, func(i, j int) bool {
	// 	return sourceAccounts[i].GetAccountNumber() < sourceAccounts[j].GetAccountNumber()
	// })
	//t.Logf("   sourceAccounts: \n%s", sourceAccounts)
	importedAccounts := importApp.AccountKeeper.GetAllAccounts(importAppCtx)
	// sort.Slice(importedAccounts, func(i, j int) bool {
	// 	return importedAccounts[i].GetAccountNumber() < importedAccounts[j].GetAccountNumber()
	// })
	// t.Logf("importedAccounts: \n%s", importedAccounts)

	for i := range sourceAccounts {
		require.Equal(t, sourceAccounts[i], importedAccounts[i])
	}

	// Also check the params match
	require.Equal(t, sourceApp.AccountKeeper.GetParams(ctx), importApp.AccountKeeper.GetParams(importAppCtx))
}

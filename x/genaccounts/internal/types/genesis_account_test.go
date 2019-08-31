package types

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	tmtime "github.com/tendermint/tendermint/types/time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/supply"
)

func TestGenesisAccountValidate(t *testing.T) {
	addr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	tests := []struct {
		name   string
		acc    GenesisAccount
		expErr error
	}{
		{
			"valid account",
			NewGenesisAccountRaw(addr, sdk.NewCoins(), sdk.NewCoins(), 0, 0, authexported.VestingPeriods{}, "", ""),
			nil,
		},
		{
			"valid module account",
			NewGenesisAccountRaw(addr, sdk.NewCoins(), sdk.NewCoins(), 0, 0, authexported.VestingPeriods{}, "mint", supply.Minter),
			nil,
		},
		{
			"invalid vesting amount",
			NewGenesisAccountRaw(addr, sdk.NewCoins(sdk.NewInt64Coin("stake", 50)),
				sdk.NewCoins(sdk.NewInt64Coin("stake", 100)), 0, 0, authexported.VestingPeriods{}, "", ""),
			errors.New("vesting amount cannot be greater than total amount"),
		},
		{
			"invalid vesting amount with multi coins",
			NewGenesisAccountRaw(addr,
				sdk.NewCoins(sdk.NewInt64Coin("uatom", 50), sdk.NewInt64Coin("eth", 50)),
				sdk.NewCoins(sdk.NewInt64Coin("uatom", 100), sdk.NewInt64Coin("eth", 20)),
				0, 0, authexported.VestingPeriods{}, "", ""),
			errors.New("vesting amount cannot be greater than total amount"),
		},
		{
			"invalid vesting times",
			NewGenesisAccountRaw(addr, sdk.NewCoins(sdk.NewInt64Coin("stake", 50)),
				sdk.NewCoins(sdk.NewInt64Coin("stake", 50)), 1654668078, 1554668078, authexported.VestingPeriods{}, "", ""),
			errors.New("vesting start-time cannot be before end-time"),
		},
		{
			"vesting periods do not match end time",
			NewGenesisAccountRaw(addr, sdk.NewCoins(sdk.NewInt64Coin("stake", 50)),
				sdk.NewCoins(sdk.NewInt64Coin("stake", 50)), 1654668078, 1654669078, authexported.VestingPeriods{authexported.VestingPeriod{PeriodLength: int64(2000), VestingAmount: sdk.Coins{sdk.NewInt64Coin("stake", 50)}}}, "", ""),
			errors.New("vesting end time does not match length of all vesting periods"),
		},
		{
			"invalid module account name",
			NewGenesisAccountRaw(addr, sdk.NewCoins(), sdk.NewCoins(), 0, 0, authexported.VestingPeriods{}, " ", ""),
			errors.New("module account name cannot be blank"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.acc.Validate()
			require.Equal(t, tt.expErr, err)
		})
	}
}

func TestToAccount(t *testing.T) {
	priv := ed25519.GenPrivKey()
	addr := sdk.AccAddress(priv.PubKey().Address())

	// base account
	authAcc := authtypes.NewBaseAccountWithAddress(addr)
	authAcc.SetCoins(sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 150)))
	genAcc := NewGenesisAccount(&authAcc)
	acc := genAcc.ToAccount()
	require.IsType(t, &authtypes.BaseAccount{}, acc)
	require.Equal(t, &authAcc, acc.(*authtypes.BaseAccount))

	// vesting account
	vacc := authtypes.NewContinuousVestingAccount(
		&authAcc, time.Now().Unix(), time.Now().Add(24*time.Hour).Unix(),
	)
	genAcc, err := NewGenesisAccountI(vacc)
	require.NoError(t, err)
	acc = genAcc.ToAccount()
	require.IsType(t, &authtypes.ContinuousVestingAccount{}, acc)
	require.Equal(t, vacc, acc.(*authtypes.ContinuousVestingAccount))

	// periodic vesting account
	now := tmtime.Now()
	periods := authexported.VestingPeriods{
		authexported.VestingPeriod{int64(12 * 60 * 60), sdk.Coins{sdk.NewInt64Coin("fee", 500), sdk.NewInt64Coin("stake", 50)}},
		authexported.VestingPeriod{int64(6 * 60 * 60), sdk.Coins{sdk.NewInt64Coin("fee", 250), sdk.NewInt64Coin("stake", 25)}},
		authexported.VestingPeriod{int64(6 * 60 * 60), sdk.Coins{sdk.NewInt64Coin("fee", 250), sdk.NewInt64Coin("stake", 25)}},
	}
	pacc := auth.NewPeriodicVestingAccount(
		&authAcc, now.Unix(), periods,
	)
	genAcc, err = NewGenesisAccountI(pacc)
	require.NoError(t, err)
	acc = genAcc.ToAccount()
	require.IsType(t, &auth.PeriodicVestingAccount{}, acc)
	require.Equal(t, pacc, acc.(*auth.PeriodicVestingAccount))

	// module account
	macc := supply.NewEmptyModuleAccount("mint", supply.Minter)
	genAcc, err = NewGenesisAccountI(macc)
	require.NoError(t, err)
	acc = genAcc.ToAccount()
	require.IsType(t, &supply.ModuleAccount{}, acc)
	require.Equal(t, macc, acc.(*supply.ModuleAccount))
}

package exported

import (
	"time"
	"fmt"
	"strings"

	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Account is an interface used to store coins at a given address within state.
// It presumes a notion of sequence numbers for replay protection,
// a notion of account numbers for replay protection for previously pruned accounts,
// and a pubkey for authentication purposes.
//
// Many complex conditions can be used in the concrete struct which implements Account.
type Account interface {
	GetAddress() sdk.AccAddress
	SetAddress(sdk.AccAddress) error // errors if already set.

	GetPubKey() crypto.PubKey // can return nil.
	SetPubKey(crypto.PubKey) error

	GetAccountNumber() uint64
	SetAccountNumber(uint64) error

	GetSequence() uint64
	SetSequence(uint64) error

	GetCoins() sdk.Coins
	SetCoins(sdk.Coins) error

	// Calculates the amount of coins that can be sent to other accounts given
	// the current time.
	SpendableCoins(blockTime time.Time) sdk.Coins

	// Ensure that account implements stringer
	String() string
}

// VestingAccount defines an account type that vests coins via a vesting schedule.
type VestingAccount interface {
	Account

	// Delegation and undelegation accounting that returns the resulting base
	// coins amount.
	TrackDelegation(blockTime time.Time, amount sdk.Coins)
	TrackUndelegation(amount sdk.Coins)

	GetVestedCoins(blockTime time.Time) sdk.Coins
	GetVestingCoins(blockTime time.Time) sdk.Coins

	GetStartTime() int64
	GetEndTime() int64

	GetOriginalVesting() sdk.Coins
	GetDelegatedFree() sdk.Coins
	GetDelegatedVesting() sdk.Coins

	GetVestingPeriods() VestingPeriods
}

// VestingPeriod defines a length of time and amount of coins that will vest
type VestingPeriod struct {
	PeriodLength  int64     // length of the period, in seconds
	VestingAmount sdk.Coins // amount of coins vesting during this period
}

// VestingPeriods stores all vesting periods passed as part of a PeriodicVestingAccount
type VestingPeriods []VestingPeriod

// String VestingPeriod implements stringer interface
func (p VestingPeriod) String() string {
	return fmt.Sprintf(`Period Length: %d
	VestingAmount: %s`, p.PeriodLength, p.VestingAmount)
}

// String VestingPeriods implements stringer interface
func (vp VestingPeriods) String() string {
	var periodsListString []string
	for _, period := range vp {
		periodsListString = append(periodsListString, period.String())
	}
	return strings.TrimSpace(fmt.Sprintf(`Vesting Periods:
		%s`, strings.Join(periodsListString, ", ")))
}
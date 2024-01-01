package v1

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TallyHandler is the interface that is used for tallying votes.
type TallyHandler interface {
	Tally(sdk.Context, Proposal) (passes bool, burnDeposits bool, tallyResults TallyResult)
}

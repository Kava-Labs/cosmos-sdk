package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

// TODO: Break into several smaller functions for clarity

// Tally iterates over the votes and updates the tally of a proposal based on the voting power of the
// voters
func (keeper Keeper) Tally(ctx sdk.Context, proposal v1.Proposal) (passes bool, burnDeposits bool, tallyResults v1.TallyResult) {
	tallyHandler := keeper.tallyHandler
	if tallyHandler == nil {
		tallyHandler = NewDefaultTallyHandler(keeper)
	}

	return tallyHandler.Tally(ctx, proposal)
}

package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

// Tally iterates over the votes and updates the tally of a proposal based on the voting power of the
// voters
func (keeper Keeper) Tally(ctx sdk.Context, proposal types.Proposal) (passes bool, burnDeposits bool, tallyResults types.TallyResult) {
	tallyHandler := keeper.tallyHandler
	if tallyHandler == nil {
		tallyHandler = NewDefaultTallyHandler(keeper)
	}
	return tallyHandler.Tally(ctx, proposal)
}

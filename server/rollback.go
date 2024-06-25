package server

import (
	"fmt"

	tmcmd "github.com/cometbft/cometbft/cmd/cometbft/commands"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/store/rootmulti"
)

// NewRollbackCmd creates a command to rollback tendermint and multistore state by one height.
func NewRollbackCmd(appCreator types.AppCreator, defaultNodeHome string) *cobra.Command {
	var removeBlock bool
	moduleKeysToDelete := make([]string, 0)

	cmd := &cobra.Command{
		Use:   "rollback",
		Short: "rollback cosmos-sdk and tendermint state by one height",
		Long: `
A state rollback is performed to recover from an incorrect application state transition,
when Tendermint has persisted an incorrect app hash and is thus unable to make
progress. Rollback overwrites a state at height n with the state at height n - 1.
The application also rolls back to height n - 1. No blocks are removed, so upon
restarting Tendermint the transactions in block n will be re-executed against the
application.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := GetServerContextFromCmd(cmd)
			cfg := ctx.Config
			home := cfg.RootDir
			db, err := openDB(home, GetAppDBBackend(ctx.Viper))
			if err != nil {
				return err
			}
			app := appCreator(ctx.Logger, db, nil, ctx.Viper)
			// rollback tendermint state
			height, hash, err := tmcmd.RollbackState(ctx.Config, removeBlock)
			if err != nil {
				return fmt.Errorf("failed to rollback tendermint state: %w", err)
			}

			// for rolling back upgrades that add modules, we must first forcibly delete those.
			// otherwise, the rollback will panic because no version of new modules will exist.
			store := app.CommitMultiStore().(*rootmulti.Store)
			for _, key := range moduleKeysToDelete {
				fmt.Printf("deleting latest version of KVStore with key %s\n", key)
				if err := store.DeleteLatestVersion(key); err != nil {
					return err
				}
			}

			// rollback the multistore

			if err := app.CommitMultiStore().RollbackToVersion(height); err != nil {
				return fmt.Errorf("failed to rollback to version: %w", err)
			}

			fmt.Printf("Rolled back state to height %d and hash %X", height, hash)
			return nil
		},
	}

	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "The application home directory")
	cmd.Flags().BoolVar(&removeBlock, "hard", false, "remove last block as well as state")
	cmd.Flags().StringSliceVar(&moduleKeysToDelete, "unsafe-remove-modules", []string{}, "force delete KV stores with the provided prefix. useful for rolling back an upgrade that adds a module")
	return cmd
}

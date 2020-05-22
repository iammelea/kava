package migrate

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/kava-labs/kava/migrate/v0_8"
	v032tendermint "github.com/kava-labs/kava/migrate/v0_8/tendermint/v0_32"
)

const (
	flagGenesisTime = "genesis-time"
	flagChainID     = "chain-id"
)

// MigrateGenesisCmd returns a command to execute genesis state migration.
func MigrateGenesisCmd(_ *server.Context, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "migrate [genesis-file]",
		Short:   "Migrate genesis from kava v0.3 to v0.8",
		Long:    "Migrate the source genesis into the current version, sorts it, and print to STDOUT.",
		Example: fmt.Sprintf(`%s migrate /path/to/genesis.json --chain-id=new-chain-id --genesis-time=1998-01-01T00:00:00Z`, version.ServerName),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			// 1) Unmarshal existing genesis.json

			importGenesis := args[0]
			genDoc, err := v032tendermint.GenesisDocFromFile(importGenesis)
			if err != nil {
				return errors.Wrapf(err, "failed to read genesis document from file %s", importGenesis)
			}

			// 2) Migrate state from kava v0.3 to v0.8

			newGenDoc := v0_8.Migrate(*genDoc)

			// 3) Create and output a new genesis file

			genesisTime := cmd.Flag(flagGenesisTime).Value.String()
			if genesisTime != "" {
				var t time.Time

				err := t.UnmarshalText([]byte(genesisTime))
				if err != nil {
					return errors.Wrap(err, "failed to unmarshal genesis time")
				}

				newGenDoc.GenesisTime = t
			}

			chainID := cmd.Flag(flagChainID).Value.String()
			if chainID != "" {
				newGenDoc.ChainID = chainID
			}

			// TODO assume current app version of codec is good for marshalling tendermint stuff
			bz, err := cdc.MarshalJSONIndent(newGenDoc, "", "  ")
			if err != nil {
				return errors.Wrap(err, "failed to marshal genesis doc")
			}

			sortedBz, err := sdk.SortJSON(bz)
			if err != nil {
				return errors.Wrap(err, "failed to sort JSON genesis doc")
			}

			fmt.Println(string(sortedBz))
			return nil
		},
	}

	cmd.Flags().String(flagGenesisTime, "", "override genesis_time with this flag")
	cmd.Flags().String(flagChainID, "", "override chain_id with this flag")

	return cmd
}

package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	"github.com/buynnex/iaos-evmd/x/lojas/types"
)

// CmdDeleteMerchant remove um merchant pelo ID
func CmdDeleteMerchant() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete-merchant [id]",
		Short: "Deleta um merchant",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			id, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			msg := &types.MsgDeleteMerchant{
				Creator: clientCtx.GetFromAddress().String(),
				Id:      id,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	"github.com/buynnex-corp/byx/x/lojas/types"
)

// CmdUpdateMerchant atualiza um merchant existente
func CmdUpdateMerchant() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-merchant [id] [nome] [endereco] [operator-address]",
		Short: "Atualiza um merchant",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			id, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			msg := &types.MsgUpdateMerchant{
				Creator:         clientCtx.GetFromAddress().String(),
				Id:              id,
				Nome:            args[1],
				Endereco:        args[2],
				OperatorAddress: args[3],
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

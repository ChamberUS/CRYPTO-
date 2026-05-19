package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	"github.com/buynnex-corp/byx/x/lojas/types"
)

// CmdCreateMerchant cria um novo merchant
func CmdCreateMerchant() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-merchant [nome] [endereco] [operator-address]",
		Short: "Cria um novo merchant",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgCreateMerchant{
				Creator:         clientCtx.GetFromAddress().String(),
				Nome:            args[0],
				Endereco:        args[1],
				OperatorAddress: args[2],
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

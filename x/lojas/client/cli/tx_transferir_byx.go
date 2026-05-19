package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	"github.com/buynnex-corp/byx/x/lojas/types"
)

// CmdTransferirByx envia uma MsgTransferirByx (transferência entre lojistas)
func CmdTransferirByx() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transferir-byx [de_lojista_id] [para_lojista_id] [valor]",
		Short: "Transfere BYX de um lojista para outro",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgTransferirByx{
				Creator:       clientCtx.GetFromAddress().String(),
				DeLojistaId:   args[0],
				ParaLojistaId: args[1],
				Valor:         args[2],
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	"github.com/buynnex-corp/byx/x/lojas/types"
)

// CmdFaucet cria o comando CLI para MsgFaucet.
// Uso: byxd tx lojas faucet [lojista-id] [amount] --from <admin>
func CmdFaucet() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "faucet [lojista-id] [amount]",
		Short: "Credita BYX no saldo de um lojista (dev faucet)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgFaucet{
				Creator:   clientCtx.GetFromAddress().String(),
				LojistaId: args[0],
				Amount:    args[1],
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

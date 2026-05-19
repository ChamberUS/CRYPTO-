package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	"github.com/buynnex/iaos-evmd/x/lojas/types"
)

// CmdCreateMerchant cria um novo merchant
func CmdCreateMerchant() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-merchant [nome] [endereco] [cpfcnpj] [telefone] [saldo]",
		Short: "Cria um novo merchant",
		Args:  cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgCreateMerchant{
				Creator:  clientCtx.GetFromAddress().String(),
				Nome:     args[0],
				Endereco: args[1],
				Cpfcnpj:  args[2],
				Telefone: args[3],
				Saldo:    args[4], // string, ex: "1000"
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

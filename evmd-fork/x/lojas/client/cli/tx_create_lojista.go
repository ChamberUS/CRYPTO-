package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	"github.com/buynnex/iaos-evmd/x/lojas/types"
)

// CmdCreateLojista cria um comando CLI para enviar uma MsgCreateLojista
func CmdCreateLojista() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-lojista [nome] [endereco] [cpfcnpj] [telefone]",
		Short: "Cria um novo lojista",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgCreateLojista{
				Creator:  clientCtx.GetFromAddress().String(),
				Nome:     args[0],
				Endereco: args[1],
				Cpfcnpj:  args[2],
				Telefone: args[3],
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "lojas",
		Short:                      "Lojas transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	// Adicione todos os comandos tx aqui
	cmd.AddCommand(CmdCreateLojista())
	cmd.AddCommand(CmdTransferirByx())
	cmd.AddCommand(CmdCreateMerchant())
	cmd.AddCommand(CmdUpdateMerchant())
	cmd.AddCommand(CmdDeleteMerchant())

	// Novo: Faucet
	cmd.AddCommand(CmdFaucet())

	return cmd
}

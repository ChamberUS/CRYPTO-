package main

import (
	"fmt"
	"os"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: %s <module-name>\n", os.Args[0])
		os.Exit(1)
	}

	cfg := sdk.GetConfig()
	cfg.SetBech32PrefixForAccount("byx", "byxpub")

	fmt.Println(authtypes.NewModuleAddress(os.Args[1]).String())
}


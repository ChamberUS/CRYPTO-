package app

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/module"

	clienttypes "github.com/cosmos/ibc-go/v10/modules/core/02-client/types"
	solomachine "github.com/cosmos/ibc-go/v10/modules/light-clients/06-solomachine"
	ibctm "github.com/cosmos/ibc-go/v10/modules/light-clients/07-tendermint"
)

// RegisterIBC mantém compatibilidade com root.go (assinatura simples)
// e NÃO faz wiring manual de módulos IBC (depinject + app.go já cuidam).
// Apenas garante o registro das interfaces dos light clients.
func RegisterIBC(cdc codec.Codec) map[string]module.AppModule {
	// registra interfaces necessárias para (de)codificação e genesis import/export
	clienttypes.RegisterInterfaces(cdc.InterfaceRegistry())
	ibctm.RegisterInterfaces(cdc.InterfaceRegistry())
	solomachine.RegisterInterfaces(cdc.InterfaceRegistry())

	// nenhum módulo extra aqui — o app.go já registra os módulos IBC.
	return map[string]module.AppModule{}
}

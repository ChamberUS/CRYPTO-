package keeper

import "github.com/buynnex/iaos-evmd/x/lojas/types"

type msgServer struct {
	Keeper // <- EMBUTIDO! assim k.GetMerchant / k.SetMerchant passam a existir
}

func NewMsgServerImpl(keeper Keeper) *msgServer {
	return &msgServer{Keeper: keeper}
}

// se tiver o handler de Params, mantenha:
var _ types.MsgServer = (*msgServer)(nil)

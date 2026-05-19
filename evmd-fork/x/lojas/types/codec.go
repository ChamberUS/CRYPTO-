// x/lojas/types/codec.go
package types

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	msgservice "github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterInterfaces(reg codectypes.InterfaceRegistry) {
	reg.RegisterImplementations((*sdk.Msg)(nil),
		&MsgCreateMerchant{},
		&MsgUpdateMerchant{},
		&MsgDeleteMerchant{},
		&MsgCreateLojista{},
		&MsgTransferirByx{},
		&MsgUpdateParams{},
	)
	msgservice.RegisterMsgServiceDesc(reg, &_Msg_serviceDesc)
}

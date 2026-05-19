package keeper

import (
	"context"

	"github.com/buynnex-corp/byx/x/lojas/types"

	errorsmod "cosmossdk.io/errors"
)

func (k msgServer) CreateLojista(ctx context.Context, msg *types.MsgCreateLojista) (*types.MsgCreateLojistaResponse, error) {
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid authority address")
	}

	// TODO: Handle the message

	return &types.MsgCreateLojistaResponse{}, nil
}

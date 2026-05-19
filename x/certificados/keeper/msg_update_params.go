package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/buynnex-corp/byx/x/certificados/types"
)

// UpdateParams updates the module parameters.
func (m msgServer) UpdateParams(goCtx context.Context, req *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if _, err := m.addressCodec.StringToBytes(req.Authority); err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "invalid authority")
	}
	authStr, err := m.addressCodec.BytesToString(m.authority)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to convert authority")
	}
	if authStr != req.Authority {
		return nil, errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "expected %s got %s", authStr, req.Authority)
	}

	if err := req.Params.Validate(); err != nil {
		return nil, err
	}
	if err := m.ParamsStore.Set(ctx, req.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}

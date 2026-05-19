package keeper

import (
	"context"
	"strconv"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/buynnex-corp/byx/x/certificados/types"
)

// TransferCertificate transfere a propriedade para um novo owner.
func (m msgServer) TransferCertificate(goCtx context.Context, msg *types.MsgTransferCertificate) (*types.MsgTransferCertificateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := m.ensureEnabled(ctx); err != nil {
		return nil, err
	}
	if err := m.ensureTransferAllowed(ctx); err != nil {
		return nil, err
	}

	if msg.NewOwner == "" {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "new_owner is required")
	}

	if _, err := m.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "invalid creator")
	}
	if _, err := m.addressCodec.StringToBytes(msg.NewOwner); err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "invalid new_owner")
	}

	cert, err := m.getCertificate(ctx, msg.CertificateId)
	if err != nil {
		return nil, err
	}

	if cert.Revoked {
		return nil, types.ErrCertificateRevoked
	}
	if cert.Owner != msg.Creator {
		return nil, errorsmod.Wrap(types.ErrOwnerMismatch, "creator must be current owner")
	}

	if err := m.updateOwnerIndex(ctx, cert, msg.NewOwner); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"certificados_transfer",
			sdk.NewAttribute("id", strconv.FormatUint(cert.Id, 10)),
			sdk.NewAttribute("from", cert.Owner),
			sdk.NewAttribute("to", msg.NewOwner),
		),
	)

	return &types.MsgTransferCertificateResponse{}, nil
}

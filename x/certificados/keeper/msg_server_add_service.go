package keeper

import (
	"context"
	"strconv"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/buynnex-corp/byx/x/certificados/types"
)

// AddServiceRecord adiciona um registro de manutenção/upgrade.
func (m msgServer) AddServiceRecord(goCtx context.Context, msg *types.MsgAddServiceRecord) (*types.MsgAddServiceRecordResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := m.ensureEnabled(ctx); err != nil {
		return nil, err
	}

	if msg.Kind == "" {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "kind is required")
	}
	if _, err := m.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "invalid creator")
	}

	cert, err := m.getCertificate(ctx, msg.CertificateId)
	if err != nil {
		return nil, err
	}
	if cert.Revoked {
		return nil, types.ErrCertificateRevoked
	}
	if cert.Owner != msg.Creator && cert.Issuer != msg.Creator {
		return nil, errorsmod.Wrap(types.ErrInvalidServiceRequest, "only owner or issuer can add service records")
	}

	record, err := m.addServiceRecord(ctx, cert.Id, msg.Creator, msg.Kind, msg.Details, ctx.BlockTime())
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"certificados_service",
			sdk.NewAttribute("id", strconv.FormatUint(cert.Id, 10)),
			sdk.NewAttribute("added_by", msg.Creator),
			sdk.NewAttribute("type", msg.Kind),
		),
	)

	return &types.MsgAddServiceRecordResponse{Record: &record}, nil
}

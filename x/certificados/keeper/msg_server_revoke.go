package keeper

import (
	"context"
	"errors"
	"strconv"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/buynnex-corp/byx/x/certificados/types"
)

// RevokeCertificate marca um certificado como revogado.
func (m msgServer) RevokeCertificate(goCtx context.Context, msg *types.MsgRevokeCertificate) (*types.MsgRevokeCertificateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := m.ensureEnabled(ctx); err != nil {
		return nil, err
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

	merchant, err := m.lojasKeeper.GetMerchant(goCtx, cert.MerchantId)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, errorsmod.Wrap(sdkerrors.ErrKeyNotFound, "merchant not found")
		}
		return nil, err
	}

	if cert.Issuer != msg.Creator && merchant.Creator != msg.Creator {
		return nil, errorsmod.Wrap(types.ErrNotMerchantOwner, "only issuer/merchant owner can revoke")
	}

	cert.Revoked = true
	cert.RevokedReason = msg.Reason

	if err := m.Certificates.Set(ctx, cert.Id, cert); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"certificados_revoke",
			sdk.NewAttribute("id", strconv.FormatUint(cert.Id, 10)),
			sdk.NewAttribute("reason", msg.Reason),
		),
	)

	return &types.MsgRevokeCertificateResponse{}, nil
}

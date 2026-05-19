package keeper

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/buynnex-corp/byx/x/certificados/types"
	lojastypes "github.com/buynnex-corp/byx/x/lojas/types"
)

// IssueCertificate cria um novo certificado.
func (m msgServer) IssueCertificate(goCtx context.Context, msg *types.MsgIssueCertificate) (*types.MsgIssueCertificateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := m.ensureEnabled(ctx); err != nil {
		return nil, err
	}

	if msg.MerchantId == 0 {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "merchant_id must be > 0")
	}
	if msg.SerialHash == "" {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "serial_hash is required")
	}
	if msg.ImageUri == "" || msg.ImageSha256 == "" || msg.ImageSeed == "" {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "image_uri, image_sha256 and image_seed are required")
	}

	issuerBz, err := m.addressCodec.StringToBytes(msg.Creator)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "invalid creator")
	}
	issuerAddr := sdk.AccAddress(issuerBz)

	owner := msg.Owner
	if owner == "" {
		owner = msg.Creator
	}
	if _, err := m.addressCodec.StringToBytes(owner); err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "invalid owner")
	}

	merchant, err := m.lojasKeeper.GetMerchant(goCtx, msg.MerchantId)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, errorsmod.Wrap(sdkerrors.ErrKeyNotFound, "merchant not found")
		}
		return nil, err
	}
	if merchant.Creator != msg.Creator {
		return nil, errorsmod.Wrap(types.ErrNotMerchantOwner, "creator must be merchant owner")
	}

	params := m.ensureParams(ctx)
	fee := sdk.NewCoin(lojastypes.DenomBYX, sdkmath.NewIntFromUint64(params.IssueFeeByx))
	if err := m.bankKeeper.SendCoinsFromAccountToModule(ctx, issuerAddr, types.ModuleName, sdk.NewCoins(fee)); err != nil {
		return nil, err
	}

	id, err := m.nextCertificateID(ctx)
	if err != nil {
		return nil, err
	}

	now := ctx.BlockTime().UTC()
	cert := types.Certificate{
		Id:            id,
		MerchantId:    msg.MerchantId,
		Issuer:        msg.Creator,
		Owner:         owner,
		Category:      strings.ToUpper(msg.Category),
		Brand:         msg.Brand,
		Model:         msg.Model,
		SerialHash:    strings.ToLower(msg.SerialHash),
		Condition:     msg.Condition,
		Notes:         msg.Notes,
		ImageUri:      msg.ImageUri,
		ImageSha256:   strings.ToLower(msg.ImageSha256),
		ImageSeed:     msg.ImageSeed,
		Revoked:       false,
		CreatedAt:     now.Format(time.RFC3339),
		RevokedReason: "",
	}

	if err := m.storeCertificate(ctx, cert); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"certificados_issue",
			sdk.NewAttribute("id", strconv.FormatUint(id, 10)),
			sdk.NewAttribute("merchant_id", strconv.FormatUint(cert.MerchantId, 10)),
			sdk.NewAttribute("owner", cert.Owner),
			sdk.NewAttribute("category", cert.Category),
			sdk.NewAttribute("image_sha256", cert.ImageSha256),
		),
	)

	return &types.MsgIssueCertificateResponse{Id: id}, nil
}

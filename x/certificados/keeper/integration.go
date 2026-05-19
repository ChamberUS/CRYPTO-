package keeper

import (
	"context"
	"strconv"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/buynnex-corp/byx/x/certificados/types"
)

// GetCertificate exposes certificate retrieval for cross-module integrations.
func (k Keeper) GetCertificate(ctx context.Context, id uint64) (types.Certificate, error) {
	return k.getCertificate(ctx, id)
}

// GetParams returns the current params (or defaults if not set).
func (k Keeper) GetParams(ctx context.Context) (types.Params, error) {
	return k.ensureParams(ctx), nil
}

// TransferCertificate transfers a certificate between owners, enforcing the same
// rules as MsgTransferCertificate (enabled, allow_transfer, not revoked, owner match).
// This method also emits the `certificados_transfer` event.
func (k Keeper) TransferCertificate(ctx context.Context, id uint64, from, to string) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if err := k.ensureEnabled(sdkCtx); err != nil {
		return err
	}
	if err := k.ensureTransferAllowed(sdkCtx); err != nil {
		return err
	}
	if from == "" || to == "" {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "from/to is required")
	}
	if _, err := k.addressCodec.StringToBytes(from); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "invalid from address")
	}
	if _, err := k.addressCodec.StringToBytes(to); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "invalid to address")
	}

	cert, err := k.getCertificate(sdkCtx, id)
	if err != nil {
		return err
	}
	if cert.Revoked {
		return types.ErrCertificateRevoked
	}
	if cert.Owner != from {
		return errorsmod.Wrap(types.ErrOwnerMismatch, "from must be current owner")
	}

	if err := k.updateOwnerIndex(sdkCtx, cert, to); err != nil {
		return err
	}

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			"certificados_transfer",
			sdk.NewAttribute("id", strconv.FormatUint(id, 10)),
			sdk.NewAttribute("from", from),
			sdk.NewAttribute("to", to),
		),
	)

	return nil
}


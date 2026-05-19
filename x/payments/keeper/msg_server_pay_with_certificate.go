package keeper

import (
	"context"
	"strconv"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	certificadostypes "github.com/buynnex-corp/byx/x/certificados/types"
	"github.com/buynnex-corp/byx/x/payments/types"
)

// PayWithCertificate pays a pending payment request and transfers a certificate from the payer to the merchant owner.
func (m msgServer) PayWithCertificate(goCtx context.Context, msg *types.MsgPayWithCertificate) (*types.MsgPayWithCertificateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "invalid request")
	}
	if msg.RequestId == "" {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "request_id is required")
	}
	if msg.CertificateId == 0 {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "certificate_id must be > 0")
	}
	if _, err := m.addressCodec.StringToBytes(msg.Payer); err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "invalid payer")
	}

	requestID, err := strconv.ParseUint(msg.RequestId, 10, 64)
	if err != nil || requestID == 0 {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "request_id must be a positive uint64 string")
	}

	pr, found := m.GetPaymentRequest(ctx, requestID)
	if !found {
		return nil, errorsmod.Wrap(types.ErrPaymentRequestNotFound, "request not found")
	}

	// Preflight validations to avoid side effects on failure (bank transfer, status update).
	if err := m.ensureCurrentStatus(ctx, &pr); err != nil {
		return nil, err
	}
	if pr.Status == types.PaymentStatus_PAYMENT_STATUS_EXPIRED {
		return nil, errorsmod.Wrap(types.ErrPaymentRequestExpired, "payment request expired")
	}
	if !pr.IsPending() {
		return nil, errorsmod.Wrap(types.ErrPaymentRequestNotPending, "payment request already processed")
	}

	certParams, err := m.certificadosKeeper.GetParams(sdk.WrapSDKContext(ctx))
	if err != nil {
		return nil, err
	}
	if !certParams.Enabled {
		return nil, certificadostypes.ErrModuleDisabled
	}
	if !certParams.AllowTransfer {
		return nil, certificadostypes.ErrTransferNotAllowed
	}

	cert, err := m.certificadosKeeper.GetCertificate(sdk.WrapSDKContext(ctx), msg.CertificateId)
	if err != nil {
		return nil, err
	}
	if cert.Revoked {
		return nil, certificadostypes.ErrCertificateRevoked
	}
	if cert.Owner != msg.Payer {
		return nil, certificadostypes.ErrOwnerMismatch
	}

	merchantOwner, err := m.payAndMarkPaid(ctx, &pr, msg.Payer)
	if err != nil {
		return nil, err
	}

	if err := m.certificadosKeeper.TransferCertificate(sdk.WrapSDKContext(ctx), msg.CertificateId, msg.Payer, merchantOwner); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"payments_paid_with_certificate",
			sdk.NewAttribute("request_id", msg.RequestId),
			sdk.NewAttribute("certificate_id", strconv.FormatUint(msg.CertificateId, 10)),
			sdk.NewAttribute("payer", msg.Payer),
			sdk.NewAttribute("merchant_owner", merchantOwner),
		),
	)

	return &types.MsgPayWithCertificateResponse{PaymentRequest: &pr}, nil
}

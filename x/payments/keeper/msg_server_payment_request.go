package keeper

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	lojas "github.com/buynnex-corp/byx/x/lojas/types"
	"github.com/buynnex-corp/byx/x/payments/types"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// CreatePaymentRequest cria um pedido PENDING associado a uma loja.
func (m msgServer) CreatePaymentRequest(goCtx context.Context, msg *types.MsgCreatePaymentRequest) (*types.MsgCreatePaymentRequestResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if _, err := m.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "invalid creator")
	}

	if msg.LojaId == 0 {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "loja_id must be > 0")
	}
	if msg.GetAmountUbyx() == 0 {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "amount must be > 0")
	}

	merchant, err := m.lojasKeeper.GetMerchant(goCtx, msg.LojaId)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, errorsmod.Wrap(sdkerrors.ErrKeyNotFound, "merchant not found")
		}
		return nil, err
	}
	// Somente dono da loja (creator) ou operador explícito pode abrir request.
	isOperator := merchant.OperatorAddress != "" && msg.Creator == merchant.OperatorAddress
	if msg.Creator != merchant.Creator && !isOperator {
		return nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "creator not authorized for this merchant")
	}

	params, err := m.Params.Get(ctx)
	if err != nil {
		params = types.DefaultParams()
	}

	expiresIn := msg.ExpiresInSeconds
	if expiresIn == 0 {
		expiresIn = params.DefaultExpiresInSeconds
	}
	if expiresIn < params.MinExpiresInSeconds || expiresIn > params.MaxExpiresInSeconds {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "expires_in_seconds must be between %d and %d", params.MinExpiresInSeconds, params.MaxExpiresInSeconds)
	}

	now := ctx.BlockTime().UTC()
	expiration := now.Add(time.Duration(expiresIn) * time.Second)

	// fingerprint for idempotence
	_, hash := fingerprintAndHash(msg.LojaId, msg.GetAmountUbyx(), msg.Memo)
	existingID, found := m.GetDedupeRequestID(ctx, msg.LojaId, msg.GetAmountUbyx(), msg.Memo)
	if found {
		if existing, ok := m.GetPaymentRequest(ctx, existingID); ok {
			if existing.Status == types.PaymentStatus_PAYMENT_STATUS_PENDING && now.Unix() < int64(existing.ExpiresAtUnix) {
				ctx.EventManager().EmitEvent(
					sdk.NewEvent(
						"byx_payment_request_reused",
						sdk.NewAttribute("request_id", strconv.FormatUint(existing.Id, 10)),
						sdk.NewAttribute("loja_id", strconv.FormatUint(existing.LojaId, 10)),
						sdk.NewAttribute("merchant_id", strconv.FormatUint(existing.LojaId, 10)),
						sdk.NewAttribute("amount_ubyx", strconv.FormatUint(existing.AmountUbyx, 10)),
						sdk.NewAttribute("memo", existing.Memo),
						sdk.NewAttribute("creator", msg.Creator),
						sdk.NewAttribute("fingerprint_hash", fmt.Sprintf("%x", hash[:])),
						sdk.NewAttribute("trace_id", traceIDFromCtx(ctx)),
					),
				)
				return &types.MsgCreatePaymentRequestResponse{Id: existing.Id}, nil
			}
		}
	}

	id := m.GetNextPaymentRequestID(ctx)
	if id == 0 {
		id = 1
	}
	m.SetNextPaymentRequestID(ctx, id+1)

	pr := types.PaymentRequest{
		Id:            id,
		LojaId:        msg.LojaId,
		AmountUbyx:    msg.GetAmountUbyx(),
		Memo:          msg.Memo,
		Status:        types.PaymentStatus_PAYMENT_STATUS_PENDING,
		CreatedAtUnix: uint64(now.Unix()),
		ExpiresAtUnix: uint64(expiration.Unix()),
	}

	m.SetPaymentRequest(ctx, pr)
	if _, found := m.GetPaymentRequest(ctx, pr.Id); !found {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "payment request not persisted")
	}
	m.AddPaymentRequestToLojaIndex(ctx, pr.LojaId, pr.Id)
	// keep collections map for compatibility
	_ = m.PaymentRequests.Set(ctx, pr.Id, pr)
	_ = m.PaymentRequestsByLoja.Set(ctx, collections.Join(pr.LojaId, pr.Id))
	m.SetDedupeRequestID(ctx, pr.LojaId, pr.GetAmountUbyx(), pr.Memo, pr.Id)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"byx_payment_request_created",
			sdk.NewAttribute("request_id", strconv.FormatUint(id, 10)),
			sdk.NewAttribute("loja_id", strconv.FormatUint(msg.LojaId, 10)),
			sdk.NewAttribute("merchant_id", strconv.FormatUint(msg.LojaId, 10)),
			sdk.NewAttribute("amount_ubyx", strconv.FormatUint(msg.GetAmountUbyx(), 10)),
			sdk.NewAttribute("expires_at_unix", strconv.FormatUint(uint64(expiration.Unix()), 10)),
			sdk.NewAttribute("fingerprint_hash", fmt.Sprintf("%x", hash[:])),
			sdk.NewAttribute("trace_id", traceIDFromCtx(ctx)),
		),
	)

	return &types.MsgCreatePaymentRequestResponse{Id: pr.Id}, nil
}

// PayPaymentRequest debits payer and marks the request as paid.
func (m msgServer) PayPaymentRequest(goCtx context.Context, msg *types.MsgPayPaymentRequest) (*types.MsgPayPaymentRequestResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	pr, found := m.GetPaymentRequest(ctx, msg.RequestId)
	if !found {
		return nil, errorsmod.Wrap(types.ErrPaymentRequestNotFound, "request not found")
	}

	if _, err := m.payAndMarkPaid(ctx, &pr, msg.Creator); err != nil {
		return nil, err
	}

	return &types.MsgPayPaymentRequestResponse{PaymentRequest: &pr}, nil
}

func (m msgServer) payAndMarkPaid(ctx sdk.Context, pr *types.PaymentRequest, payer string) (merchantOwner string, err error) {
	payerBz, err := m.addressCodec.StringToBytes(payer)
	if err != nil {
		return "", errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "invalid payer")
	}
	payerAddr := sdk.AccAddress(payerBz)

	if err := m.ensureCurrentStatus(ctx, pr); err != nil {
		return "", err
	}
	if pr.Status == types.PaymentStatus_PAYMENT_STATUS_EXPIRED {
		return "", errorsmod.Wrap(types.ErrPaymentRequestExpired, "payment request expired")
	}
	if !pr.IsPending() {
		return "", errorsmod.Wrap(types.ErrPaymentRequestNotPending, "payment request already processed")
	}

	merchant, err := m.lojasKeeper.GetMerchant(sdk.WrapSDKContext(ctx), pr.LojaId)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return "", errorsmod.Wrap(sdkerrors.ErrKeyNotFound, "merchant not found")
		}
		return "", err
	}

	merchantAddrBz, err := m.addressCodec.StringToBytes(merchant.Creator)
	if err != nil {
		return "", errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "invalid merchant address")
	}
	merchantAddr := sdk.AccAddress(merchantAddrBz)

	amountUbyx := sdk.NewCoin(lojas.DenomBYX, sdkmath.NewIntFromUint64(pr.GetAmountUbyx()))
	if !amountUbyx.Amount.IsPositive() {
		return "", errorsmod.Wrap(sdkerrors.ErrInvalidCoins, "amount must be positive")
	}
	if err := m.bankKeeper.SendCoins(ctx, payerAddr, merchantAddr, sdk.NewCoins(amountUbyx)); err != nil {
		return "", err
	}
	if err := m.lojasKeeper.AddMerchantSaldo(sdk.WrapSDKContext(ctx), pr.LojaId, amountUbyx.Amount); err != nil {
		return "", err
	}

	now := ctx.BlockTime().UTC()
	pr.Status = types.PaymentStatus_PAYMENT_STATUS_PAID
	pr.Payer = payer
	pr.PaidAtUnix = uint64(now.Unix())
	_, hash := fingerprintAndHash(pr.LojaId, pr.GetAmountUbyx(), pr.Memo)

	m.SetPaymentRequest(ctx, *pr)
	_ = m.PaymentRequests.Set(ctx, pr.Id, *pr)
	m.DeleteDedupeRequestID(ctx, pr.LojaId, pr.GetAmountUbyx(), pr.Memo)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"byx_payment_request_paid",
			sdk.NewAttribute("request_id", strconv.FormatUint(pr.Id, 10)),
			sdk.NewAttribute("loja_id", strconv.FormatUint(pr.LojaId, 10)),
			sdk.NewAttribute("merchant_id", strconv.FormatUint(pr.LojaId, 10)),
			sdk.NewAttribute("payer", payer),
			sdk.NewAttribute("amount_ubyx", strconv.FormatUint(pr.GetAmountUbyx(), 10)),
			sdk.NewAttribute("paid_at_unix", strconv.FormatInt(now.Unix(), 10)),
			sdk.NewAttribute("fingerprint_hash", fmt.Sprintf("%x", hash[:])),
			sdk.NewAttribute("trace_id", traceIDFromCtx(ctx)),
		),
	)

	return merchant.Creator, nil
}

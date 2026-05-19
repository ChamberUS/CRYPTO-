package keeper

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/buynnex-corp/byx/x/lojas/types"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func validateKYCStatus(status string) (string, error) {
	if status == "" {
		return "pending", nil
	}
	switch status {
	case "pending", "approved", "rejected", "suspended":
		return status, nil
	default:
		return "", errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "kyc_status inválido")
	}
}

func validateDocumentHash(hash string) error {
	if hash == "" {
		return nil
	}
	if len(hash) != 64 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "document_hash deve ter 64 caracteres hex")
	}
	if _, err := hex.DecodeString(strings.ToLower(hash)); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "document_hash inválido (esperado SHA-256 hex)")
	}
	return nil
}

func (k msgServer) CreateMerchant(ctx context.Context, msg *types.MsgCreateMerchant) (*types.MsgCreateMerchantResponse, error) {
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidAddress, fmt.Sprintf("invalid address: %s", err))
	}
	operator := msg.OperatorAddress
	if operator != "" {
		if _, err := k.addressCodec.StringToBytes(operator); err != nil {
			return nil, errorsmod.Wrap(sdkerrors.ErrInvalidAddress, fmt.Sprintf("invalid operator address: %s", err))
		}
	}
	if err := validateDocumentHash(msg.DocumentHash); err != nil {
		return nil, err
	}
	kycStatus, err := validateKYCStatus(msg.KycStatus)
	if err != nil {
		return nil, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	id := k.AllocateMerchantID(sdkCtx)

	merchant := types.Merchant{
		Id:              id,
		Creator:         msg.Creator,
		Nome:            msg.Nome,
		Endereco:        msg.Endereco,
		Saldo:           "0",
		OperatorAddress: operator,
		KycRef:          msg.KycRef,
		DocumentHash:    msg.DocumentHash,
		KycStatus:       kycStatus,
	}

	if err := k.SetMerchant(sdkCtx, merchant); err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "failed to set merchant")
	}

	k.SetMerchantByCreator(sdkCtx, msg.Creator, id)

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			"byx_merchant_created",
			sdk.NewAttribute("merchant_id", fmt.Sprintf("%d", id)),
			sdk.NewAttribute("creator", msg.Creator),
			sdk.NewAttribute("nome", msg.Nome),
		),
	)

	return &types.MsgCreateMerchantResponse{
		Id: id,
	}, nil
}

func (k msgServer) UpdateMerchant(ctx context.Context, msg *types.MsgUpdateMerchant) (*types.MsgUpdateMerchantResponse, error) {
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidAddress, fmt.Sprintf("invalid address: %s", err))
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	current, found := k.getMerchant(sdkCtx, msg.Id)
	if !found {
		return nil, errorsmod.Wrap(sdkerrors.ErrKeyNotFound, fmt.Sprintf("key %d doesn't exist", msg.Id))
	}

	if msg.Creator != current.Creator {
		return nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "incorrect owner")
	}

	if err := validateDocumentHash(msg.DocumentHash); err != nil {
		return nil, err
	}
	kycStatus, err := validateKYCStatus(msg.KycStatus)
	if err != nil {
		return nil, err
	}

	operator := current.OperatorAddress
	if msg.OperatorAddress != "" {
		operator = msg.OperatorAddress
	}
	if operator != "" {
		if _, err := k.addressCodec.StringToBytes(operator); err != nil {
			return nil, errorsmod.Wrap(sdkerrors.ErrInvalidAddress, fmt.Sprintf("invalid operator address: %s", err))
		}
	}

	merchant := types.Merchant{
		Creator:         current.Creator,
		Id:              current.Id,
		Nome:            msg.Nome,
		Endereco:        msg.Endereco,
		Saldo:           current.Saldo,
		OperatorAddress: operator,
		KycRef:          msg.KycRef,
		DocumentHash:    msg.DocumentHash,
		KycStatus:       kycStatus,
	}

	if err := k.SetMerchant(sdkCtx, merchant); err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "failed to update merchant")
	}
	k.SetMerchantByCreator(sdkCtx, msg.Creator, msg.Id)

	return &types.MsgUpdateMerchantResponse{}, nil
}

func (k msgServer) DeleteMerchant(ctx context.Context, msg *types.MsgDeleteMerchant) (*types.MsgDeleteMerchantResponse, error) {
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidAddress, fmt.Sprintf("invalid address: %s", err))
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	current, found := k.getMerchant(sdkCtx, msg.Id)
	if !found {
		return nil, errorsmod.Wrap(sdkerrors.ErrKeyNotFound, fmt.Sprintf("key %d doesn't exist", msg.Id))
	}

	if msg.Creator != current.Creator {
		return nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "incorrect owner")
	}

	store := k.getStore(sdkCtx)
	store.Delete(k.merchantKey(msg.Id))

	prefix.NewStore(store, types.MerchantByCreatorPrefix).Delete([]byte(msg.Creator))

	return &types.MsgDeleteMerchantResponse{}, nil
}

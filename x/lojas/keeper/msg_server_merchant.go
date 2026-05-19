package keeper

import (
	"context"
	"fmt"

	"github.com/buynnex-corp/byx/x/lojas/types"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k msgServer) CreateMerchant(ctx context.Context, msg *types.MsgCreateMerchant) (*types.MsgCreateMerchantResponse, error) {
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidAddress, fmt.Sprintf("invalid address: %s", err))
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	id := k.AllocateMerchantID(sdkCtx)

	merchant := types.Merchant{
		Id:       id,
		Creator:  msg.Creator,
		Nome:     msg.Nome,
		Endereco: msg.Endereco,
		Cpfcnpj:  msg.Cpfcnpj,
		Telefone: msg.Telefone,
		Saldo:    msg.Saldo,
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
			sdk.NewAttribute("cpfcnpj", msg.Cpfcnpj),
			sdk.NewAttribute("telefone", msg.Telefone),
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

	merchant := types.Merchant{
		Creator:  msg.Creator,
		Id:       msg.Id,
		Nome:     msg.Nome,
		Endereco: msg.Endereco,
		Cpfcnpj:  msg.Cpfcnpj,
		Telefone: msg.Telefone,
		Saldo:    msg.Saldo,
	}

	current, found := k.getMerchant(sdkCtx, msg.Id)
	if !found {
		return nil, errorsmod.Wrap(sdkerrors.ErrKeyNotFound, fmt.Sprintf("key %d doesn't exist", msg.Id))
	}

	if msg.Creator != current.Creator {
		return nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "incorrect owner")
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

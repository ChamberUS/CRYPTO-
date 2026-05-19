package keeper

import (
	"context"
	"strconv"

	"github.com/buynnex-corp/byx/x/lojas/types"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k msgServer) TransferirByx(ctx context.Context, msg *types.MsgTransferirByx) (*types.MsgTransferirByxResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	fromID, err := strconv.ParseUint(msg.DeLojistaId, 10, 64)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "de_lojista_id inválido")
	}
	toID, err := strconv.ParseUint(msg.ParaLojistaId, 10, 64)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "para_lojista_id inválido")
	}
	if fromID == toID {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "origem e destino não podem ser iguais")
	}

	val, ok := sdkmath.NewIntFromString(msg.Valor)
	if !ok || !val.IsPositive() {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "valor inválido")
	}

	from, err := k.GetMerchant(ctx, fromID)
	if err != nil {
		if errorsmod.IsOf(err, collections.ErrNotFound) {
			return nil, errorsmod.Wrapf(sdkerrors.ErrKeyNotFound, "lojista %d não encontrado", fromID)
		}
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "falha ao carregar lojista de origem")
	}

	params, err := k.ParamsStore.Get(sdkCtx)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "falha ao ler params do módulo")
	}
	isOwner := msg.Creator == from.Creator
	isOperator := from.OperatorAddress != "" && msg.Creator == from.OperatorAddress
	isAdmin := params.FaucetAdmin != "" && msg.Creator == params.FaucetAdmin
	if !isOwner && !isOperator && !isAdmin {
		return nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "creator não autorizado para transferir da loja de origem")
	}

	to, err := k.GetMerchant(ctx, toID)
	if err != nil {
		if errorsmod.IsOf(err, collections.ErrNotFound) {
			return nil, errorsmod.Wrapf(sdkerrors.ErrKeyNotFound, "lojista %d não encontrado", toID)
		}
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "falha ao carregar lojista destino")
	}

	fromBal, _ := sdkmath.NewIntFromString(from.Saldo)
	toBal, _ := sdkmath.NewIntFromString(to.Saldo)

	if fromBal.LT(val) {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInsufficientFunds, "saldo insuficiente: %s < %s", fromBal.String(), val.String())
	}

	fromBal = fromBal.Sub(val)
	toBal = toBal.Add(val)

	from.Saldo = fromBal.String()
	to.Saldo = toBal.String()

	if err := k.SetMerchant(sdkCtx, from); err != nil {
		return nil, err
	}
	if err := k.SetMerchant(sdkCtx, to); err != nil {
		return nil, err
	}

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent("lojas_transferir",
			sdk.NewAttribute("de", msg.DeLojistaId),
			sdk.NewAttribute("para", msg.ParaLojistaId),
			sdk.NewAttribute("valor", msg.Valor),
			sdk.NewAttribute("admin", msg.Creator),
		),
	)
	return &types.MsgTransferirByxResponse{}, nil
}

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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Faucet credita BYX no saldo de um lojista, conforme regras em Params.
func (k msgServer) Faucet(goCtx context.Context, msg *types.MsgFaucet) (*types.MsgFaucetResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// 0) Log de debug enxuto (pode deixar para produção, ajuda bastante)
	ctx.Logger().Info("lojas.Faucet called",
		"creator", msg.Creator,
	)

	// 1) Carregar params do módulo
	params, err := k.ParamsStore.Get(ctx)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "falha ao ler params do módulo")
	}

	// 2) Verificar se o faucet está habilitado
	if !params.FaucetEnabled {
		return nil, status.Error(codes.PermissionDenied, "faucet desabilitado pelo módulo")
	}

	// 3) Resolver qual admin deve ser usado (params)
	adminStr := params.FaucetAdmin

	// 4) Em qualquer cenário habilitado, admin deve estar configurado.
	if adminStr == "" {
		return nil, status.Error(codes.FailedPrecondition, "faucet_admin obrigatório quando faucet estiver habilitado")
	}

	creatorAddr, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "creator inválido (endereço mal formatado)")
	}

	adminAddr, err := sdk.AccAddressFromBech32(adminStr)
	if err != nil {
		// erro interno: configuramos FaucetAdmin errado no código ou params
		return nil, status.Error(codes.Internal, "FaucetAdmin mal configurado (bech32 inválido)")
	}

	if !creatorAddr.Equals(adminAddr) {
		return nil, status.Error(codes.PermissionDenied, "apenas o admin pode usar o faucet")
	}

	// 5) Validar amount
	amt, ok := sdkmath.NewIntFromString(msg.Amount)
	if !ok || !amt.IsPositive() {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "amount inválido")
	}

	// Respeitar limite por transação se configurado
	if params.FaucetMaxPerTx != "" {
		if max, ok := sdkmath.NewIntFromString(params.FaucetMaxPerTx); ok && max.IsPositive() && amt.GT(max) {
			return nil, errorsmod.Wrapf(
				sdkerrors.ErrInvalidRequest,
				"valor excede limite por tx (%s)", params.FaucetMaxPerTx,
			)
		}
	}

	// 6) Converter e carregar lojista
	lojistaID, err := strconv.ParseUint(msg.LojistaId, 10, 64)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "lojista_id inválido (esperado numérico)")
	}

	lojista, err := k.GetMerchant(goCtx, lojistaID)
	if err != nil {
		if errorsmod.IsOf(err, collections.ErrNotFound) {
			return nil, errorsmod.Wrapf(sdkerrors.ErrKeyNotFound, "lojista %d não encontrado", lojistaID)
		}
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "falha ao carregar lojista")
	}

	recipient := lojista.OperatorAddress
	if recipient == "" {
		recipient = lojista.Creator
	}
	lojistaAddr, err := sdk.AccAddressFromBech32(recipient)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "endereço operacional do lojista inválido")
	}

	// 7) Somar saldo (string -> Int -> soma -> string)
	saldo, ok := sdkmath.NewIntFromString(lojista.Saldo)
	if !ok {
		saldo = sdkmath.ZeroInt()
	}
	saldo = saldo.Add(amt)
	lojista.Saldo = saldo.String()

	if err := k.SetMerchant(ctx, lojista); err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "falha ao salvar lojista")
	}

	// 8) Crédito em BYX para o destinatário via helper centralizado
	if err := k.MintBYXTo(ctx, lojistaAddr, amt); err != nil {
		// Traduzir erro interno em erro gRPC
		return nil, status.Error(codes.Internal, err.Error())
	}

	// 9) Evento para indexação/log
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"lojas_faucet",
			sdk.NewAttribute("lojista_id", msg.LojistaId),
			sdk.NewAttribute("amount", amt.String()),
			sdk.NewAttribute("admin", msg.Creator),
		),
	)

	return &types.MsgFaucetResponse{}, nil
}

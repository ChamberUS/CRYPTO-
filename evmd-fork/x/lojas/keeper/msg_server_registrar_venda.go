package keeper

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/buynnex/iaos-evmd/x/lojas/types"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (m msgServer) RegistrarVenda(goCtx context.Context, msg *types.MsgRegistrarVenda) (*types.MsgRegistrarVendaResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// 1) Validação básica do creator
	if _, err := m.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "invalid creator address")
	}

	// 2) Converter loja_id para uint64 para buscar o Merchant
	if msg.LojaId == "" {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "loja_id is required")
	}
	lojaIDUint, err := strconv.ParseUint(msg.LojaId, 10, 64)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "invalid loja_id")
	}

	// 3) Buscar merchant
	merchant, err := m.GetMerchant(goCtx, lojaIDUint)
	if err != nil {
		if errorsmod.IsOf(err, collections.ErrNotFound) {
			return nil, errorsmod.Wrap(sdkerrors.ErrKeyNotFound, "merchant not found")
		}
		return nil, err
	}

	// Opcional: garantir que o creator é o mesmo que o creator do merchant
	if merchant.Creator != msg.Creator {
		return nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "only merchant creator can register sales")
	}

	params, err := m.ParamsStore.Get(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to load module params")
	}

	defaults := types.DefaultParams()
	if params.MaxValorVendaEmCentavos == 0 {
		params.MaxValorVendaEmCentavos = defaults.MaxValorVendaEmCentavos
	}
	if params.MaxCashbackMicroByxPorVenda == 0 {
		params.MaxCashbackMicroByxPorVenda = defaults.MaxCashbackMicroByxPorVenda
	}
	if params.MaxCashbackDailyPerLojaMicrobyx == 0 {
		params.MaxCashbackDailyPerLojaMicrobyx = defaults.MaxCashbackDailyPerLojaMicrobyx
	}
	if params.MaxSalesPerBlockPerLoja == 0 {
		params.MaxSalesPerBlockPerLoja = defaults.MaxSalesPerBlockPerLoja
	}

	// 4) Validar valor da venda
	if msg.ValorEmCentavos == 0 {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "valor_em_centavos must be > 0")
	}

	// Limite de valor máximo por venda para evitar abusos.
	if msg.ValorEmCentavos > params.MaxValorVendaEmCentavos {
		return nil, errorsmod.Wrapf(
			sdkerrors.ErrInvalidRequest,
			"valor_em_centavos acima do limite permitido (%d)",
			params.MaxValorVendaEmCentavos,
		)
	}

	// 5) Resolver endereço do cliente (se informado)
	var clienteAddr sdk.AccAddress
	if msg.Cliente != "" {
		addr, err := m.addressCodec.StringToBytes(msg.Cliente)
		if err != nil {
			return nil, errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "invalid cliente address")
		}
		clienteAddr = sdk.AccAddress(addr)
	}

	// 6) Calcular cashback
	cashbackCoin := m.CalculateCashbackFromCentavos(ctx, int64(msg.ValorEmCentavos))
	cashbackAmount := cashbackCoin.Amount

	// Aplica limite máximo de cashback por venda.
	maxCashbackPerSale := sdkmath.NewIntFromUint64(params.MaxCashbackMicroByxPorVenda)
	if cashbackAmount.GT(maxCashbackPerSale) {
		cashbackAmount = maxCashbackPerSale
	}

	if msg.Cliente == "" {
		// sem cliente não há cashback creditado
		cashbackAmount = sdkmath.NewInt(0)
	}

	cashbackMicro := cashbackAmount.Uint64()

	// Limites anti-abuso (diário e por bloco)
	if params.LimitsEnabled {
		// Limite de vendas por bloco por loja (estado compacto)
		if params.MaxSalesPerBlockPerLoja > 0 {
			state, err := m.SalesCountStateByLoja.Get(ctx, lojaIDUint)
			if err != nil && !errors.Is(err, collections.ErrNotFound) {
				return nil, status.Error(codes.Internal, err.Error())
			}
			if state.LastHeight != ctx.BlockHeight() {
				state.LastHeight = ctx.BlockHeight()
				state.Count = 0
			}
			if state.Count >= params.MaxSalesPerBlockPerLoja {
				return nil, errorsmod.Wrap(types.ErrBlockLimitExceeded, "limite de vendas por bloco excedido")
			}
			state.Count++
			if err := m.SalesCountStateByLoja.Set(ctx, lojaIDUint, state); err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
		}

		// Limite diário de cashback por loja
		if params.MaxCashbackDailyPerLojaMicrobyx > 0 && cashbackAmount.IsPositive() {
			todayKey := dayKey(ctx.BlockTime())
			pair := collections.Join(lojaIDUint, todayKey)
			total, err := m.DailyCashbackByLoja.Get(ctx, pair)
			if err != nil {
				if !errors.Is(err, collections.ErrNotFound) {
					return nil, status.Error(codes.Internal, err.Error())
				}
				total = 0
			}
			next := total + cashbackMicro
			if next > params.MaxCashbackDailyPerLojaMicrobyx {
				return nil, errorsmod.Wrap(types.ErrDailyLimitExceeded, "limite diário de cashback excedido")
			}
			if err := m.DailyCashbackByLoja.Set(ctx, pair, next); err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
		}
	}

	// retenção simples: remove dia fora da janela (independe de ter cashback na venda atual)
	if params.LimitsEnabled && params.CashbackDailyRetentionDays > 0 {
		oldDay := dayKey(ctx.BlockTime().AddDate(0, 0, -int(params.CashbackDailyRetentionDays)))
		oldPair := collections.Join(lojaIDUint, oldDay)
		_ = m.DailyCashbackByLoja.Remove(ctx, oldPair)
	}

	// 7) Mint de cashback para o cliente usando helper centralizado
	if cashbackAmount.IsPositive() && msg.Cliente != "" {
		if err := m.MintBYXTo(ctx, clienteAddr, sdkmath.NewIntFromBigInt(cashbackAmount.BigInt())); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	// 8) Persistir venda com timestamp atual
	saleID, err := m.nextSaleID(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	sale := types.Sale{
		Id:               saleID,
		LojaId:           lojaIDUint,
		Creator:          msg.Creator,
		ValorEmCentavos:  msg.ValorEmCentavos,
		Cliente:          msg.Cliente,
		CashbackMicroByx: cashbackMicro,
		Timestamp:        ctx.BlockTime().UTC().Format(time.RFC3339),
		BlockTime:        uint64(ctx.BlockTime().UTC().Unix()),
	}

	if err := m.storeSale(ctx, sale); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if !cashbackAmount.IsPositive() || msg.Cliente == "" {
		// Sem cliente ou cashback = 0 => apenas registra a venda "lógica" (por enquanto sem estado)
		ctx.EventManager().EmitEvent(
			sdk.NewEvent("lojas_registrar_venda",
				sdk.NewAttribute("sale_id", strconv.FormatUint(saleID, 10)),
				sdk.NewAttribute("loja_id", msg.LojaId),
				sdk.NewAttribute("valor_em_centavos", strconv.FormatUint(msg.ValorEmCentavos, 10)),
				sdk.NewAttribute("cashback_micro_byx", cashbackAmount.String()),
				sdk.NewAttribute("cliente", msg.Cliente),
				sdk.NewAttribute("creator", msg.Creator),
			),
		)
		return &types.MsgRegistrarVendaResponse{
			CashbackMicroByx: cashbackMicro,
		}, nil
	}

	// 9) Evento final com cashback > 0
	ctx.EventManager().EmitEvent(
		sdk.NewEvent("lojas_registrar_venda",
			sdk.NewAttribute("sale_id", strconv.FormatUint(saleID, 10)),
			sdk.NewAttribute("loja_id", msg.LojaId),
			sdk.NewAttribute("valor_em_centavos", strconv.FormatUint(msg.ValorEmCentavos, 10)),
			sdk.NewAttribute("cashback_micro_byx", cashbackAmount.String()),
			sdk.NewAttribute("cliente", msg.Cliente),
			sdk.NewAttribute("creator", msg.Creator),
		),
	)

	// 9) Resposta com o valor de cashback creditado
	return &types.MsgRegistrarVendaResponse{
		CashbackMicroByx: cashbackMicro,
	}, nil
}

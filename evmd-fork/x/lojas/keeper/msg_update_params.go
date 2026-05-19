package keeper

import (
	"bytes"
	"context"
	"strconv"

	errorsmod "cosmossdk.io/errors"

	"github.com/buynnex/iaos-evmd/x/lojas/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) UpdateParams(ctx context.Context, req *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	authority, err := k.addressCodec.StringToBytes(req.Authority)
	if err != nil {
		return nil, errorsmod.Wrap(err, "invalid authority address")
	}

	if !bytes.Equal(k.GetAuthority(), authority) {
		expectedAuthorityStr, _ := k.addressCodec.BytesToString(k.GetAuthority())
		return nil, errorsmod.Wrapf(types.ErrInvalidSigner, "invalid authority; expected %s, got %s", expectedAuthorityStr, req.Authority)
	}

	current, err := k.ParamsStore.Get(ctx)
	if err != nil {
		current = types.DefaultParams()
	}

	if err := req.Params.Validate(); err != nil {
		return nil, err
	}

	if err := k.ParamsStore.Set(ctx, req.Params); err != nil {
		return nil, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	ev := sdk.NewEvent(
		"lojas_params_updated",
		sdk.NewAttribute("authority", req.Authority),
		sdk.NewAttribute("faucet_enabled", strconv.FormatBool(req.Params.FaucetEnabled)),
		sdk.NewAttribute("cashback_rate_micro_byx_per_real", strconv.FormatUint(req.Params.CashbackRateMicroByxPerReal, 10)),
		sdk.NewAttribute("max_valor_venda_em_centavos", strconv.FormatUint(req.Params.MaxValorVendaEmCentavos, 10)),
		sdk.NewAttribute("max_cashback_micro_byx_por_venda", strconv.FormatUint(req.Params.MaxCashbackMicroByxPorVenda, 10)),
		sdk.NewAttribute("max_cashback_daily_per_loja_microbyx", strconv.FormatUint(req.Params.MaxCashbackDailyPerLojaMicrobyx, 10)),
		sdk.NewAttribute("max_sales_per_block_per_loja", strconv.FormatUint(uint64(req.Params.MaxSalesPerBlockPerLoja), 10)),
		sdk.NewAttribute("limits_enabled", strconv.FormatBool(req.Params.LimitsEnabled)),
		sdk.NewAttribute("cashback_daily_retention_days", strconv.FormatUint(uint64(req.Params.CashbackDailyRetentionDays), 10)),
	)

	// diffs por campo
	emitDiff := func(name, oldVal, newVal string) {
		if oldVal == newVal {
			return
		}
		ev = ev.AppendAttributes(
			sdk.NewAttribute("param_name", name),
			sdk.NewAttribute("old_value", oldVal),
			sdk.NewAttribute("new_value", newVal),
		)
	}
	emitDiff("faucet_enabled", strconv.FormatBool(current.FaucetEnabled), strconv.FormatBool(req.Params.FaucetEnabled))
	emitDiff("faucet_admin", current.FaucetAdmin, req.Params.FaucetAdmin)
	emitDiff("faucet_max_per_tx", current.FaucetMaxPerTx, req.Params.FaucetMaxPerTx)
	emitDiff("cashback_rate_micro_byx_per_real", strconv.FormatUint(current.CashbackRateMicroByxPerReal, 10), strconv.FormatUint(req.Params.CashbackRateMicroByxPerReal, 10))
	emitDiff("max_valor_venda_em_centavos", strconv.FormatUint(current.MaxValorVendaEmCentavos, 10), strconv.FormatUint(req.Params.MaxValorVendaEmCentavos, 10))
	emitDiff("max_cashback_micro_byx_por_venda", strconv.FormatUint(current.MaxCashbackMicroByxPorVenda, 10), strconv.FormatUint(req.Params.MaxCashbackMicroByxPorVenda, 10))
	emitDiff("max_cashback_daily_per_loja_microbyx", strconv.FormatUint(current.MaxCashbackDailyPerLojaMicrobyx, 10), strconv.FormatUint(req.Params.MaxCashbackDailyPerLojaMicrobyx, 10))
	emitDiff("max_sales_per_block_per_loja", strconv.FormatUint(uint64(current.MaxSalesPerBlockPerLoja), 10), strconv.FormatUint(uint64(req.Params.MaxSalesPerBlockPerLoja), 10))
	emitDiff("limits_enabled", strconv.FormatBool(current.LimitsEnabled), strconv.FormatBool(req.Params.LimitsEnabled))
	emitDiff("cashback_daily_retention_days", strconv.FormatUint(uint64(current.CashbackDailyRetentionDays), 10), strconv.FormatUint(uint64(req.Params.CashbackDailyRetentionDays), 10))

	sdkCtx.EventManager().EmitEvent(ev)

	return &types.MsgUpdateParamsResponse{}, nil
}

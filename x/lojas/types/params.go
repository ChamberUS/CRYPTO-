package types

import (
	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// NewParams cria um Params preenchendo os 3 campos do faucet.
func NewParams(admin string, enabled bool, maxPerTx string) Params {
	return Params{
		FaucetAdmin:                     admin,
		FaucetEnabled:                   enabled,
		FaucetMaxPerTx:                  maxPerTx,
		CashbackRateMicroByxPerReal:     2500,
		MaxValorVendaEmCentavos:         1_000_000,
		MaxCashbackMicroByxPorVenda:     100_000,
		MaxCashbackDailyPerLojaMicrobyx: 5_000_000,
		MaxSalesPerBlockPerLoja:         20,
		LimitsEnabled:                   true,
		CashbackDailyRetentionDays:      90,
	}
}

// DefaultParams define os defaults usados no init genesis.
func DefaultParams() Params {
	return Params{
		// Safe-by-default: faucet desabilitado até configuração explícita.
		FaucetAdmin:                     "",
		FaucetEnabled:                   false,
		FaucetMaxPerTx:                  "1000", // ex.: 1000 BYX por tx
		CashbackRateMicroByxPerReal:     2500,   // 0.0025 BYX por R$1 (micro BYX)
		MaxValorVendaEmCentavos:         1_000_000,
		MaxCashbackMicroByxPorVenda:     100_000,
		MaxCashbackDailyPerLojaMicrobyx: 5_000_000,
		MaxSalesPerBlockPerLoja:         20,
		LimitsEnabled:                   true,
		CashbackDailyRetentionDays:      90,
	}
}

// Validate valida coerência dos params.
func (p Params) Validate() error {
	// faucet_max_per_tx deve ser Int >= 0 (string Int)
	// Obs: permitimos vazio (sem limite); quando informado, valida range.
	if p.FaucetMaxPerTx != "" {
		v, ok := sdkmath.NewIntFromString(p.FaucetMaxPerTx)
		if !ok || v.IsNegative() {
			return errorsmod.Wrapf(
				sdkerrors.ErrInvalidRequest,
				"faucet_max_per_tx inválido: %q", p.FaucetMaxPerTx,
			)
		}
		if v.GT(sdkmath.NewInt(1_000_000_000)) {
			return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "faucet_max_per_tx muito alto: %s", p.FaucetMaxPerTx)
		}
	}

	if p.CashbackRateMicroByxPerReal == 0 || p.CashbackRateMicroByxPerReal > 1_000_000 {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "cashback_rate_micro_byx_per_real muito alto: %d", p.CashbackRateMicroByxPerReal)
	}

	if p.MaxValorVendaEmCentavos == 0 || p.MaxValorVendaEmCentavos > 10_000_000 {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "max_valor_venda_em_centavos fora do intervalo: %d", p.MaxValorVendaEmCentavos)
	}

	if p.MaxCashbackMicroByxPorVenda == 0 || p.MaxCashbackMicroByxPorVenda > 1_000_000 {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "max_cashback_micro_byx_por_venda fora do intervalo: %d", p.MaxCashbackMicroByxPorVenda)
	}

	if p.MaxCashbackDailyPerLojaMicrobyx > 100_000_000_000 {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "max_cashback_daily_per_loja_microbyx fora do intervalo: %d", p.MaxCashbackDailyPerLojaMicrobyx)
	}

	if p.MaxSalesPerBlockPerLoja > 10_000 {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "max_sales_per_block_per_loja fora do intervalo: %d", p.MaxSalesPerBlockPerLoja)
	}

	if p.CashbackDailyRetentionDays == 0 || p.CashbackDailyRetentionDays > 3650 {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "cashback_daily_retention_days fora do intervalo: %d", p.CashbackDailyRetentionDays)
	}

	if p.FaucetAdmin != "" {
		if _, err := sdk.AccAddressFromBech32(p.FaucetAdmin); err != nil {
			return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, "faucet_admin inválido")
		}
	}
	if p.FaucetEnabled && p.FaucetAdmin == "" {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "faucet_admin é obrigatório quando faucet_enabled=true")
	}
	return nil
}

// ParamSetPairs implementa ParamSet.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair([]byte("FaucetAdmin"), &p.FaucetAdmin, validateStringNotNil),
		paramtypes.NewParamSetPair([]byte("FaucetEnabled"), &p.FaucetEnabled, validateBool),
		paramtypes.NewParamSetPair([]byte("FaucetMaxPerTx"), &p.FaucetMaxPerTx, validateStringNotNil),
		paramtypes.NewParamSetPair([]byte("CashbackRateMicroByxPerReal"), &p.CashbackRateMicroByxPerReal, validateUint64Positive),
		paramtypes.NewParamSetPair([]byte("MaxValorVendaEmCentavos"), &p.MaxValorVendaEmCentavos, validateUint64Positive),
		paramtypes.NewParamSetPair([]byte("MaxCashbackMicroByxPorVenda"), &p.MaxCashbackMicroByxPorVenda, validateUint64Positive),
		paramtypes.NewParamSetPair([]byte("MaxCashbackDailyPerLojaMicrobyx"), &p.MaxCashbackDailyPerLojaMicrobyx, validateUint64Positive),
		paramtypes.NewParamSetPair([]byte("MaxSalesPerBlockPerLoja"), &p.MaxSalesPerBlockPerLoja, validateUint32Positive),
		paramtypes.NewParamSetPair([]byte("LimitsEnabled"), &p.LimitsEnabled, validateBool),
		paramtypes.NewParamSetPair([]byte("CashbackDailyRetentionDays"), &p.CashbackDailyRetentionDays, validateUint32Positive),
	}
}

func validateStringNotNil(i interface{}) error { return nil }

func validateBool(i interface{}) error { return nil }

func validateUint64Positive(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid type %T for uint64 param", i)
	}
	if v == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "value must be > 0")
	}
	return nil
}

func validateUint32Positive(i interface{}) error {
	v, ok := i.(uint32)
	if !ok {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid type %T for uint32 param", i)
	}
	if v == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "value must be > 0")
	}
	return nil
}

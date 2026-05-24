package types

const (
	// ByxExponent define quantas casas decimais existem entre BYX e ubyx.
	ByxExponent = 6
	// UbyxPerByx é o fator de conversão da unidade de display para a unidade base.
	UbyxPerByx = int64(1_000_000)
	// MaxSupplyBYX é o teto absoluto de supply em unidade de display.
	MaxSupplyBYX = int64(1_000_000_000)
	// MaxSupplyUbyx é o teto absoluto de supply em unidade base on-chain.
	MaxSupplyUbyx = MaxSupplyBYX * UbyxPerByx
	// SupplyCapBYX mantém compatibilidade com o código existente, em unidade base.
	SupplyCapBYX = MaxSupplyUbyx
)

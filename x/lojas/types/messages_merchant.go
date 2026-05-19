package types

func NewMsgCreateMerchant(creator string, nome string, endereco string, cpfcnpj string, telefone string, saldo string) *MsgCreateMerchant {
	return &MsgCreateMerchant{
		Creator:  creator,
		Nome:     nome,
		Endereco: endereco,
		Cpfcnpj:  cpfcnpj,
		Telefone: telefone,
		Saldo:    saldo,
	}
}

func NewMsgUpdateMerchant(creator string, id uint64, nome string, endereco string, cpfcnpj string, telefone string, saldo string) *MsgUpdateMerchant {
	return &MsgUpdateMerchant{
		Id:       id,
		Creator:  creator,
		Nome:     nome,
		Endereco: endereco,
		Cpfcnpj:  cpfcnpj,
		Telefone: telefone,
		Saldo:    saldo,
	}
}

func NewMsgDeleteMerchant(creator string, id uint64) *MsgDeleteMerchant {
	return &MsgDeleteMerchant{
		Id:      id,
		Creator: creator,
	}
}

package types

func NewMsgCreateLojista(creator string, nome string, endereco string, cpfcnpj string, telefone string) *MsgCreateLojista {
	return &MsgCreateLojista{
		Creator:  creator,
		Nome:     nome,
		Endereco: endereco,
		Cpfcnpj:  cpfcnpj,
		Telefone: telefone,
	}
}

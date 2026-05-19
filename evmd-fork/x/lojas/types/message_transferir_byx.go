package types

func NewMsgTransferirByx(creator string, deLojistaId string, paraLojistaId string, valor string) *MsgTransferirByx {
	return &MsgTransferirByx{
		Creator:       creator,
		DeLojistaId:   deLojistaId,
		ParaLojistaId: paraLojistaId,
		Valor:         valor,
	}
}

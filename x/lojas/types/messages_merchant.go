package types

func NewMsgCreateMerchant(creator string, nome string, endereco string, operatorAddress string, kycRef string, documentHash string, kycStatus string) *MsgCreateMerchant {
	return &MsgCreateMerchant{
		Creator:         creator,
		Nome:            nome,
		Endereco:        endereco,
		OperatorAddress: operatorAddress,
		KycRef:          kycRef,
		DocumentHash:    documentHash,
		KycStatus:       kycStatus,
	}
}

func NewMsgUpdateMerchant(creator string, id uint64, nome string, endereco string, operatorAddress string, kycRef string, documentHash string, kycStatus string) *MsgUpdateMerchant {
	return &MsgUpdateMerchant{
		Id:              id,
		Creator:         creator,
		Nome:            nome,
		Endereco:        endereco,
		OperatorAddress: operatorAddress,
		KycRef:          kycRef,
		DocumentHash:    documentHash,
		KycStatus:       kycStatus,
	}
}

func NewMsgDeleteMerchant(creator string, id uint64) *MsgDeleteMerchant {
	return &MsgDeleteMerchant{
		Id:      id,
		Creator: creator,
	}
}

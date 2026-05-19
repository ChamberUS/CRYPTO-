package lojas

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	"github.com/buynnex-corp/byx/x/lojas/types"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: types.Query_serviceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Mostra os parâmetros do módulo",
				},
				{
					RpcMethod: "MerchantAll", // lista com paginação
					Use:       "merchant-all",
					Short:     "Lista merchants com paginação",
				},
				{
					RpcMethod:      "Merchant", // <- ajuste aqui se o seu proto usa GetMerchant
					Use:            "merchant [id]",
					Short:          "Consulta um merchant pelo ID",
					Alias:          []string{"show-merchant"},
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "id"}},
				},
				// this line is used by ignite scaffolding # autocli/query
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service:              types.Msg_serviceDesc.ServiceName,
			EnhanceCustomCommand: true,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{RpcMethod: "UpdateParams", Skip: true},
				{
					RpcMethod:      "CreateLojista",
					Use:            "create-lojista [nome] [endereco] [cpfcnpj] [telefone]",
					Short:          "Envia uma tx CreateLojista",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "nome"}, {ProtoField: "endereco"}, {ProtoField: "cpfcnpj"}, {ProtoField: "telefone"}},
				},
				{
					RpcMethod:      "TransferirByx",
					Use:            "transferir-byx [de-lojista-id] [para-lojista-id] [valor]",
					Short:          "Envia uma tx TransferirByx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "de_lojista_id"}, {ProtoField: "para_lojista_id"}, {ProtoField: "valor"}},
				},
				{
					RpcMethod:      "CreateMerchant",
					Use:            "create-merchant [nome] [endereco] [operator-address] [kyc-ref] [document-hash] [kyc-status]",
					Short:          "Cria merchant",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "nome"}, {ProtoField: "endereco"}, {ProtoField: "operator_address"}, {ProtoField: "kyc_ref"}, {ProtoField: "document_hash"}, {ProtoField: "kyc_status"}},
				},
				{
					RpcMethod:      "UpdateMerchant",
					Use:            "update-merchant [id] [nome] [endereco] [operator-address] [kyc-ref] [document-hash] [kyc-status]",
					Short:          "Atualiza merchant",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "id"}, {ProtoField: "nome"}, {ProtoField: "endereco"}, {ProtoField: "operator_address"}, {ProtoField: "kyc_ref"}, {ProtoField: "document_hash"}, {ProtoField: "kyc_status"}},
				},
				{
					RpcMethod:      "DeleteMerchant",
					Use:            "delete-merchant [id]",
					Short:          "Remove merchant",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "id"}},
				},
				// this line is used by ignite scaffolding # autocli/tx
			},
		},
	}
}

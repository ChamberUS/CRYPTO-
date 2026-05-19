package module

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	"github.com/buynnex-corp/byx/x/certificados/types"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: types.Query_serviceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{RpcMethod: "Params", Use: "params", Short: "Mostra os parâmetros do módulo"},
				{
					RpcMethod:      "Certificate",
					Use:            "certificate [id]",
					Short:          "Consulta um certificado pelo ID",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "id"}},
				},
				{
					RpcMethod:      "CertificatesByOwner",
					Use:            "certificates-by-owner [owner]",
					Short:          "Lista certificados por owner",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "owner"}},
				},
				{
					RpcMethod:      "CertificatesByMerchant",
					Use:            "certificates-by-merchant [merchant-id]",
					Short:          "Lista certificados por merchant",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "merchant_id"}},
				},
				{
					RpcMethod:      "CertificatesBySerial",
					Use:            "certificates-by-serial [serial-hash]",
					Short:          "Lista certificados por serial_hash",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "serial_hash"}},
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service:              types.Msg_serviceDesc.ServiceName,
			EnhanceCustomCommand: true,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{RpcMethod: "UpdateParams", Skip: true},
				{
					RpcMethod: "IssueCertificate",
					Use:       "issue-certificate [merchant-id] [category] [brand] [model] [serial-hash] [condition] [image-uri] [image-sha256] [image-seed]",
					Short:     "Emite um certificado vinculado a um merchant",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "merchant_id"},
						{ProtoField: "category"},
						{ProtoField: "brand"},
						{ProtoField: "model"},
						{ProtoField: "serial_hash"},
						{ProtoField: "condition"},
						{ProtoField: "image_uri"},
						{ProtoField: "image_sha256"},
						{ProtoField: "image_seed"},
					},
				},
				{
					RpcMethod: "TransferCertificate",
					Use:       "transfer-certificate [certificate-id] [new-owner]",
					Short:     "Transfere a propriedade de um certificado",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "certificate_id"},
						{ProtoField: "new_owner"},
					},
				},
				{
					RpcMethod: "AddServiceRecord",
					Use:       "add-service-record [certificate-id] [kind] [details]",
					Short:     "Adiciona registro de manutenção/upgrade",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "certificate_id"},
						{ProtoField: "kind"},
						{ProtoField: "details"},
					},
				},
				{
					RpcMethod: "RevokeCertificate",
					Use:       "revoke-certificate [certificate-id] [reason]",
					Short:     "Revoga um certificado",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "certificate_id"},
						{ProtoField: "reason"},
					},
				},
			},
		},
	}
}

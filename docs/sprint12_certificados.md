# Sprint 12 — RWA LITE: Certificado Digital de Proveniência (CDP) (backend/on-chain)

Objetivo: implementar o módulo `x/certificados` (Cosmos SDK) para emissão e gestão de “Certificados” NFT-like (não fracionados) para itens físicos (notebooks, periféricos gamers etc.). Nesta sprint **não há alterações de UX/frontend**.

## Visão geral

- Módulo Cosmos SDK: `x/certificados` (`ModuleName = "certificados"`)
- Entidade principal: `Certificate` (NFT-like, não fracionado)
- Taxa fixa por emissão: **1499ubyx** (0,001499 BYX em base `ubyx`)
- Imagem associada ao certificado:
  - Nesta sprint: **pipeline off-chain** determinística para gerar PNG “3D-like”
  - On-chain armazena: `image_uri`, `image_sha256`, `image_seed`
  - Não enviamos bytes da imagem on-chain
- Integração com `x/lojas`:
  - **Somente o merchant owner** (`merchant.creator`) pode emitir certificados vinculados a `merchant_id`
  - Emissão cobra taxa do emissor (merchant owner) e credita na conta do módulo `certificados`
- Política: **certificado revogado não pode ser transferido** (tx falha)

## Modelo de dados (proto)

Arquivo: `proto/byx/certificados/v1/certificate.proto`

Campos principais:
- `id` (uint64): ID incremental
- `merchant_id` (uint64): vínculo com `x/lojas`
- `issuer` (string): address bech32 do emissor (merchant owner)
- `owner` (string): owner atual
- `category` (string): `NOTEBOOK|MOUSE|KEYBOARD|HEADSET|MONITOR|GPU|OTHER`
- `brand`, `model` (string)
- `serial_hash` (string): `sha256(serial_number_normalized)` (hex). Não armazenar serial em plaintext.
- `condition` (string): ex. `"A+"`, `"A"`, `"B"`, `"C"`, com notas
- `notes` (string): opcional
- `image_uri` (string): ex `file://...`, `ipfs://...`, `https://...`
- `image_sha256` (string): sha256 hex do PNG gerado
- `image_seed` (string): seed determinística usada na geração
- `revoked` (bool), `revoked_reason` (string)
- `created_at` (string): RFC3339 UTC

Service records:
- `ServiceRecord` + contagem por certificado (`ServiceCount`)
- `ServiceRecords` são armazenados on-chain por `(certificate_id, index)`

## Storage (collections)

Implementação: `x/certificados/keeper/keeper.go`

- `Certificates`: `id -> Certificate`
- `CertificateCount`: contador incremental
- Índices (KeySet):
  - `ByOwner`: `(owner, id) -> NoValue`
  - `ByMerchant`: `(merchant_id, id) -> NoValue`
  - `BySerialHash`: `(serial_hash, id) -> NoValue`
- `ServiceRecords`: `(certificate_id, index) -> ServiceRecord`
- `ServiceRecordsCount`: `certificate_id -> uint64`

## Mensagens (tx) e eventos

Arquivo: `proto/byx/certificados/v1/tx.proto`

Tx:
- `MsgIssueCertificate` (emite)
  - cobra taxa `issue_fee_byx` do emissor para a conta do módulo `certificados`
  - evento: `certificados_issue` (`id`, `merchant_id`, `owner`, `category`, `image_sha256`)
- `MsgTransferCertificate` (transferência)
  - falha se `revoked=true`
  - evento: `certificados_transfer` (`id`, `from`, `to`)
- `MsgAddServiceRecord` (manutenção/upgrade)
  - permitido para `owner` ou `issuer` (política atual)
  - falha se `revoked=true`
  - evento: `certificados_service` (`id`, `added_by`, `type`)
- `MsgRevokeCertificate` (revogar)
  - permitido para `issuer` ou `merchant owner`
  - evento: `certificados_revoke` (`id`, `reason`)

## Queries gRPC/REST

Arquivo: `proto/byx/certificados/v1/query.proto`

Endpoints REST (via grpc-gateway):
- `GET /byx/certificados/v1/certificates/{id}`
- `GET /byx/certificados/v1/owners/{owner}/certificates`
- `GET /byx/certificados/v1/merchants/{merchant_id}/certificates`
- `GET /byx/certificados/v1/serial/{serial_hash}/certificates`
- `GET /byx/certificados/v1/params`

## Parâmetros (params)

Arquivo: `proto/byx/certificados/v1/params.proto` e `x/certificados/types/params.go`

- `enabled` (default `true`)
- `issue_fee_byx` (default `1499`)
- `allow_transfer` (default `true`)

## Ferramenta de geração de imagem (Sprint 12)

Pasta: `tools/cdp_image_gen/`

Instalação:
```bash
cd tools/cdp_image_gen
npm install
```

Geração (determinística por seed):
```bash
node tools/cdp_image_gen/generate.js --json '{
  "category":"NOTEBOOK",
  "brand":"Dell",
  "model":"XPS 15",
  "serial_hash":"<sha256hex>",
  "seed":"seed-1"
}'
```

Saída: JSON com `image_uri`, `image_sha256`, `image_seed`, `out_path`.

Nota: a renderização é determinística no mesmo ambiente (fontes/render pode variar entre OSs).

## Exemplos de CLI

Emissão:
```bash
byxd tx certificados issue-certificate 1 NOTEBOOK Dell XPS15 <serial_hash> A file:///tmp/cdp.png <image_sha256> seed-1 \
  --from marcelo --fees 20000ubyx --chain-id byx --node tcp://0.0.0.0:26657 --yes
```

Consulta:
```bash
byxd query certificados certificate 1 -o json
byxd query certificados certificates-by-owner <bech32> -o json
```

## Exemplo de curl (REST)

```bash
curl -s "http://127.0.0.1:1317/byx/certificados/v1/certificates/1" | jq .
```

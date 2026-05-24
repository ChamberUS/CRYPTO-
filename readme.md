# BYX / Buynnex

BYX is the blockchain protocol of Buynnex, focused on payment rails for merchants, programmable settlement flows, and operational tooling for private/public network evolution.

This repository contains the chain codebase (`byxd`) and application modules used by the BYX network.

Monetary unit convention:

- base on-chain denom: `ubyx`
- human display denom: `BYX`
- conversion: `1 BYX = 1_000_000 ubyx`

---

## Project status

- Recommended now: local development, internal testing, and private devnet/testnet.
- Not recommended now: open public production usage without governance/security checklist completion.

See operational docs in `docs/` for security baseline and private devnet deployment.

---

## Core principles

- **Security by default**: conservative defaults (faucet off unless explicitly enabled, local-only RPC/API by default in deploy scripts).
- **Deterministic economics**: fixed supply model target with explicit reserve usage.
- **Auditable flows**: payment request lifecycle, event traces, and module-level validations.
- **Operational rigor**: reproducible deploy scripts, config validation, and CI checks.

---

## Architecture overview

Main modules:

- `x/lojas`: merchant registry, cashback/faucet controls, transfer and sales anti-fraud limits.
- `x/payments`: payment request creation, payment settlement, idempotency controls.
- `x/certificados`: certificate lifecycle and transfer integration with payment flows.
- `x/feesplit`: fee split logic (validators / treasury / burn policies).

App entrypoints:

- daemon: `cmd/byxd`
- app wiring: `app/`

---

## Quick start (developer)

### Requirements

- Go (compatible with `go.mod`)
- `ignite` CLI (for `proto-gen`)
- `jq`, `curl`
- Linux/macOS shell environment

### Build and tests

```bash
make test-unit
make lint
make govulncheck
```

### Proto generation

```bash
make proto-gen
```

### Local/private genesis bootstrap

```bash
./scripts/genesis_private_devnet.sh
byxd genesis validate genesis.json
```

---

## Generic Linux deploy (Ubuntu/Debian, any VPS provider)

Provider-agnostic deployment toolkit:

- `scripts/deploy/.env.example`
- `scripts/deploy/bootstrap.sh`
- `scripts/deploy/configure_node.sh`
- `scripts/deploy/validate_node_config.sh`
- `scripts/deploy/install_systemd.sh`
- `scripts/deploy/healthcheck.sh`
- `scripts/deploy/backup_snapshot.sh`

Full guide:

- `docs/deploy_linux_generic.md`

Key default posture:

- RPC/API bound to localhost unless explicitly changed.
- P2P configurable through `.env`.
- TOML-safe config updates + post-update validation.

---

## Contribution guide

We welcome serious contributions. Please follow this process:

1. Open an Issue with:
   - problem statement
   - expected behavior
   - scope boundaries
2. Fork and create a focused branch:
   - `feat/<topic>` or `fix/<topic>`
3. Keep PRs small and reviewable.
4. Add/adjust tests for behavioral changes.
5. Run mandatory checks before PR:
   - `make test-unit`
   - `make lint`
   - `make govulncheck`
6. Include migration notes if API/proto/state changes are introduced.

### Coding standards

- Prefer root-cause fixes over superficial patches.
- Do not introduce secrets into code, docs, or examples.
- Keep module boundaries explicit; avoid hidden cross-module side effects.
- Preserve compatibility or document breaking changes clearly.

### Security expectations for contributors

- Never commit private keys, validator keys, `.env` secrets, or local chain state.
- Follow `.gitignore` and deployment hardening docs.
- Treat webhook/faucet/admin tooling as sensitive surfaces.

---

## Governance and economics notes

- BYX supply target: fixed cap model with controlled reserves.
- Merchant `saldo` must be updated only through authorized economic flows.
- Faucet and operational admin paths must remain explicitly authorized.

---

## Roadmap orientation

Near-term engineering priorities:

- strengthen fixed-supply invariants and validations;
- continue hardening private devnet operations;
- complete public-facing protocol/docs cleanup before open testnet scale.

---

## License

Define project license in a dedicated `LICENSE` file (recommended for external contributors).

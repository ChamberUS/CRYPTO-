#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENV_FILE="${ENV_FILE:-${SCRIPT_DIR}/.env}"
DRY_RUN="${DRY_RUN:-false}"

if [ ! -f "${ENV_FILE}" ]; then
  echo "env file not found: ${ENV_FILE}" >&2
  exit 1
fi

# shellcheck source=/dev/null
source "${ENV_FILE}"

CONFIG_TOML="${BYX_HOME}/config/config.toml"
APP_TOML="${BYX_HOME}/config/app.toml"

if [ ! -f "${CONFIG_TOML}" ] || [ ! -f "${APP_TOML}" ]; then
  echo "missing config files. run bootstrap first." >&2
  exit 1
fi

export CONFIG_TOML APP_TOML DRY_RUN
export BYX_RPC_LADDR BYX_P2P_LADDR BYX_EXTERNAL_ADDRESS BYX_PERSISTENT_PEERS BYX_SEEDS
export BYX_PEX BYX_PROMETHEUS BYX_PROMETHEUS_LISTEN BYX_PPROF_LADDR
export BYX_MIN_GAS_PRICES BYX_API_ENABLE BYX_API_ADDRESS BYX_GRPC_ENABLE BYX_GRPC_ADDRESS
export BYX_PRUNING BYX_PRUNING_KEEP_RECENT BYX_PRUNING_KEEP_EVERY BYX_PRUNING_INTERVAL

python3 <<'PY'
import os
import re
import shutil
import tempfile
from pathlib import Path

DRY_RUN = os.environ.get("DRY_RUN", "false").lower() == "true"

def to_toml_bool(value: str) -> str:
    v = value.strip().lower()
    if v not in {"true", "false"}:
        raise ValueError(f"invalid TOML bool: {value}")
    return v

def update_toml(path: Path, updates: dict[tuple[str | None, str], tuple[str, bool, bool]]) -> None:
    lines = path.read_text(encoding="utf-8").splitlines(keepends=True)
    current_section = None
    key_re = re.compile(r'^(\s*)([A-Za-z0-9_.-]+)(\s*=\s*)(.*?)(\s*(#.*)?)$')
    section_re = re.compile(r'^\s*\[([^\]]+)\]\s*$')
    found = set()
    out = []

    for line in lines:
        section_match = section_re.match(line.rstrip("\n"))
        if section_match:
            current_section = section_match.group(1).strip()
            out.append(line)
            continue

        key_match = key_re.match(line.rstrip("\n"))
        if key_match:
            key = key_match.group(2)
            composite = (current_section, key)
            if composite in updates:
                value, quoted, _required = updates[composite]
                rendered = f'"{value}"' if quoted else value
                newline = f"{key_match.group(1)}{key}{key_match.group(3)}{rendered}{key_match.group(5)}\n"
                out.append(newline)
                found.add(composite)
                continue
        out.append(line)

    missing_required = [item for item, (_v, _q, req) in updates.items() if req and item not in found]
    if missing_required:
        miss_fmt = ", ".join([f"[{s}] {k}" if s else k for s, k in missing_required])
        raise RuntimeError(f"required keys not found in {path}: {miss_fmt}")

    if DRY_RUN:
        print(f"[dry-run] validated changes for {path}")
        return

    with tempfile.NamedTemporaryFile("w", delete=False, encoding="utf-8", dir=str(path.parent)) as tmp:
        tmp.writelines(out)
        tmp_path = Path(tmp.name)

    shutil.copymode(path, tmp_path)
    tmp_path.replace(path)

config_path = Path(os.environ["CONFIG_TOML"])
app_path = Path(os.environ["APP_TOML"])

config_updates = {
    ("rpc", "laddr"): (os.environ["BYX_RPC_LADDR"], True, True),
    ("p2p", "laddr"): (os.environ["BYX_P2P_LADDR"], True, True),
    ("p2p", "external_address"): (os.environ["BYX_EXTERNAL_ADDRESS"], True, True),
    ("p2p", "persistent_peers"): (os.environ["BYX_PERSISTENT_PEERS"], True, True),
    ("p2p", "seeds"): (os.environ["BYX_SEEDS"], True, True),
    ("p2p", "pex"): (to_toml_bool(os.environ["BYX_PEX"]), False, True),
    ("instrumentation", "prometheus"): (to_toml_bool(os.environ["BYX_PROMETHEUS"]), False, True),
    ("instrumentation", "prometheus_listen_addr"): (os.environ["BYX_PROMETHEUS_LISTEN"], True, True),
    ("rpc", "pprof_laddr"): (os.environ["BYX_PPROF_LADDR"], True, True),
}

app_updates = {
    (None, "minimum-gas-prices"): (os.environ["BYX_MIN_GAS_PRICES"], True, True),
    ("api", "enable"): (to_toml_bool(os.environ["BYX_API_ENABLE"]), False, True),
    ("api", "address"): (os.environ["BYX_API_ADDRESS"], True, True),
    ("grpc", "enable"): (to_toml_bool(os.environ["BYX_GRPC_ENABLE"]), False, True),
    ("grpc", "address"): (os.environ["BYX_GRPC_ADDRESS"], True, True),
    (None, "pruning"): (os.environ["BYX_PRUNING"], True, True),
    (None, "pruning-keep-recent"): (os.environ["BYX_PRUNING_KEEP_RECENT"], True, True),
    (None, "pruning-keep-every"): (os.environ["BYX_PRUNING_KEEP_EVERY"], True, False),
    (None, "pruning-interval"): (os.environ["BYX_PRUNING_INTERVAL"], True, True),
}

update_toml(config_path, config_updates)
update_toml(app_path, app_updates)
print("node configuration updated" + (" (dry-run)" if DRY_RUN else ""))
PY

if [ "${DRY_RUN}" != "true" ]; then
  ENV_FILE="${ENV_FILE}" "${SCRIPT_DIR}/validate_node_config.sh"
fi

import { execFile } from "node:child_process";
import { randomUUID } from "node:crypto";
import express from "express";

const REQUIRED_ENV = ["NODE", "CHAIN_ID", "BYXD_HOME", "KEYRING_BACKEND"];
const config = {
  port: envNumber("PORT", 8080),
  node: requiredEnv("NODE"),
  chainId: requiredEnv("CHAIN_ID"),
  byxdHome: requiredEnv("BYXD_HOME"),
  keyringBackend: requiredEnv("KEYRING_BACKEND"),
  byxdBin: process.env.BYXD_BIN || "byxd",
  allowProductionRun: process.env.BYX_BACKEND_ALLOW_PRODUCTION === "true",
  apiToken: process.env.BYX_BACKEND_API_TOKEN || "",
  allowUnauthenticated: process.env.BYX_ALLOW_UNAUTHENTICATED === "true",
  createPaymentKey: process.env.BYX_CREATE_PAYMENT_KEY || "",
  devnetPayerKey: process.env.BYX_DEVNET_PAYER_KEY || "",
  txFees: process.env.BYX_TX_FEES || "5000byx",
  txGas: process.env.BYX_TX_GAS || "auto",
  txGasAdjustment: process.env.BYX_TX_GAS_ADJUSTMENT || "1.3",
  maxPaymentMicrobyx: envBigInt("BYX_MAX_PAYMENT_MICROBYX", 1_000_000_000_000n),
  cliTimeoutMs: envNumber("BYX_CLI_TIMEOUT_MS", 20_000),
  txWaitMs: envNumber("BYX_TX_WAIT_MS", 15_000),
};

for (const name of REQUIRED_ENV) {
  if (!process.env[name]) {
    throw new Error(`missing required env ${name}`);
  }
}
if (process.env.NODE_ENV === "production" && !config.allowProductionRun) {
  throw new Error("refusing to run with NODE_ENV=production unless BYX_BACKEND_ALLOW_PRODUCTION=true");
}

const app = express();
app.disable("x-powered-by");
app.use(express.json({ limit: "32kb" }));

app.use((req, res, next) => {
  req.requestId = req.header("x-request-id") || randomUUID();
  res.setHeader("x-request-id", req.requestId);
  next();
});

app.use((req, _res, next) => {
  if (req.path === "/v1/devnet/health") {
    next();
    return;
  }
  if (config.allowUnauthenticated) {
    next();
    return;
  }
  if (!config.apiToken) {
    next(httpError(503, "BYX_BACKEND_API_TOKEN is required unless BYX_ALLOW_UNAUTHENTICATED=true"));
    return;
  }
  if (!hasBearerAuth(req)) {
    next(httpError(401, "unauthorized"));
    return;
  }
  next();
});

app.get("/v1/devnet/health", asyncHandler(async () => {
  const status = await byxdJSON(["status"]);
  return {
    environment: "DEVNET_TESTE_FECHADO",
    chain_id: config.chainId,
    node: config.node,
    latest_block_height: status?.sync_info?.latest_block_height || null,
    catching_up: status?.sync_info?.catching_up ?? null,
    node_info: {
      network: status?.node_info?.network || null,
      moniker: status?.node_info?.moniker || null,
    },
  };
}));

app.get("/v1/devnet/merchants/:id", asyncHandler(async (req) => {
  const id = parsePositiveUint(req.params.id, "id");
  return byxdJSON(["query", "lojas", "merchant", id.toString()]);
}));

app.get("/v1/devnet/merchants/:id/saldo", asyncHandler(async (req) => {
  const id = parsePositiveUint(req.params.id, "id");
  const data = await byxdJSON(["query", "lojas", "merchant", id.toString()]);
  const merchant = data?.merchant || data?.Merchant || data;
  return {
    environment: "DEVNET_TESTE_FECHADO",
    loja_id: id,
    saldo: merchant?.saldo || merchant?.Saldo || "0",
    merchant,
  };
}));

app.post("/v1/devnet/payment-requests", asyncHandler(async (req) => {
  requireTxKey(config.createPaymentKey, "BYX_CREATE_PAYMENT_KEY");
  const lojaId = parsePositiveUint(req.body?.loja_id, "loja_id");
  const amountMicrobyx = parsePositiveUint(req.body?.amount_microbyx, "amount_microbyx");
  if (BigInt(amountMicrobyx) > config.maxPaymentMicrobyx) {
    throw httpError(400, "amount_microbyx exceeds BYX_MAX_PAYMENT_MICROBYX");
  }

  const memo = parseOptionalString(req.body?.memo, "memo", 140);
  const expiresInSeconds = parseOptionalUint(req.body?.expires_in_seconds, "expires_in_seconds");

  const args = [
    "tx", "payments", "create-payment-request",
    lojaId.toString(),
    amountMicrobyx.toString(),
    ...txArgs(config.createPaymentKey),
  ];
  if (memo) args.push("--memo", memo);
  if (expiresInSeconds !== undefined) args.push("--expires-in-seconds", expiresInSeconds.toString());

  const tx = await broadcastTx(args);
  const txhash = extractTxHash(tx);
  const status = txStatus(tx);
  const requestId = extractEventAttribute(tx, "byx_payment_request_created", "request_id");
  logInfo("payment_request_created", { txhash, request_id: requestId, loja_id: lojaId, status });

  return {
    environment: "DEVNET_TESTE_FECHADO",
    status,
    txhash,
    request_id: requestId,
    loja_id: lojaId,
    raw: tx,
  };
}));

app.get("/v1/devnet/payment-requests/:id", asyncHandler(async (req) => {
  const id = parsePositiveUint(req.params.id, "id");
  return byxdJSON(["query", "payments", "payment-request", id.toString()]);
}));

app.get("/v1/devnet/payment-requests/:id/qr", asyncHandler(async (req) => {
  const id = parsePositiveUint(req.params.id, "id");
  const data = await byxdJSON(["query", "payments", "payments-qr", id.toString()]);
  return {
    environment: "DEVNET_TESTE_FECHADO",
    ...data,
  };
}));

app.get("/v1/devnet/game/petz/balance", asyncHandler(async (req) => {
  const wallet = parseOptionalString(req.query?.wallet, "wallet", 128);
  return {
    environment: "DEVNET_TESTE_FECHADO",
    feature: "game.petz.balance",
    status: "placeholder",
    wallet: wallet || null,
    balances: [],
    message: "Petz balance placeholder for closed devnet app testing.",
  };
}));

app.post("/v1/devnet/game/petz/reward", requireBearerRoute, asyncHandler(async (req) => {
  const wallet = parseOptionalString(req.body?.wallet, "wallet", 128);
  const rewardId = parseOptionalString(req.body?.reward_id, "reward_id", 64);
  logInfo("petz_reward_placeholder", {
    request_id: req.requestId,
    wallet: wallet || null,
    reward_id: rewardId || null,
    status: "blocked_placeholder",
  });
  throw httpError(501, "Petz reward is a DEVNET placeholder and is not enabled");
}));

app.post("/v1/devnet/payment-requests/:id/pay", asyncHandler(async (req) => {
  requireTxKey(config.devnetPayerKey, "BYX_DEVNET_PAYER_KEY");
  const id = parsePositiveUint(req.params.id, "id");
  const before = await byxdJSON(["query", "payments", "payment-request", id.toString()]);
  const paymentRequest = before?.payment_request || before?.paymentRequest || before;
  const lojaId = Number(paymentRequest?.loja_id || paymentRequest?.lojaId || 0) || null;

  const tx = await broadcastTx([
    "tx", "payments", "pay-payment-request",
    id.toString(),
    ...txArgs(config.devnetPayerKey),
  ]);
  const txhash = extractTxHash(tx);
  const status = txStatus(tx);
  logInfo("payment_request_paid", { txhash, request_id: id, loja_id: lojaId, status });

  return {
    environment: "DEVNET_TESTE_FECHADO",
    status,
    txhash,
    request_id: id,
    loja_id: lojaId,
    raw: tx,
  };
}));

app.use((err, req, res, _next) => {
  const status = err.status || 500;
  logError("request_failed", {
    request_id: req.requestId,
    method: req.method,
    path: req.path,
    status,
    error: err.message,
  });
  res.status(status).json({
    environment: "DEVNET_TESTE_FECHADO",
    error: err.message,
    request_id: req.requestId,
  });
});

app.listen(config.port, () => {
  logInfo("server_started", {
    environment: "DEVNET_TESTE_FECHADO",
    port: config.port,
    chain_id: config.chainId,
    node: config.node,
    keyring_backend: config.keyringBackend,
    auth: config.allowUnauthenticated ? "disabled" : "bearer",
    production_run: config.allowProductionRun,
  });
});

function txArgs(fromKey) {
  return [
    "--from", fromKey,
    "--chain-id", config.chainId,
    "--node", config.node,
    "--home", config.byxdHome,
    "--keyring-backend", config.keyringBackend,
    "--fees", config.txFees,
    "--gas", config.txGas,
    "--gas-adjustment", config.txGasAdjustment,
    "--yes",
    "--output", "json",
  ];
}

function queryArgs() {
  return [
    "--node", config.node,
    "--home", config.byxdHome,
    "--output", "json",
  ];
}

function byxdJSON(args) {
  const fullArgs = args[0] === "tx" ? args : [...args, ...queryArgs()];
  return new Promise((resolve, reject) => {
    execFile(config.byxdBin, fullArgs, {
      timeout: config.cliTimeoutMs,
      maxBuffer: 1024 * 1024,
      env: process.env,
    }, (error, stdout, stderr) => {
      if (error) {
        reject(httpError(502, cleanCliError(stderr || stdout || error.message)));
        return;
      }
      try {
        resolve(JSON.parse(stdout || "{}"));
      } catch {
        reject(httpError(502, "byxd returned non-JSON output"));
      }
    });
  });
}

async function broadcastTx(args) {
  const tx = await byxdJSON(args);
  const txhash = extractTxHash(tx);
  if (!txhash || config.txWaitMs === 0) return tx;

  const receipt = await waitForTx(txhash);
  return receipt || tx;
}

async function waitForTx(txhash) {
  const deadline = Date.now() + config.txWaitMs;
  let lastError = null;
  while (Date.now() <= deadline) {
    try {
      return await byxdJSON(["query", "tx", txhash]);
    } catch (err) {
      lastError = err;
      await sleep(1000);
    }
  }
  logError("tx_receipt_timeout", { txhash, error: lastError?.message || "timeout" });
  return null;
}

function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

function asyncHandler(handler) {
  return async (req, res, next) => {
    try {
      const result = await handler(req, res);
      if (!res.headersSent) res.json(result);
    } catch (err) {
      next(err);
    }
  };
}

function requiredEnv(name) {
  return process.env[name] || "";
}

function envNumber(name, fallback) {
  const value = process.env[name];
  if (!value) return fallback;
  const parsed = Number(value);
  if (!Number.isInteger(parsed) || parsed <= 0) {
    throw new Error(`${name} must be a positive integer`);
  }
  return parsed;
}

function envBigInt(name, fallback) {
  const value = process.env[name];
  if (!value) return fallback;
  if (!/^[1-9][0-9]*$/.test(value)) {
    throw new Error(`${name} must be a positive integer`);
  }
  return BigInt(value);
}

function parsePositiveUint(value, field) {
  const normalized = typeof value === "number" ? String(value) : value;
  if (typeof normalized !== "string" || !/^[1-9][0-9]*$/.test(normalized)) {
    throw httpError(400, `${field} must be a positive integer`);
  }
  const parsed = Number(normalized);
  if (!Number.isSafeInteger(parsed)) {
    throw httpError(400, `${field} exceeds safe integer range`);
  }
  return parsed;
}

function parseOptionalUint(value, field) {
  if (value === undefined || value === null || value === "") return undefined;
  return parsePositiveUint(value, field);
}

function parseOptionalString(value, field, maxLength) {
  if (value === undefined || value === null) return "";
  if (typeof value !== "string") {
    throw httpError(400, `${field} must be a string`);
  }
  const trimmed = value.trim();
  if (trimmed.length > maxLength) {
    throw httpError(400, `${field} must be at most ${maxLength} chars`);
  }
  return trimmed;
}

function requireBearerRoute(req, _res, next) {
  if (!config.apiToken) {
    next(httpError(503, "BYX_BACKEND_API_TOKEN is required for this endpoint"));
    return;
  }
  if (!hasBearerAuth(req)) {
    next(httpError(401, "unauthorized"));
    return;
  }
  next();
}

function hasBearerAuth(req) {
  return req.header("authorization") === `Bearer ${config.apiToken}`;
}

function requireTxKey(value, envName) {
  if (!value) {
    throw httpError(503, `${envName} is required for this DEVNET endpoint`);
  }
}

function txStatus(tx) {
  const code = Number(tx?.code || 0);
  return code === 0 ? "success" : "failed";
}

function extractTxHash(tx) {
  return tx?.txhash || tx?.tx_response?.txhash || tx?.txResponse?.txhash || null;
}

function extractEventAttribute(tx, eventType, attrKey) {
  const events = [
    ...(tx?.events || []),
    ...(tx?.logs || []).flatMap((log) => log.events || []),
    ...(tx?.tx_response?.events || []),
    ...(tx?.txResponse?.events || []),
  ];
  for (const event of events) {
    if (event.type !== eventType) continue;
    for (const attr of event.attributes || []) {
      const key = decodeMaybeBase64(attr.key);
      if (key === attrKey) return decodeMaybeBase64(attr.value);
    }
  }
  return null;
}

function decodeMaybeBase64(value) {
  if (typeof value !== "string") return value;
  if (!/^[A-Za-z0-9+/]+={0,2}$/.test(value)) return value;
  try {
    const decoded = Buffer.from(value, "base64").toString("utf8");
    return /^[\x20-\x7E]+$/.test(decoded) ? decoded : value;
  } catch {
    return value;
  }
}

function cleanCliError(message) {
  return String(message).replace(/\s+/g, " ").trim().slice(0, 500);
}

function httpError(status, message) {
  const err = new Error(message);
  err.status = status;
  return err;
}

function logInfo(event, fields) {
  console.log(JSON.stringify({ level: "info", event, time: new Date().toISOString(), ...fields }));
}

function logError(event, fields) {
  console.error(JSON.stringify({ level: "error", event, time: new Date().toISOString(), ...fields }));
}

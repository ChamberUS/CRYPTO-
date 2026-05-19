import express from "express";
import { createProxyMiddleware } from "http-proxy-middleware";
import https from "https";
import selfsigned from "selfsigned";

const TARGET = process.env.WEBHOOK_TARGET || "http://127.0.0.1:3001";
const PORT = Number(process.env.WEBHOOK_PROXY_PORT || 3443);

const attrs = [{ name: "commonName", value: "localhost" }];
const pems = selfsigned.generate(attrs, { days: 365, keySize: 2048 });

const app = express();

app.use(
  "/",
  createProxyMiddleware({
    target: TARGET,
    changeOrigin: true,
    secure: false,
    logLevel: "info",
  })
);

const server = https.createServer(
  {
    key: pems.private,
    cert: pems.cert,
  },
  app
);

server.listen(PORT, () => {
  console.log(`[WEBHOOK-PROXY] HTTPS listening on https://127.0.0.1:${PORT} -> ${TARGET}`);
  console.log(`[WEBHOOK-PROXY] Set MERCHANT_WEBHOOK_URL=https://127.0.0.1:${PORT}/webhook`);
});

// server.js (CommonJS)
const express = require('express');
const cors = require('cors');
const path = require('path');
const { execFile } = require('child_process');
const util = require('util');
require('dotenv').config();

const execFileP = util.promisify(execFile);
const app = express();

// estáticos
const publicDir = path.join(__dirname, 'public');
app.use(express.static(publicDir));
app.get('/', (_req, res) => res.sendFile(path.join(publicDir, 'index.html')));

// env
const {
  PORT = 3000,
  NODE_RPC = 'tcp://127.0.0.1:26657',
  CHAIN_ID = 'byx',
  ADMIN_KEY = 'marcelo',
  KEYRING = 'test',
  FEES = '25000ubyx',
  CORS_ORIGIN = 'http://localhost:3000',
  ENABLE_WEB_FAUCET = 'false',
} = process.env;

if (ENABLE_WEB_FAUCET !== 'true') {
  throw new Error('web-faucet desabilitado. Defina ENABLE_WEB_FAUCET=true apenas em ambiente de desenvolvimento.');
}

if (process.env.NODE_ENV === 'production') {
  if (KEYRING === 'test') {
    throw new Error('KEYRING=test não é permitido em produção.');
  }
  if (ADMIN_KEY === 'marcelo') {
    throw new Error('ADMIN_KEY padrão não é permitido em produção.');
  }
}

app.use(cors({ origin: CORS_ORIGIN }));
app.use(express.json());

const okStderr = (s) => !s || /gas estimate|I\[/.test(s);
const safeJSON  = (s) => { try { return JSON.parse(s); } catch { return s; } };

// Faucet: usa --lojista-id e --amount
app.post('/api/faucet', async (req, res) => {
  try {
    const { id, amount } = req.body || {};
    if (!id || !amount) return res.json({ ok:false, error:'id/amount required' });

    const args = [
      'tx', 'lojas', 'faucet',
      '--lojista-id', String(id),
      '--amount',     String(amount),
      '--from', ADMIN_KEY,
      '--keyring-backend', KEYRING,
      '--chain-id', CHAIN_ID,
      '--node', NODE_RPC,
      '--broadcast-mode','sync',
      '--gas','auto','--fees', FEES,
      '-y','-o','json'
    ];

    const { stdout, stderr } = await execFileP('byxd', args, { timeout: 20000 });
    if (!okStderr(stderr)) return res.json({ ok:false, error: stderr.trim() });
    return res.json({ ok:true, tx: safeJSON(stdout) });
  } catch (e) {
    return res.json({ ok:false, error: String(e.message || e) });
  }
});

// Transfer (como você já validou)
app.post('/api/transfer', async (req, res) => {
  try {
    const { fromId, toId, valor } = req.body || {};
    if (fromId == null || toId == null || !valor) {
      return res.json({ ok:false, error:'fromId/toId/valor required' });
    }
    const args = [
      'tx','lojas','transferir-byx',
      String(fromId), String(toId), String(valor), // seu CLI aceita posicionais aqui
      '--from', ADMIN_KEY,
      '--keyring-backend', KEYRING,
      '--chain-id', CHAIN_ID,
      '--node', NODE_RPC,
      '--broadcast-mode','sync',
      '--gas','auto','--fees', FEES,
      '-y','-o','json'
    ];
    const { stdout, stderr } = await execFileP('byxd', args, { timeout: 20000 });
    if (!okStderr(stderr)) return res.json({ ok:false, error: stderr.trim() });
    return res.json({ ok:true, tx: safeJSON(stdout) });
  } catch (e) {
    return res.json({ ok:false, error: String(e.message || e) });
  }
});

app.post('/api/transferir', (req, res, next) => {
  // alias simples
  req.url = '/api/transfer'; next();
});

// Consulta simples via CLI (dev)
app.get('/api/merchant/:id', async (req, res) => {
  try {
    const args = ['q','lojas','get-merchant', String(req.params.id), '--node', NODE_RPC, '-o','json'];
    const { stdout } = await execFileP('byxd', args, { timeout: 20000 });
    return res.json(safeJSON(stdout));
  } catch (e) {
    return res.status(500).json({ ok:false, error:String(e.message || e) });
  }
});

app.listen(PORT, () => {
  console.log(`BYX rodando em http://localhost:${PORT}`);
  console.log(`Servindo estáticos de: ${publicDir}`);
});

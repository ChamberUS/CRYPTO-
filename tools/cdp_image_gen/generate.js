import { createCanvas } from "canvas";
import fs from "fs";
import crypto from "crypto";
import path from "path";

const args = process.argv.slice(2);
const jsonIndex = args.indexOf("--json");

if (jsonIndex === -1 || !args[jsonIndex + 1]) {
  console.error("Uso: node generate.js --json '{...}'");
  process.exit(1);
}

const input = JSON.parse(args[jsonIndex + 1]);

const {
  category = "ITEM",
  brand = "UNKNOWN",
  model = "MODEL",
  serial_hash,
  seed = "seed",
  condition = "A"
} = input;

if (!serial_hash) {
  console.error("serial_hash é obrigatório");
  process.exit(1);
}

const width = 800;
const height = 800;
const canvas = createCanvas(width, height);
const ctx = canvas.getContext("2d");

// background gradiente
const grad = ctx.createLinearGradient(0, 0, width, height);
grad.addColorStop(0, "#0f2027");
grad.addColorStop(0.5, "#203a43");
grad.addColorStop(1, "#2c5364");
ctx.fillStyle = grad;
ctx.fillRect(0, 0, width, height);

// card
ctx.fillStyle = "rgba(0,0,0,0.35)";
ctx.fillRect(80, 80, 640, 640);

// texto
ctx.fillStyle = "#ffffff";
ctx.font = "bold 36px Sans";
ctx.fillText("BYX CERTIFICATE", 200, 140);

ctx.font = "24px Sans";
ctx.fillText(`Category: ${category}`, 140, 220);
ctx.fillText(`Brand: ${brand}`, 140, 260);
ctx.fillText(`Model: ${model}`, 140, 300);
ctx.fillText(`Condition: ${condition}`, 140, 340);

ctx.font = "16px Monospace";
ctx.fillText(`Serial Hash: ${serial_hash.slice(0, 16)}...`, 140, 420);
ctx.fillText(`Seed: ${seed}`, 140, 460);

// saída
const buffer = canvas.toBuffer("image/png");
const hash = crypto.createHash("sha256").update(buffer).digest("hex");

const outDir = path.resolve("./out");
if (!fs.existsSync(outDir)) fs.mkdirSync(outDir);

const outPath = path.join(outDir, `${serial_hash}.png`);
fs.writeFileSync(outPath, buffer);

console.log(JSON.stringify({
  image_uri: `file://${outPath}`,
  image_sha256: hash
}, null, 2));

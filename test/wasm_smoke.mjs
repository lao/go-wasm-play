// End-to-end smoke test: loads the real json.wasm through wasm_exec.js and
// exercises the exported async (Promise) functions, byte transfer, and error
// propagation.
//
// Usage: node test/wasm_smoke.mjs [assets-dir]   (defaults to ../assets)
// Requires: GOOS=js GOARCH=wasm go build -o assets/json.wasm ./cmd/wasm
import fs from "node:fs";
import path from "node:path";
import zlib from "node:zlib";
import { fileURLToPath } from "node:url";
import { createRequire } from "node:module";

const here = path.dirname(fileURLToPath(import.meta.url));
const assets = path.resolve(process.argv[2] ?? path.join(here, "..", "assets"));
const require = createRequire(import.meta.url);

// wasm_exec.js is a classic script defining globalThis.Go.
globalThis.require = require;
globalThis.fs = fs;
globalThis.path = path;
require(path.join(assets, "wasm_exec.js"));

// --- build a tiny valid 2x2 RGB PNG so we can test the image pipeline ---
function chunk(type, data) {
  const len = Buffer.alloc(4);
  len.writeUInt32BE(data.length, 0);
  const typeBuf = Buffer.from(type, "ascii");
  const crc = Buffer.alloc(4);
  crc.writeUInt32BE(zlib.crc32(Buffer.concat([typeBuf, data])) >>> 0, 0);
  return Buffer.concat([len, typeBuf, data, crc]);
}
function makePNG() {
  const sig = Buffer.from([137, 80, 78, 71, 13, 10, 26, 10]);
  const ihdr = Buffer.alloc(13);
  ihdr.writeUInt32BE(2, 0); // width
  ihdr.writeUInt32BE(2, 4); // height
  ihdr[8] = 8;  // bit depth
  ihdr[9] = 2;  // colour type RGB
  // rows: filter byte 0 + 2 RGB pixels each
  const raw = Buffer.from([
    0, 255, 0, 0, 0, 255, 0,
    0, 0, 0, 255, 255, 255, 0,
  ]);
  const idat = zlib.deflateSync(raw);
  return Buffer.concat([sig, chunk("IHDR", ihdr), chunk("IDAT", idat), chunk("IEND", Buffer.alloc(0))]);
}

function assert(cond, msg) {
  if (!cond) throw new Error("ASSERT FAILED: " + msg);
  console.log("  ok -", msg);
}

const go = new Go();
const { instance } = await WebAssembly.instantiate(fs.readFileSync(path.join(assets, "json.wasm")), go.importObject);
go.run(instance); // never resolves (main blocks); globals are set synchronously

// Let the module finish initialising.
await new Promise((r) => setTimeout(r, 50));

try {
  // formatJSON (sync)
  const pretty = formatJSON('{"b":1,"a":2}');
  assert(pretty.includes("\n  "), "formatJSON indents output");

  // md5Hash (sync, backwards compatible)
  assert(md5Hash("abc") === "900150983cd24fb0d6963f7d28e17f72", "md5Hash matches reference");

  // capabilities advertised to the UI
  assert(wasmCapabilities.checksum.includes("sha256"), "capabilities list sha256");

  // checksum (async, byte transfer)
  const bytes = new TextEncoder().encode("abc");
  const sha = await checksum("sha256", bytes);
  assert(sha === "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad", "checksum sha256 of abc");

  // error propagation: bad algorithm rejects the promise
  let rejected = false;
  try { await checksum("nope", bytes); } catch (e) { rejected = true; }
  assert(rejected, "checksum rejects unsupported algorithm");

  // compression benchmark
  const blob = new TextEncoder().encode("repeat ".repeat(5000));
  const results = await benchmarkCompression(blob);
  assert(Array.isArray(results) && results.length === 3, "benchmarkCompression returns 3 results");
  assert(results.every((r) => r.compressedSize < r.originalSize), "all algorithms shrink the data");
  assert(typeof results[0].durationMs === "number", "results include timing");

  // single compress returns structured stats
  const one = await compress("gzip", blob);
  assert(one.algorithm === "gzip" && one.ratio > 0 && one.ratio < 1, "compress returns gzip stats");

  // image pipeline: info + resize + grayscale on a generated PNG
  const png = new Uint8Array(makePNG());
  const info = await imageInfo(png);
  assert(info.format === "png" && info.width === 2 && info.height === 2, "imageInfo reads 2x2 png");

  const resized = await resizeImage(png, 8, 8, "png");
  assert(resized instanceof Uint8Array && resized.length > 0, "resizeImage returns bytes");
  const resizedInfo = await imageInfo(resized);
  assert(resizedInfo.width === 8 && resizedInfo.height === 8, "resized image is 8x8");

  const gray = await grayscaleImage(png, "png");
  assert(gray instanceof Uint8Array && gray.length > 0, "grayscaleImage returns bytes");

  // resize of invalid data rejects
  let imgRejected = false;
  try { await resizeImage(new Uint8Array([1, 2, 3]), 4, 4, "png"); } catch (e) { imgRejected = true; }
  assert(imgRejected, "resizeImage rejects invalid image data");

  console.log("\nALL SMOKE TESTS PASSED");
  process.exit(0);
} catch (err) {
  console.error("\nSMOKE TEST FAILURE:", err.message);
  process.exit(1);
}

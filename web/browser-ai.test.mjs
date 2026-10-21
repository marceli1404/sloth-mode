import test from "node:test";
import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { BROWSER_GEMMA, formatBytes, checkBrowserAI } from "./browser-ai.mjs";

const html = readFileSync(new URL("./index.html", import.meta.url), "utf8");

test("model is a fixed, browser-compatible upstream release", () => {
  assert.equal(BROWSER_GEMMA.bytes, 2008432640);
  assert.match(BROWSER_GEMMA.modelURL, /\/resolve\/73d35ec\/gemma-4-E2B-it-web\.litertlm$/);
  assert.match(BROWSER_GEMMA.runtimeURL, /@litert-lm\/core@0\.18\.1\/\+esm$/);
  assert.equal(formatBytes(BROWSER_GEMMA.bytes), "2.01 GB");
});

test("browser panel is dashboard-only, and does not depend on account sign-in", () => {
  for (const id of ["sm-browser-ai", "sm-browser-start", "sm-browser-status", "sm-browser-input", "sm-browser-output", "sm-browser-plan", "sm-browser-stop", "sm-browser-progress", "sm-browser-remove"]) {
    assert.ok(html.includes('id="' + id + '"'), "missing UI element: " + id);
  }
  assert.match(html, /html\[data-sloth-page="home"\] #sm-browser-ai\{display:none!important\}/);
  assert.match(html, /<script type="module" src="\/browser-ai\.mjs"><\/script>/);
  assert.match(html, /No installation, account or pairing required/);
});

test("insecure pages cannot start browser inference", async () => {
  const result = await checkBrowserAI({ isSecureContext: false, navigator: {} });
  assert.equal(result.ok, false);
  assert.match(result.reason, /HTTPS/);
});

test("absence of WebGPU or private file storage has a useful fallback", async () => {
  const noGPU = await checkBrowserAI({ isSecureContext: true, navigator: { storage: { getDirectory() {} } } });
  assert.equal(noGPU.ok, false);
  assert.match(noGPU.reason, /desktop companion/);
  const noOPFS = await checkBrowserAI({ isSecureContext: true, navigator: { gpu: { requestAdapter() {} } } });
  assert.equal(noOPFS.ok, false);
  assert.match(noOPFS.reason, /OPFS/);
});

test("WebGPU adapter check handles success and refusal", async () => {
  const make = adapter => ({
    isSecureContext: true,
    navigator: {
      gpu: { requestAdapter: async () => adapter },
      storage: { getDirectory: async () => ({}) }
    }
  });
  assert.equal((await checkBrowserAI(make({}))).ok, true);
  assert.equal((await checkBrowserAI(make(null))).ok, false);
});

# Browser Gemma 4 E2B (experimental)

Sloth Mode's dashboard has an optional, account-free browser inference panel. The existing paired Windows/Linux companion and Gemma 3/llama.cpp path are untouched.

## Flow

1. Open the dashboard in an HTTPS, WebGPU-capable browser.
2. The page checks WebGPU adapter availability and Origin Private File System (OPFS) support. These checks do not fetch the model.
3. Click Download Gemma (~2.01 GB). The model downloads directly from the pinned upstream Hugging Face revision, with progress, cancellation and byte-length checks.
4. Model bytes are streamed into OPFS under the current website origin. Completed cached weights are reused, unless browser storage is cleared or evicted.
5. Click Make a quiet plan to stream responses from LiteRT-LM on the GPU. Stop cancels generation. Remove downloaded model deletes locally stored weights.

No Sloth Mode account, pairing token, connector installation or cloud inference API is required. Notes are not automatically synced or sent to Sloth Mode's backend. The Quiet Lab import button only copies textarea content inside the browser.

## Versioned upstream dependencies

- Google LiteRT-LM JavaScript 0.18.1 from jsDelivr (@litert-lm/core@0.18.1).
- Hugging Face litert-community/gemma-4-E2B-it-litert-lm, web variant at revision 73d35ec, 2,008,432,640 bytes.
- Upstream model SHA-256: 3a08e8d94e23b814ae5414469c370c503813949acb8ceaa17e4ebf8a35af35b5.

The current browser implementation verifies the byte count, not the full cryptographic digest. Browser SubtleCrypto does not offer an incremental hash API, and hashing a multi-GB Blob in one go risks high memory usage. A reviewed streaming digest implementation and self-hosted/version-pinned WASM assets are needed before claiming a supply-chain-verified, offline-ready release. HTTPS and a pinned revision help, but do not replace hash verification.

The web-specific LiteRT-LM model build requires WebGPU. Do not silently fall back to an untested CPU model. Browser model selection is not arbitrary GGUF quantisation; the native connector's hardware/quantisation workflow remains separate.

## Deployment checklist

This change modifies web/index.html and adds web/browser-ai.mjs. Publish both files to the same origin. The /browser-ai.mjs URL must return JavaScript with the correct MIME type rather than an HTML fallback.

- Serve the site over HTTPS.
- Make sure script-src and WASM-related CSP rules permit the version-pinned jsDelivr runtime and its module/asset loads. Avoid broad CSP relaxations.
- Confirm that the Hugging Face model fetch completes with suitable CORS headers and that OPFS has enough space.
- Test Chrome/Edge with WebGPU and enough GPU memory for roughly 2 GB weights plus runtime buffers. WebGPU existence alone does not prove inference works.
- Test download, cancel, corrupt partial cleanup, subsequent cache reuse, model deletion, stop generation, site storage exhaustion, browser reload and guest users.
- Confirm prompts never leave the machine by inspecting the Network panel. The expected external connections are for the opted-in model download and inference runtime assets.
- Re-test account task saving, native connector pairing, local notes and Quiet Lab separately.

## Known limits

Browser inference is beta: continuous integration cannot exercise a physical WebGPU device or the full multi-GB download. After the runtime and model are loaded into an open page, the model executes locally, but offline reopening is not guaranteed while JavaScript/WASM assets come from a CDN. Browser eviction can delete model files. This cannot monitor desktop processes, access arbitrary files or continue when the tab is closed.

If browser inference fails, keep using the existing Windows/Linux companion. Never automatically transmit user notes to another model provider.

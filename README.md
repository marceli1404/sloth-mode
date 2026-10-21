# 🦥 Sloth Mode

**Slow down. Get things done.** A local-first AI productivity assistant with optional cloud accounts.

[Website](https://sloth.marceli.dev) · [Try the workspace](https://sloth.marceli.dev/dashboard) · [Issue tracker](https://github.com/marceli1404/sloth-mode/issues)

Sloth Mode supports **optional browser-native Gemma 4 E2B with WebGPU** (beta, explicit ~2.01 GB download) and **Gemma 3 GGUF through `llama.cpp`** in the Windows/Linux companion for optional local inference. The companion runs in the background on Windows or Linux and pairs explicitly with the web interface through a loopback API. **Ollama and an account are not required.** The public Quiet Lab sample demo uses illustrative deterministic sorting, not actual Gemma inference.

## Browser AI beta (no installation needed)

Open [the dashboard](https://sloth.marceli.dev/dashboard) and use the **Run Gemma in your browser** panel. Compatible browsers can opt in to a ~2.01 GB Gemma 4 E2B download, stored locally in browser-private storage. WebGPU inference generates a private, streamed plan from your own notes. No Sloth Mode sign-in, local connector or API key is required.

This is an experimental path: the runtime and first model download need network access; offline reload, cross-browser GPU compatibility and on-device performance are not yet verified. It does not provide operating system integration. The Windows/Linux companion remains supported. See [browser model setup and deployment notes](docs/BROWSER_AI.md).

## What works today

- Browser-native Gemma 4 planning (beta) with user-initiated download, WebGPU, cancellation and local model storage.
- Local task capture and a paired browser workspace, with notes stored on the user's computer.
- Optional local Gemma using separately downloaded and checksum-verified GGUF weights and `llama.cpp`.
- A Windows native connector + installer, and a Linux x86-64/ARM64 background connector with an application launcher.
- Optional account sign-in via separately configured Cloudflare Access (Google, GitHub, email); *only explicitly saved account-tracked tasks* are stored in the account database. Local notes are not silently synced.
- A public demo and source-linked triage interface at `sloth.marceli.dev`.

## Start from source

Requirements: **Go 1.23+**. Windows and Linux build scripts are under [`scripts/`](scripts/).

**Linux (x86-64, systemd user services preferred):**
```bash
./scripts/build-linux.sh
./linux/install.sh
```
Run `sloth-mode-connector --setup` via the installed launcher or `~/.local/share/sloth-mode/bin/sloth-mode-connector --setup`.

**Windows (PowerShell with Go on PATH):**
```powershell
.\scripts\build-windows.ps1
```
The installer is created as `build/SlothMode-Setup.exe`. Run it, then open **Setup & Pairing**. The installer builds the embedded connector first; binaries are intentionally excluded from Git history.

**Test:**
```bash
cd connector && go test ./... && go vet ./...
cd ../cloud && node --test account.test.mjs
cd .. && node --test web/browser-ai.test.mjs
```
Cloud tests require Node.js 22+ with the built-in `node:sqlite` available.

## Privacy and accounts

| Feature | Guest | Signed in |
| --- | --- | --- |
| Local connector & notes | ✅ | ✅ |
| Browser Gemma (supported GPU, optional model download) | ✅ | ✅ |
| Quiet Lab sample demo | ✅ | ✅ |
| Local Gemma | ✅ (after model setup) | ✅ |
| Save tasks in cloud account | — | ✅, opt-in |
| Manage account-linked devices | — | ✅ |
| Access your home Gemma remotely | Not implemented | Not implemented |

The companion listens on **127.0.0.1:39871** and requires explicit browser pairing. A signed-in browser account is *not* the same thing as permission to use the local connector. See [SECURITY.md](SECURITY.md).

## Project map

- [`connector/`](connector/) — native Windows + Linux background app, note storage, local model management, pairing.
- [`web/browser-ai.mjs`](web/browser-ai.mjs) — optional browser WebGPU inference and model caching.
- [`installer/`](installer/) — graphical Windows installer/updater source.
- [`linux/`](linux/) — per-user Linux installer and launcher assets.
- [`web/`](web/) — static page source (the hosted version is served by Cloudflare Workers).
- [`cloud/`](cloud/) — Cloudflare Access/D1 account service source and schema, with separately documented integration steps.
- [`scripts/`](scripts/) — reproducible source builds and signing helpers.

**Self-hosting note:** The account service uses Cloudflare Access, D1 and registered OAuth providers. The files under `cloud/` include an earlier, testable service implementation and task-API integration sample; they are **not an out-of-the-box copy of the entire live deployment**. Configure your own identity providers, domains, secrets and databases. Never reuse the owner's production account or database identifiers. For that reason, self-hosting parity with production is still being documented; see [docs/SELF_HOSTING.md](docs/SELF_HOSTING.md).

## Release status

For hardware tests and guest/account privacy checks, use the [release validation checklist](docs/RELEASE_VALIDATION.md). CI builds alone do not establish device-level reliability.


The Linux port is **beta**. Linux x86-64 has passed a headless connector smoke test; ARM64 was cross-compiled but not tested on physical hardware. GPU acceleration and first real Gemma inference require a machine-level test. Windows binaries are not code-signed. Don't rely on the preview in sensitive workflows without reviewing the code.

## Contributing

Contributions and issue reports are welcome. See [CONTRIBUTING.md](CONTRIBUTING.md), [SECURITY.md](SECURITY.md) and [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md).

Released under the **MIT License** (see [LICENSE](LICENSE)). Gemma weights, `llama.cpp`, Google/GitHub logos and other third-party services retain their own licenses and terms; **models are not included in this repo**.

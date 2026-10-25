# 🦥 Sloth Mode

**Slow down. Get things done.** A local-first AI productivity assistant with optional cloud accounts.

**Release status: public web beta, unsigned Windows preview, experimental Linux build. Not yet a signed stable release.** See the [public beta guide](docs/PUBLIC_BETA_LAUNCH.md) for test evidence, download instructions, security limitations and release gates.

[**Try the public beta**](https://sloth.marceli.dev/dashboard) · [**Windows unsigned preview**](https://github.com/marceli1404/sloth-mode/actions/runs/40133971950) · [**Send beta feedback**](https://github.com/marceli1404/sloth-mode/issues/30) · [Beta limitations](docs/PUBLIC_BETA_LAUNCH.md)

**Week 3 judges / testers:** [60-second project overview, build-window evidence and repeatable offline Gemma proof](docs/WEEK3_JUDGES_QUICKSTART.md). The interactive guest demo is intentionally not presented as real model inference; the opt-in test uses three fictional notes and reports actual local timings.

The web beta is open without an account. The downloadable Windows build is **unsigned and experimental**, requires a GitHub artifact download (GitHub login may be required), and has unresolved native security review gates. Don't disable security protections to run it. If you test it, please report bugs with fictional sample tasks and no credentials.

Sloth Mode supports optional **browser-native Gemma 4 E2B using WebGPU** (experimental; opt-in ~2.01 GB model download), alongside **Gemma 3 GGUF through `llama.cpp`** for native inference. The companion runs in the background on Windows or Linux and pairs explicitly with the web interface through a loopback API. **Ollama and an account are not required.** The public Quiet Lab sample demo uses illustrative deterministic sorting, not actual Gemma inference.

## Browser Gemma beta (account-free)

From the [dashboard](https://sloth.marceli.dev/dashboard), use **Run Gemma in your browser**. A compatible HTTPS/WebGPU browser can explicitly download the ~2.01 GB Gemma 4 E2B model into browser-private storage. Planning is local, and no account, native companion or cloud inference endpoint is needed. The original Windows/Linux companion and opt-in account tasks remain separate.

The browser path is **experimental**. Real GPU inference, browser cache reuse, CORS and runtime asset delivery are not yet validated on production hardware; downloads are checked by exact size but not a streaming SHA-256 digest. See [browser Gemma technical and deployment checklist](docs/BROWSER_AI.md). **The live site has been deployed separately from this repository.** Its embedded production Worker includes account/API routing not fully represented by the standalone reference source; merging a GitHub change alone does not update that Worker.

## One integrated workspace

The [dashboard](https://sloth.marceli.dev/dashboard) combines the optional account task list, Quiet Lab, browser-native Gemma, Windows/Linux companion, integrations and interactive demo in one accessible, responsive window. Guest mode works without an account; the connector remains optional and runs locally. The [unified workspace guide](docs/UNIFIED_WORKSPACE.md) explains privacy, deep links and deployment.

## What works today

- Optional, opt-in browser-native Gemma 4 task planning on capable WebGPU devices (beta).
- Local task capture and a paired browser workspace, with notes stored on the user's computer.
- Optional local Gemma using separately downloaded and checksum-verified GGUF weights and `llama.cpp`.
- A Windows native connector + installer, and a Linux x86-64/ARM64 background connector with an application launcher.
- Optional account sign-in through separately configured Cloudflare Access (Google, GitHub, email). The account page supports an explicit **Save to my account** task list; it stores only tasks you enter there. Local notes and Gemma prompts remain on your computer and are never imported automatically. The public production session implementation is separate from this self-hosting reference.
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
| Browser Gemma (supported WebGPU, explicit download) | ✅ | ✅ |
| Local connector & notes | ✅ | ✅ |
| Quiet Lab sample demo | ✅ | ✅ |
| Local Gemma | ✅ (after model setup) | ✅ |
| Explicitly save account tasks (separate list) | — | ✅, opt-in |
| Manage account-linked devices | — | ✅ |
| Access your home Gemma remotely | Not implemented | Not implemented |

Account task entries can be read by other browsers signed into the same account. This is separate from local notes; **no task sync or remote Gemma relay is implemented**. The Cloudflare reference requires an explicit D1 schema migration and must be independently validated before deployment.

The companion listens on **127.0.0.1:39871** and requires explicit browser pairing. A signed-in browser account is *not* the same thing as permission to use the local connector. See [SECURITY.md](SECURITY.md).

## Project map

- [`connector/`](connector/) — native Windows + Linux background app, note storage, local model management, pairing.
- [`web/browser-ai.mjs`](web/browser-ai.mjs) — optional browser WebGPU inference and local model caching.
- [`installer/`](installer/) — graphical Windows installer/updater source.
- [`linux/`](linux/) — per-user Linux installer and launcher assets.
- [`web/`](web/) — static page source (the hosted version is served by Cloudflare Workers).
- [`cloud/`](cloud/) — Cloudflare Access/D1 account service source and schema, with separately documented integration steps.
- [`scripts/`](scripts/) — reproducible source builds and signing helpers.

**Self-hosting note:** The account service uses Cloudflare Access, D1 and registered OAuth providers. The files under `cloud/` include an earlier, testable service implementation and task-API integration sample; they are **not an out-of-the-box copy of the entire live deployment**. Configure your own identity providers, domains, secrets and databases. Never reuse the owner's production account or database identifiers. For that reason, self-hosting parity with production is still being documented; see [docs/SELF_HOSTING.md](docs/SELF_HOSTING.md).

## Release status

The [public beta launch guide](docs/PUBLIC_BETA_LAUNCH.md) links to the latest documented unsigned Windows preview and explains the remaining security/hardware gates. For supported hardware claims and guest/account privacy verification, follow the [release validation checklist](docs/RELEASE_VALIDATION.md). **CI builds alone do not establish device-level reliability or clear CodeQL alerts.**


The Linux port is **beta**. Linux x86-64 has passed a headless connector smoke test; ARM64 was cross-compiled but not tested on physical hardware. GPU acceleration and first real Gemma inference require a machine-level test. Windows binaries are not code-signed. Don't rely on the preview in sensitive workflows without reviewing the code.

## Contributing

Contributions and issue reports are welcome. See [CONTRIBUTING.md](CONTRIBUTING.md), [SECURITY.md](SECURITY.md) and [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md).

Released under the **MIT License** (see [LICENSE](LICENSE)). Gemma weights, `llama.cpp`, Google/GitHub logos and other third-party services retain their own licenses and terms; **models are not included in this repo**.

# 🦥 Sloth Mode

**Slow down. Get things done.** A local-first AI productivity companion with optional accounts.

[Website](https://sloth.marceli.dev) · [Workspace](https://sloth.marceli.dev/dashboard) · [Contributing](CONTRIBUTING.md) · [Security](SECURITY.md)

The Sloth Mode companion runs on Windows and Linux. It pairs explicitly with the browser and can run Gemma 3 locally through llama.cpp without requiring Ollama or an account.

- **Guest:** use the public Quiet Lab demonstration, local companion, and local notes.
- **Signed in (optional):** explicitly save account-tracked tasks and manage linked devices.
- **Privacy:** local notes are not automatically synced to the cloud. The public Quiet Lab example uses deterministic demonstration rules, not local Gemma inference.

## Open-source import status

The website source is available under `web/`. The cross-platform connector, Windows installer, Linux installer, and cloud service source have been prepared and tested as a separate source release and are being imported into this repository. Until then, **this GitHub tree is not a full standalone build of the live application**.

Once the source import is complete, the `connector/`, `installer/`, `linux/`, `cloud/`, and `scripts/` directories will contain the native and cloud components and reproducible build instructions.

## Security and third parties

See [SECURITY.md](SECURITY.md). Cloudflare Access, OAuth providers, and the D1 database require your own configuration; no production credentials or model weights are distributed.

**MIT Licensed.** Gemma weights and llama.cpp have their own licenses and distribution terms.

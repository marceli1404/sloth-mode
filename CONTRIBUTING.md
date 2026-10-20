# Contributing to Sloth Mode

Thanks for helping make local-first AI more accessible.

1. Open an issue describing a bug or improvement before making invasive changes.
2. Fork the repo and work on a feature branch; keep changes small and focused.
3. Run `go test ./...` and `go vet ./...` under `connector/`; for installer changes, run `go test ./...` under `installer/` after building the Windows payload.
4. Include tests for pairing, authentication, credential handling, and model downloads. Respect opt-in cloud sync and guest mode.
5. Do not include credentials, personally identifying logs, real saved tasks, paid API keys, binary weights or any private update-signing key.
6. Open a PR with setup steps, screenshots for UI changes, and a concise risk assessment.

We prefer accessible interfaces, native reduced-motion support, safe defaults and real benchmarks over simulated claims.

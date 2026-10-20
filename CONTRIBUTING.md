# Contributing to Sloth Mode

Contributions and reproducible bug reports are welcome. Open an issue, make a focused feature branch, add regression tests, and submit a PR explaining behavior and security/privacy implications.

Please preserve **guest mode**, explicit pairing, reduced-motion accessibility, and the opt-in nature of cloud task storage. Never commit local task databases, logs, passwords, API credentials, model weights or private signing keys.

For the source release, use Go 1.23+, run `go test ./...` and `go vet ./...` inside `connector/`, and test installer changes on Windows.

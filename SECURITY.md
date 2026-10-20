# Security Policy

Report vulnerabilities **privately**. Do not publish tokens, account details, exploit steps, or real task data in issues. Use the repository's private vulnerability-reporting feature when available; otherwise request a private channel without disclosing details.

- The local connector binds to `127.0.0.1` and requires explicit browser pairing.
- Cloud accounts are optional. Signing in is not permission to access a user's local notes or Gemma.
- Cloud-tracked tasks are opt-in and distinct from local notes.
- Credential storage uses Windows DPAPI or an unlocked Linux Secret Service keyring.
- Runtime/model downloads are user-approved and verified against pinned hashes.

This is pre-1.0 software. Do not expose the local connector's port to the internet.

# Security Policy

Sloth Mode runs software on a personal computer and handles private tasks and, optionally, third-party API keys. Treat security-sensitive changes carefully.

## Reporting a vulnerability

**Do not open a public issue containing an exploit, pairing key, API credential or personal data.** Use GitHub's **Report a vulnerability** flow under the repository's Security tab, if enabled. If private reporting is unavailable, open a non-sensitive issue requesting a private contact channel without disclosing the vulnerability itself.

## Trust boundaries

- Connector binds to `127.0.0.1`; only explicitly paired browsers can invoke private operations. Do not expose its port publicly.
- Cloud login is **optional**. Authentication to the account service does not automatically grant access to the local connector.
- Optional cloud-tracked tasks are different from local notes. Never silently migrate or upload local data.
- Windows credentials use DPAPI; Linux integration/account credentials require an unlocked Secret Service keyring. There is no silent plaintext fallback.
- Model/runtime binaries must be downloaded from known URLs, checked against pinned hashes and installed only with user approval.
- Cloudflare Access/D1 example sources require independent security review and private deployment configuration before production use.

## Supported versions

Only the most recent source and explicitly announced beta builds are maintained. This is a pre-1.0 experimental release.

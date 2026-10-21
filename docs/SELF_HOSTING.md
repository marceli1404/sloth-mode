# Self-hosting and cloud integration

The connector can be run locally without the cloud account service. The website is a mostly static page with optional authenticated endpoints.

The `cloud/` directory contains a tested Cloudflare Access account/device registry implementation, a device-enrolment service, a D1 schema, and an **explicit opt-in account task endpoint** implemented in the Access Worker, plus a separate legacy session-auth task API example. **These are reference components, not a complete, production-ready deployment scaffold.** The current live site has additional session handoff logic, deployment configuration, and pages not yet consolidated into this repository.

For a self-hosted environment: provision a separate Cloudflare account/zone; configure Access with *your own* Google/GitHub OAuth apps and allow policies; create a new D1 database; apply the revised `cloud/schema.sql` migration to a fresh staging D1 (or review and apply the new `account_tasks` table on an existing DB); bind the same DB to your account and enrolment Workers; configure domain/origin allowlists; and verify all authentication and CORS checks. Do not publish OAuth client secrets or reuse the project's production database IDs.

The Access Worker uses an authenticated `/v1/tasks` GET/POST endpoint for its separate task list. Every write requires the exact account origin and a JSON body; server-derived account IDs scope all task reads/updates/deletes. Task text is capped at 280 characters and 200 entries per account. The UI does **not** read or sync local connector notes. The old `task-api.example.mjs` illustrates a *different*, browser-cookie session architecture and is not directly interoperable with Access-worker sessions.

If you publish a public instance, remove any owner-specific policies, enforce rate limits, add an explicit privacy policy and obtain users' consent for cloud task storage. Audit Google/GitHub/email Access identity configuration and user separation on your own deployment. Cloudflare Access rules, secrets and production routing are **not** created by this PR.

## Open items

- Consolidate the complete first-party cookie/session-handoff Worker from the running production app into this source tree.
- Provide infrastructure-as-code for staging and production worker deployments.
- Verify Access identity, account task persistence, write-origin checks and multi-user separation against actual staging D1 and browsers. Local unit tests are not a production OAuth test.
- Release reproducible, signed Linux/Windows archives and verify real inference on target hardware.

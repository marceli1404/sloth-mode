# Self-hosting and cloud integration

The connector can be run locally without the cloud account service. The website is a mostly static page with optional authenticated endpoints.

The `cloud/` directory contains a tested Cloudflare Access account/device registry implementation, a device-enrolment service, a D1 schema, and the task API example. **These are reference components, not a complete, production-ready deployment scaffold.** The current live site has additional session handoff logic, deployment configuration, and pages not yet consolidated into this repository.

For a self-hosted environment: provision a separate Cloudflare account/zone; configure Access with *your own* Google/GitHub OAuth apps and allow policies; create a new D1 database; run the schema migration; bind it to your account and enrolment Workers; configure domain/origin allowlists; and verify all authentication and CORS checks. Do not publish OAuth client secrets or reuse the project's production database IDs.

If you publish a public instance, remove any owner-specific policies, enforce rate limits, add an explicit privacy policy and obtain users' consent for any cloud task storage.

## Open items

- Consolidate the complete first-party cookie/session-handoff Worker from the running production app into this source tree.
- Provide infrastructure-as-code for staging and production worker deployments.
- Add tests for the browser-authenticated cloud task API in a real deployment.
- Release reproducible, signed Linux/Windows archives and verify real inference on target hardware.

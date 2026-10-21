# One-window Sloth Mode workspace

The dashboard `/dashboard` now becomes a single responsive application window, with persistent navigation and panels for Today, Browser Gemma, Desktop companion, Integrations and Explore. The former `/integrations` URL opens the same window with Integrations selected. This is a UI composition of the existing browser features, not a replacement for the native Windows/Linux connector.

## Interaction and privacy

- **Today** includes optional account task sync and the locally processed, transparent-rules Quiet Lab.
- **Browser Gemma** retains the user-approved WebGPU model download, OPFS cache, stop and planning functionality; no account or connector is needed.
- **Desktop companion** includes device pairing, local tasks, hardware detection, Gemma 3/llama.cpp and installation help. No browser window can directly inspect arbitrary local files or processes; the separately installed connector does the OS-specific work.
- **Integrations** offers the existing, explicitly configured companion API tools. API keys are not copied to the cloud.
- **Explore** contains the illustrative walkthrough and privacy background.

Navigation only moves existing HTML nodes into one tabbed shell after page load. It does not clone inputs, replace handlers or send notes to Sloth Mode servers. URL hashes map old anchors such as `#quiet-lab`, `#connector` and `#integrations` to the relevant panel. Buttons expose tablist/tab/tabpanel ARIA semantics, arrow-key navigation, focus outlines and responsive wrapping.

## Cloudflare release procedure

The website is served by the production Worker `sloth-mode-demo`, which contains separate live account/auth/task APIs not fully represented by the static repo. Do **not** overwrite it with a static build or the self-hosted account reference Worker.

1. Back up/inspect the currently deployed version and verify it still has the `AUTH_DB` D1 binding and `/auth/*`, `/api/tasks`, `/browser-ai.mjs` routing.
2. Patch only the live Worker’s existing embedded `PAGE` string: change the startup data page condition to include `/integrations`, add one stylesheet link `/workspace.css` and one module script `/workspace.mjs`. Add the exact version-controlled CSS and JS as constants, and serve those two paths with `text/css` and `text/javascript` content types, including HEAD support.
3. Add `'self'` to the existing `style-src` CSP policy so the same-origin CSS is permitted; preserve the current Gemma CDN, WASM and local connector policies. Avoid adding wildcard external origins.
4. Upload as a **non-deployed** Worker version and confirm bindings/metadata; then promote to traffic after PR CI is green. Keep the previous version available to revert.
5. Use a browser to confirm the shell appears on `/dashboard`, the old `/integrations` path selects Integrations, tab navigation shows all feature panels, Gemma code still loads, account task APIs remain intact, and the home page is unchanged.

When JavaScript does not load, the old dashboard remains visible instead of becoming a blank app window; the enhanced design only hides marketing sections after its module has completed successfully. Browser Gemma inference and the large model download still need separate real-device verification.

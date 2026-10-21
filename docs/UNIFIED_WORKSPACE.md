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

## Research-led dashboard redesign (October 2026)

Design and implementation informed by the Sloth Mode deep-research review of app navigation, WCAG 2.2, WAI-ARIA patterns, cognitive load, and privacy boundaries.

**Product hierarchy:** Today is useful immediately as guest. Its Quiet Lab editor appears before optional account tasks; the heading-level structure makes the action obvious. The contextual quick links are navigational, not fake capability toggles.

**Information architecture:** The same five top-level destinations are now genuine navigation links inside a labelled `<nav>`. `aria-current="page"` marks the selected destination. Native Tab/Shift+Tab and browser history work without custom arrow-key tab widget semantics; legacy `#connector`, `#quiet-lab`, `#sm-browser-ai`, and `#integrations` hashes still resolve. At mobile widths, a five-item fixed bottom navigation uses shorter visible labels while full accessible names remain on the links.

**Visual direction:** Warm paper background, deep pine navigation, restrained sage highlights, Georgia editorial headings, system sans-serif labels, consistent 10–18 px radii and visible focus outlines. No invented productivity metrics, streaks or dashboard charts. Internal surfaces share a single visual rhythm with the browser Gemma beta, native companion and optional integrations.

**Privacy/capabilities:** Independent status labels distinguish `Account · guest` / `Account tasks · signed in`, `Companion · not paired` / `Companion · paired`, and browser Gemma's independent beta/readiness status. Being signed into an account never implies that the local computer is paired, nor that local notes synchronize. Model-quality controls appear under **This computer**, preserving their existing event listeners and ID references.

**Progressive enhancement:** Original feature DOM elements, forms, event listeners and model state are moved into the shell, never cloned. Only a successfully mounted shell suppresses the old page layout. No new server endpoints, hidden prompt upload, telemetry, cloud migrations, external fonts or SDKs.

**Checks:** Node tests cover semantic navigation, route aliases, status wording, original DOM references, guest-first ordering, reduced motion, bottom safe-area spacing and homepage isolation. A production promotion additionally requires a **real** DOM smoke test for dashboard/Integrations, stylesheet/JS CSP and MIME, task/auth presence, and rollback. Browser WebGPU inference and full 2.01 GB model integrity remain separately gated, not implied by a visually polished UI.

**Deferred engineering work:** Self-host pinned LiteRT/WASM, implement streaming SHA-256 model verification, explicit Unload GPU, persistent storage UX, real-browser accessibility/performance CI and physical GPU release tests. UI design alone does not close these security and runtime release gates.

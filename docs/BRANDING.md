# Sloth Mode brand mark

**Canonical source:** `web/favicon.svg` (the homepage's original, line-drawn sloth mark).

The same mark, dark forest background (`#173b35`) and pale lime lines (`#d3ebb5`) are used in:

- Homepage header, browser favicon and account preview header.
- Unified dashboard sidebar (no Unicode placeholder glyph).
- `connector/icon.svg`, `linux/sloth-mode.svg`, and embedded Windows `connector/sloth.ico`.

When the logo is updated, edit the canonical SVG first, then copy the SVG *identically* into the connector/Linux icon sources. Windows `.ico` is a multi-resolution raster of that SVG, with 16 × 16, 32 × 32, and 256 × 256 PNG-backed images; regenerate it from the same art (not an alternative mascot). Use alpha-aware anti-aliasing and review the 16px appearance before shipping a release.

`node --test web/brand.test.mjs` checks the SVG assets, inline vector shape, color tokens and basic ICO container integrity. It does not prove human visual similarity or validate the signed native installer.

Keep the wordmark spelling `slothmode.` consistent. The illustrated mascot in marketing scenes is decoration, **not** an alternative product logo. Account site and native installers may be deployed on independent schedules; source parity does not prove every published artifact has been replaced.

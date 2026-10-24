# Windows installer refinement — 24 October 2026

The old running preview visibly clipped body text, painted dark rectangles behind labels and could extend below a small display. The new native Go/Win32 layout uses explicit font roles, cream/forest surfaces, the canonical Sloth Mode icon, three setup steps and one primary action. It states that this is an unsigned preview and that Gemma requires a separate multi-GB download.

The refinement starts from main `ae35a26`, retaining its runtime/archive fixes and real `/v1/health` startup probe. It integrates the earlier reviewed asynchronous install/update work, staged checksum-verified application replacement, rollback reporting, native checkbox/keyboard navigation and single-installer update handoff. Repeated launch clicks cannot start overlapping readiness checks. Startup choices are preserved during an upgrade and become read-only once applied.

## Verification

- Windows connector tests, installer build and both `go vet ./...` checks pass using portable Go 1.27.3.
- Installer: **12 top-level tests pass**, with one opt-in visual fixture skipped by default. Tests include failed replacement/rollback, incomplete payload preservation, singleton exclusion/handoff, async operation controls and rejecting unrelated localhost health responses.
- Window bounds fit 100%, 125%, 150% and 200% DPI across four work areas, including a 1366×728 work area and a monitor left of the primary display. A failing 200% test caught a non-client title-bar sizing error; fitting now subtracts the actual native frame dimensions.
- GDI measures every static control's text at scales 0.85, 1, 1.25, 1.5 and 2. No measured text overflows its height and no control lies outside the client. This is layout evidence, not a claim that Windows display settings were changed on physical monitors.
- The actual production UI was opened in an opt-in native test fixture. Install, launch, update and folder-opening commands are intercepted in that fixture; it does not install or modify the user's existing companion. The visible window, primary/Close focus navigation and Escape were inspected. [Screenshot](images/installer-preview.jpg) shows the fixture with its preview notice.
- **35 Node tests pass** for account, browser AI, workspace, brand and beta access. A pre-existing Windows checkout line-ending mismatch caused the generated account-bundle equality test to fail initially; normalizing the local cloud files to LF resolved it with no Git content changes.
- Local unsigned package: **15,165,440 bytes**, SHA-256 `1a24cda6f8a72f3b606568b96d7d11584d2ff0b93c0205188c89d1c95673910a`. Its hash is specific to the local build; CI publishes its own checksum.

CI now retains the unsigned preview artifact and SHA-256 for branch/PR builds as well as main and runs Windows static analysis. No stable updater manifest, signing key, live website deployment or release tag was changed.

## Limits

No new clean-profile install, reboot, browser pairing, real-model run or network-off test was performed for this candidate. The previously reported successful Windows experience belongs to older previews. Independent security review, signing, additional device coverage and the other [release checks](RELEASE_VALIDATION.md) remain open. [Previous-winner comparison](WINNER_COMPARISON.md) informed the clearer setup and article; it does not predict an award.

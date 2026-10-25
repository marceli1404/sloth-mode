# Final native acceptance — 24 October 2026

The actual unsigned installer was exercised on the owner's existing Windows
profile, using the real **Install & open** control. This was an upgrade, not a
clean-profile installation. A private local backup was made before the upgrade.
The installed executable matched its packaged payload, and the owner's notes
file and launch-at-sign-in registry value were unchanged. Chrome reconnected
using its existing device-bound signing key.

## Findings and changes

The installed pairing window still used the old gray surfaces and extended
below the captured work area. Its replacement uses cream/forest surfaces,
regular body type, distinct headings and a compact work-area-centered layout.
The coordinator inspected the actual native window after building it: all four
steps, startup control and bottom tray/exit controls were visible. A first
visual pass caught excessive bold text and a clipped model instruction; both
were corrected. Geometry and GDI text-fit tests use the actual font roles.
Very small work areas below the layout's readability floor remain unsupported.

The first real live-site/Q6 CPU test rejected the fictional garden-club task
because the model returned `now/work` as its priority. Validation correctly
left the note unclassified. The prompt's shorthand examples were ambiguous.
Examples now name each field separately, and llama.cpp receives a JSON decoding
schema constraining source IDs, item count, field names, priorities, categories
and string lengths. Existing semantic validation and atomic snapshot checks
remain in place; a runtime that ignores the schema still cannot save the
malformed priority. This uses llama.cpp's documented
[schema-constrained JSON format](https://github.com/ggml-org/llama.cpp/tree/master/grammars).

## Real model verification

- Production Q6_K model and CPU llama.cpp runtime; **zero GPU layers**.
- Unmodified production engine settings; six CPU inference threads reported by
  llama.cpp. No test-only thread or batch substitutions.
- Three synthetic notes returned valid `now`, `later`, and `no_action` results.
  Startup including model SHA-256 verification: **4.320 s**; triage: **19.456 s**.
- The exact garden-club note was retried through the live Chrome website and
  corrected candidate. It saved priority `now`, category `work`, a concrete next
  step, and a summary retaining **tomorrow at 4pm**.
- Inference used isolated local storage with fictional notes. Existing approved
  model/runtime files were reused; personal notes and account links were not
  sent to the model. Private pairing material and local backups are excluded
  from published evidence.

The three-note log is retained locally at
`validation/installed-acceptance-20261024/schema-real-gemma.log`. This is focused
acceptance evidence, not general model accuracy, quantified human time savings,
a clean installation, a reboot test or machine-wide network-off verification.
The prior 12-note corpus and recording remain historical evidence for their
specific earlier source revisions.

## Final installed-candidate acceptance — 25 October

The owner explicitly resumed Computer Use. The final local installer built
from clean implementation `511bc0fe58479896adcac475951b5a6c3a8e3087` was
installed through its actual **Install & open** control. Setup SHA-256:
`cea33949a836936da4b177519a732aa49bbcb89bb97901f7d52c1eeb981291d0`.
Installed connector SHA-256:
`fca86aeb392ff9a1e275af706553a5603221e81a6f4eb302c9752b3a04aaaef1`,
matching the packaged payload. The health endpoint reported `0.15.0-beta` and
`ok: true`. This verifies this local build; the separately compiled CI archive
has its own checksum and was not installed by this run.

Passed on the existing Windows profile:

- Existing notes-file SHA-256 and launch-at-sign-in registry value unchanged.
- Actual installed pairing window opened with all controls visible.
- Live Chrome reconnected using its existing trusted-browser signing key.
- Normal **Exit connector** stopped the loopback listener; relaunching the
  installed executable restored health, pairing and model readiness.
- Exact installed executable, with isolated fictional-note storage, ran Q6_K
  CPU inference through the live website. The fictional garden-club fee became
  priority `now`, category `admin`, with its original note preserved. The model
  omitted timing; the UI correctly warned about that omission and linked it to
  source note #1. The displayed completion text correctly used singular
  “1 task”. This is a functional check, not a representative accuracy score.
- Closing the normal pairing window hid it to the tray while the companion
  remained healthy. The normal workspace was left connected; the test model
  process was stopped. Personal notes were not used for inference.

The synthetic result is saved alongside the earlier evidence as
`docs/evidence/2026-10-24/final-installed-cpu-task.json`.

The web export control was clicked, but delivery of its Markdown download was
not independently verified. Browser policy blocked opening `chrome://downloads/`
because only HTTP(S) URLs were allowed. No alternate route bypassed that block.
The previous version's verified export recording remains historical evidence;
it is not substituted for this missing final-build download observation.

Clean-profile installation, a complete setup/download recording, machine-wide
network-off testing, signing/security review and broader hardware coverage
remain separate open checks. No Windows network/security settings were changed.

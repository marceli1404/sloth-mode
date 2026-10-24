# Competition entry readiness — 24 October 2026

The entry is published, reviewable without an account, and supported by an
uncut real-model demonstration and public source evidence. These are reasons
to treat it as a credible entry; they do not establish a numerical chance of
winning or guarantee an award.

## Current entry audit

| Requirement or criterion | Direct evidence | Result |
| --- | --- | --- |
| Published DEV submission and challenge tag | [Existing entry](https://dev.to/marceli/sloth-mode-i-built-a-local-gemma-assistant-so-my-brain-can-be-a-little-lazier-ci4), published 22 October; `hf26challenge` tag and template sections present | Verified |
| New project in the entry window | Public initial commit dated 20 October; current source and entry linked | Project history is consistent with the window; personal entrant eligibility is not independently established here |
| Slow and Steady theme | The entry identifies the repeated chore between capturing a task and deciding what to start | Clear fit; human effort saved remains unmeasured |
| Open innovation at the core | Real Gemma 3 GGUF through the native companion and `llama.cpp`; source-linked suggestions and published prompt/evaluation | Verified within the recorded scope |
| Demo and code | Anonymous guest workspace; public uncut MP4 and pinned source/evidence links | Verified; Quiet Lab is explicitly rules-based |
| Best Use of Gemma category | Category is explicitly identified in the existing entry | Verified |
| Writing quality | Concrete opening, specific task workflow, model failures and limits explained; prior-winner comparison published | Qualitative review completed, not a judge's score |
| Technical execution | Source tests/CI, real CPU/browser recording, atomic stale-input rejection and timing-review checks | Verified for the documented source, fixtures and small convenience corpus |

The [official rules](https://dev.to/page/hacktoberfest-week3-2026-10-19-contest-rules)
weight writing most heavily and also assess theme, creativity, execution and
partner use. The public-entry audit found no missing required submission
element. The deadline is 25 October 2026 at 11:59 PM PDT, equivalent to
26 October at 06:59 GMT in London.

## Anonymous judge path

1. Read the [published entry](https://dev.to/marceli/sloth-mode-i-built-a-local-gemma-assistant-so-my-brain-can-be-a-little-lazier-ci4).
2. Open the [guest workspace](https://sloth.marceli.dev/dashboard). The sample
   works without a Sloth Mode account and labels its deterministic rules.
3. Download the [uncut real Gemma recording](https://raw.githubusercontent.com/marceli1404/sloth-mode/f73cfc4a01b6cbb672cac174aaa004dbe6d0b6fe/docs/evidence/2026-10-24/real-browser-gemma-uncut.mp4).
   The public MP4 returns successfully and is about 1.7 MB. It shows fictional
   tasks, pairing, actual CPU inference, source-linked results, editing,
   export and explicit reconnect. It is separate from the promotional video.
4. Inspect the [pinned recording and evaluation evidence](https://github.com/marceli1404/sloth-mode/blob/f73cfc4a01b6cbb672cac174aaa004dbe6d0b6fe/docs/evidence/2026-10-24/README.md).
5. Use the candidate build only if testing installation. Its Actions artifact
   may require a GitHub login; the entry discloses that extra friction. It is
   an unsigned preview from [draft PR #34](https://github.com/marceli1404/sloth-mode/pull/34).

## What is still unproven

- Clean-profile installation, upgrade preservation and a full setup recording
  for the final candidate. Native layout fixtures do not prove installation.
- An independent network-off run for that candidate. Older owner reports are
  useful history, not new-candidate acceptance evidence.
- Windows Authenticode signing and an independently reviewed stable release.
- Representative model quality, multilingual deadline handling, cross-device
  GPU/Linux compatibility and measured human time savings.

The small frozen corpus is reported as such: candidate priorities matched
12/12 author-defined cases, but timing phrases were often omitted. The limited
warning layer points back to original notes; it does not repair semantics.

## Cached-model integrity check

The final source audit found that model reuse previously relied on size alone.
The candidate now checks an existing regular model file against its pinned
SHA-256 before explicit setup reuse and before engine launch. Verification is
bounded, cancelable and separate from status polling. Explicit setup stages a
fully verified replacement before moving a damaged model; failed publication
restores the old file or retains an identified recovery copy. Normal setup
consent remains required before any download.

Focused tests cover valid content, same-size corruption, non-regular paths,
size limits, cancellation and replacement/restore failure. Connector tests and
vet pass. A fresh isolated CPU smoke also accepted the existing 2.49 GB pinned
Q4_K_M model and triaged three fictional notes. Whole startup, including hash
verification and loading, was 5.813 seconds; generation was 19.491 seconds.
These are a separate single run, not an updated measurement of the published
four-note video or a controlled performance comparison. Hash-only overhead is
unmeasured; each check reads the complete model file.

The isolated source copy used four CPU threads and zero GPU layers, with
separate test data and existing task-local assets. No downloads or personal
notes were used. The machine remained online; this does not close the offline
acceptance gate.

## Next acceptance action

The user's current companion is running on the production loopback port.
Replacing it or exercising the installer against this profile would affect
the running application and potentially its startup/shortcut settings. The
remaining independent acceptance check therefore needs either a disposable
Windows environment or explicit permission to update this installed companion.
Do not stop it, change networking or publish a stable updater merely to turn
an incomplete checklist into a passing result.

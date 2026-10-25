# Sloth Mode — Week 3 judges' quickstart

**Hacktoberfest Open-Source AI Challenge 2026 • Week 3, Slow and Steady • Best Use of Gemma**

**Submission:** [DEV article](https://dev.to/marceli/sloth-mode-i-built-a-local-gemma-assistant-so-my-brain-can-be-a-little-lazier-ci4) · [Live guest workspace](https://sloth.marceli.dev/dashboard) · [23-second promo overview](https://www.youtube.com/watch?v=GnaEFKQW9Vw) · [Public beta feedback](https://github.com/marceli1404/sloth-mode/issues/30)

> **What it does:** takes a messy set of local notes and asks an optional on-device Gemma model to suggest priorities and a next action. The person reviews the results; Sloth Mode does not autonomously send messages, execute commands, or upload the local notes to an account.

## First, make sure you're testing the right thing

| Experience | What is actually executing? | What this proves |
|---|---|---|
| Guest workspace, [Quiet Lab](https://sloth.marceli.dev/dashboard#quiet-lab) | Illustrative deterministic preview | You can explore the interaction without signing up or installing anything. **Not model output** |
| Paired native companion, [Gemma test](../connector/real_gemma_test.go) | Actual local GGUF Gemma 3 model via `llama.cpp` | Real model inference and structured suggestions for the *synthetic* notes below; only if runtime and model are already installed |
| Browser model tab, [Gemma 4 E2B](../web/browser-ai.mjs) | Experimental local WebGPU/LiteRT-LM after ~2GB explicit download | Source exists, but cross-device GPU inference, model integrity, and production compatibility are not yet established |
| Optional account | Explicitly account-saved tasks via Cloudflare D1 | Not the local notes or native prompts; account authorization is distinct from local pairing |

The promotional overview and actual model demonstration are separate:

- **[Watch the uncut 41-second browser-to-Gemma run](https://raw.githubusercontent.com/marceli1404/sloth-mode/f73cfc4a01b6cbb672cac174aaa004dbe6d0b6fe/docs/evidence/2026-10-24/real-browser-gemma-uncut.mp4).** Four fictional notes, actual browser pairing and signed requests, real Gemma 3 4B CPU inference, source-linked results, editing, export and reconnection. Click-to-visible-results took 25.063 seconds, including startup and polling, with four CPU threads and zero GPU layers.
- **[Read the inputs, output, limits and reproduction steps](https://github.com/marceli1404/sloth-mode/blob/f73cfc4a01b6cbb672cac174aaa004dbe6d0b6fe/docs/evidence/2026-10-24/README.md).** This controlled local run used already-installed assets. It does not establish clean installation, model-download reliability, machine-wide offline behavior or time saved.
- The earlier short video is a **promotional overview**, not proof of model execution.

For the native Windows route, use the identified unsigned candidate linked in
the [published entry](https://dev.to/marceli/sloth-mode-i-built-a-local-gemma-assistant-so-my-brain-can-be-a-little-lazier-ci4).
Its Actions download may require a GitHub login. Quiet Lab and the recording
remain available without installing the companion or creating a Sloth Mode account.

## Thirty-second read for judges

**Latest Windows check — 25 October:** the final local installer from implementation `511bc0f` passed an existing-profile upgrade, exact installed-payload checksum verification, normal exit/relaunch, live-browser reconnect and Q6 CPU inference with a fictional note. Personal notes and startup settings were preserved. [Acceptance report and synthetic result](FINAL_ACCEPTANCE_2026-10-24.md#final-installed-candidate-acceptance--25-october). This was not a clean-profile install or a network-off test. The separate CI artifact has its own checksum and was not installed in that run.

Sloth Mode addresses the uncomfortable gap between *capturing a thought* and *knowing what to do next*. A typical to-do list asks you to organise your entries; Sloth Mode offers an optional local model that drafts that organisation for you. You can start without an account, keep a local inbox, review suggestions beside the original notes, and edit or ignore them. Native inference is designed to run locally after setup; independent offline acceptance for the new installer remains pending.

**Why open weights matter here:** with Gemma 3 and `llama.cpp`, the source notes don't need to be sent to a hosted model endpoint for task planning. The runtime and quantization can be inspected or changed, and per-prompt vendor inference billing isn't mandatory. Initial model downloads still need connectivity, memory/disk requirements still apply, and the account and web delivery paths still use online services.

### What actually saves effort?

A representative *hypothetical* inbox:

1. `Pay electricity bill tomorrow.`
2. `Read a chapter of a novel when I have time.`
3. `I finished tidying my desk yesterday; no more action is needed.`

Without an assistant, you manually choose a priority and category, work out whether an entry is actionable, and invent a next step for each line. With real local inference, the same list is passed to Gemma and the suggestions are returned for your review; the [smoke test](../connector/real_gemma_test.go) checks that each returned item refers to an original input ID. **These lines are test inputs, not a transcript of model output.** We have not measured time saved in a controlled study, and the site Quiet Lab is not a proxy for such measurements.

## Repeatable real-Gemma demonstration (optional)

**Only run on a machine with Sloth Mode, a verified GGUF model and the native `llama.cpp` runtime already installed.** The test does *not* download files, touch actual notes, or send prompts to hosted AI services. The command runs on your machine, not in DEV's browser.

From the repo root:

Windows PowerShell:

```powershell
cd connector
$env:SLOTH_REAL_GEMMA_TEST = '1'
$env:SLOTH_TEST_BACKEND = 'cpu'
go test -run '^TestRealGemmaTriageOptIn$' -count=1 -v .
```

Linux:

```bash
cd connector
SLOTH_REAL_GEMMA_TEST=1 SLOTH_TEST_BACKEND=cpu go test -run '^TestRealGemmaTriageOptIn$' -count=1 -v .
```

**Expected procedure:** verify `llama-server` starts, send three fictional tasks, validate that model output preserves task IDs and structured fields, then record the returned `REAL GEMMA RESULT` with local startup time, triage time and quantization. The test has a five-minute deadline; it fails if installed runtime or model is unavailable. Repeat with networking disabled **after initial setup** to verify offline inference. GPU tests use explicit `SLOTH_TEST_BACKEND=vulkan` where drivers support it; CPU is the baseline.

**Evidence to collect (do not invent numbers):**

| Field | Record |
|---|---|
| Test date, repo commit, installer checksum | Pending tester report |
| OS, RAM band and backend | Pending tester report |
| GGUF quantization and verification | Pending tester report |
| Synthetic tasks with structured result | Pending tester report |
| Cold startup and inference duration | Pending tester report |
| Offline repeat successful | Pending tester report |
| Failure/recovery tests | Pending tester report |

Do not post device identifiers, account keys, pairing secrets or private task text. GitHub CI does **not** run the real-Gemma test because model weights are not bundled.

## Development timeline and build-window eligibility

The contest requires a new project and repository started and completed inside 19–25 October 2026. The [GitHub repository](https://github.com/marceli1404/sloth-mode) was created **20 October 2026 at 21:53:28 UTC**, with [first commit](https://github.com/marceli1404/sloth-mode/commit/0a51feca7a22a0b3ce06a5585740c8219f842e70) at 21:53:30 UTC.

- **20 Oct:** source was created and expanded to Windows/Linux native connectors, launcher/installer, site and reference cloud code; CodeQL and basic tests introduced.
- **21 Oct:** the project moved through release-hardening and Windows runtime repairs, including real-machine troubleshooting.
- **22 Oct:** Windows-owner installation, offline inference and persistence checks reported as successful; additional archive/path checks merged in [PR #28](https://github.com/marceli1404/sloth-mode/pull/28); the beta guidance and public access/feedback surfaces shipped in [PRs #29](https://github.com/marceli1404/sloth-mode/pull/29) and [#31](https://github.com/marceli1404/sloth-mode/pull/31).

The repo was created in the entry window. This chronology **does not itself prove all production or security gates** and shouldn't be presented that way. See [the release validation checklist](RELEASE_VALIDATION.md).

## Why the implementation isn't a generic to-do list

The technical distinction is the combination of:

1. **No login at the door:** guest capture and preview work without a cloud account.
2. **Explicit authority boundary:** the companion binds to `127.0.0.1:39871` and requires pairing; a signed-in cloud browser cannot silently access the machine.
3. **True local model execution:** optional, checked GGUFs and llama.cpp, CPU-first, manual quantization override, and an opt-in test that checks source IDs and measures local latency.
4. **Deliberate privacy separation:** local notes stay in local storage; only tasks deliberately entered in the separate account list go into D1.
5. **Offline *inference*, not magical offline website:** local model can run without Internet after downloading, while website hosting/account services still need connectivity.

## Limits we want judges and testers to know

- Windows is an **unsigned early-access preview**: [candidate artifact and SHA-256](https://github.com/marceli1404/sloth-mode/actions/runs/40133971950); don't disable security protection to use it.
- Final local-build upgrade, restart, pairing, data preservation and CPU inference were checked on one existing Windows profile. Offline success was owner-reported on older previews; independent final-candidate offline and clean-profile checks remain open.
- [Security review #7](https://github.com/marceli1404/sloth-mode/issues/7) remains open (historical path-handling/Zip Slip alerts are not independently dispositioned); passing CodeQL execution is not a security clearance.
- [Hardware/account validation #8](https://github.com/marceli1404/sloth-mode/issues/8) and [browser Gemma GPU validation #18](https://github.com/marceli1404/sloth-mode/issues/18) remain open.
- No signed stable updater release; source and website may not mirror every account-production binding. The official status is **public beta**.

**If you try it:** [tell us what works and what doesn't](https://github.com/marceli1404/sloth-mode/issues/30). A reproducible failure report is more valuable than an inflated claim of perfection.

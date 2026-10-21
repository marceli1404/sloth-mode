# Sloth Mode beta — release validation checklist

This is the **manual evidence gate** for publishing stable Windows/Linux binaries. GitHub CI passing is necessary but **not** proof that real devices, models, display scaling or OAuth work. Do not mark a step passed without recording the machine, version and observed result.

## Before testing

- Confirm the exact git commit and artifact SHA-256. Use test accounts and fabricated task data only.
- Back up existing Sloth Mode local notes. Never include original pairing keys, access tokens, OAuth secrets, API keys or private model data in bug reports.
- Start with a clean install on a disposable machine or profile, then separately test an upgrade that preserves existing notes.
- Never download a multi-gigabyte model without consent. Downloaded model weights and native runtimes have independent third-party licensing.
- Keep ports bound to loopback. **Do not open localhost or the connector port to the public internet.**

## Platform and installer matrix

| Check | Windows x64 | Linux x64 | Linux ARM64 | Evidence needed |
| --- | --- | --- | --- | --- |
| Clean install without admin/root or account | Pending | Pending | Pending | OS build, package, installer log |
| Setup/pairing window or launcher works | Pending | Pending | Pending | UI screenshot, no secrets |
| Existing notes survive upgrade/restart | Pending | Pending | Pending | Before/after fictional note counts |
| Connector starts and stops cleanly | Pending | Pending | Pending | Log excerpt with sensitive values redacted |
| Unauthenticated localhost requests rejected | Pending | Pending | Pending | HTTP status, not token |
| Guest browser pairing works without login | Pending | Pending | Pending | Functional checklist |
| Account linking, revocation and reconnect | Pending | Pending | Pending | Account test identity, screenshots |
| Model download/checksum and first real inference | Pending | Pending | Pending | Quantization, runtime hash and timings |
| Offline inference after initial download | Pending | Pending | Pending | Network-off verification |
| Malformed/oversized archive safely refused | Automated | Automated | Cross-build only | CI & test log |

For Linux, check a systemd user service where supported; also check a desktop launcher on a representative distribution. Verify CPU inference first. Treat Vulkan/GPU execution as **separate** testing and record actual detected GPU and driver details, not a claimed speedup.

For Windows, verify the installer and setup window at **100%, 125%, 150% and 200% display scaling**. Check the copy/paste link-code controls and system tray behaviour. For Linux, also verify X11/Wayland launch paths where applicable.

## Model selection

1. Record detected CPU, available RAM and GPU/driver information without uploading hardware serial numbers.
2. Confirm the proposed Gemma quantization reflects the system profile and allows manual override.
3. Confirm model and llama.cpp archives are downloaded over HTTPS; mismatched hashes, oversized files, corrupt archives and unsafe members must be rejected.
4. Confirm an existing runtime and local task store are not overwritten during a failed install.
5. Run a **real** local prompt, record that an answer was generated, then disconnect the network and repeat. Never substitute the deterministic website sample for this test.
6. Measure cold-start time, memory use and response latency on the actual machine; do not claim performance based on build results.

## Website and optional account

- Visit the public homepage, dashboard and Integrations page **without** being signed in. Quiet Lab and paired local-connector features must still work.
- Verify clicking **Sign in** authenticates through an approved Google, GitHub or email provider, and returns to an authenticated dashboard. Guest use must never redirect unexpectedly into mandatory registration.
- Explicitly create an account-tracked task, sign out, and check that guest/local notes remain accessible while cloud task data is not exposed.
- Verify another signed-in browser can access the same **account-saved** task; test revocation and session expiry. Check that unrelated accounts cannot retrieve it.
- Confirm linking a PC is distinct from browser pairing, and revoking an account-linked device does not delete local notes.
- Confirm remote Gemma inference and general-purpose note syncing are **not advertised as available** until deployed and security tested.
- Test layouts at widths 320, 375, 768, 1024, 1440, 1920 and 2560px; repeat at 150% and 200% zoom. Check for clipping, focus visibility, readable controls, keyboard navigation, reduced-motion preference and loading feedback.
- Verify the published promo video is the approved 1080p/24-fps cut and doesn't claim unsupported cloud features.

## Opt-in real-Gemma smoke test and reproducible timing

The repository now includes a **real inference** smoke test that is deliberately
**skipped in CI by default**. It requires an already installed, verified Gemma
GGUF and llama.cpp runtime; it does not download files, create an account,
or read personal task notes. It sends three fabricated notes and checks the
returned task IDs, structured fields, and response validity.

Linux:

```bash
cd connector
SLOTH_REAL_GEMMA_TEST=1 SLOTH_TEST_BACKEND=cpu go test -run '^TestRealGemmaTriageOptIn

- [ ] All required Windows/Linux/cloud tests and CodeQL jobs green on the *release commit*.
- [ ] No untriaged high-severity code-scanning findings or known exploitable vulnerabilities.
- [ ] Independent reviewer checks runtime extraction, update verification, OAuth account boundaries and loopback pairing.
- [ ] Real platform tests above documented with verifiable redacted evidence.
- [ ] Fresh Windows/Linux deliverables built from the identified commit, accompanied by SHA-256 checksums.
- [ ] Release archive contents and source/provenance match; no secrets, customer data or private signing keys included.
- [ ] Offline signing key kept out of CI; signed updater manifest points to this release and verifies the exact file size/hash.
- [ ] Versioned release notes state remaining limitations clearly, including optional sign-in and guest privacy.

## Test report format

Copy this into a GitHub issue comment or the release checklist issue:

```text
Commit / artifact SHA-256:
Tester (handle only):
OS / version / architecture:
CPU / RAM / GPU / driver (omit identifiers):
Display scaling / browser / zoom:
Guest pairing: PASS / FAIL / NOT TESTED
Local data preserved: PASS / FAIL / NOT TESTED
Model variant and verified hash:
Real inference (online then offline): PASS / FAIL / NOT TESTED
Account/device tests: PASS / FAIL / NOT TESTED
UI clipping/accessibility: PASS / FAIL / NOT TESTED
Logs/screenshots (redacted):
Problems + reproduction steps:
```

A failed or untested step is a release blocker only when it materially affects a claimed supported feature. Never manufacture evidence or close a security finding based on a CI green check alone.
 -count=1 -v .
```

Windows PowerShell:

```powershell
cd connector
$env:SLOTH_REAL_GEMMA_TEST = '1'
$env:SLOTH_TEST_BACKEND = 'cpu'
go test -run '^TestRealGemmaTriageOptIn

- [ ] All required Windows/Linux/cloud tests and CodeQL jobs green on the *release commit*.
- [ ] No untriaged high-severity code-scanning findings or known exploitable vulnerabilities.
- [ ] Independent reviewer checks runtime extraction, update verification, OAuth account boundaries and loopback pairing.
- [ ] Real platform tests above documented with verifiable redacted evidence.
- [ ] Fresh Windows/Linux deliverables built from the identified commit, accompanied by SHA-256 checksums.
- [ ] Release archive contents and source/provenance match; no secrets, customer data or private signing keys included.
- [ ] Offline signing key kept out of CI; signed updater manifest points to this release and verifies the exact file size/hash.
- [ ] Versioned release notes state remaining limitations clearly, including optional sign-in and guest privacy.

## Test report format

Copy this into a GitHub issue comment or the release checklist issue:

```text
Commit / artifact SHA-256:
Tester (handle only):
OS / version / architecture:
CPU / RAM / GPU / driver (omit identifiers):
Display scaling / browser / zoom:
Guest pairing: PASS / FAIL / NOT TESTED
Local data preserved: PASS / FAIL / NOT TESTED
Model variant and verified hash:
Real inference (online then offline): PASS / FAIL / NOT TESTED
Account/device tests: PASS / FAIL / NOT TESTED
UI clipping/accessibility: PASS / FAIL / NOT TESTED
Logs/screenshots (redacted):
Problems + reproduction steps:
```

A failed or untested step is a release blocker only when it materially affects a claimed supported feature. Never manufacture evidence or close a security finding based on a CI green check alone.
 -count=1 -v .
```

Set `SLOTH_TEST_BACKEND=vulkan` to test the separately installed Vulkan
runtime. Record the printed cold-start and triage durations, tested quant,
OS, CPU, RAM, GPU/driver, and whether every synthetic task received a
plausible category/priority/next action. Repeat **after turning off network**
to confirm offline operation; local loopback should still work.

These durations measure **model execution**, not human time saved. To make a
time-savings claim in the DEV submission, separately stopwatch a human
performing the *same* task-set categorisation and compare the results.
Disclose sample size, setup/download time exclusions, and failure cases.
If a real run has not been completed, leave timings and savings as
**not yet measured**. Never use the Quiet Lab deterministic demo as
evidence of real Gemma execution.

## Signed release gate

- [ ] All required Windows/Linux/cloud tests and CodeQL jobs green on the *release commit*.
- [ ] No untriaged high-severity code-scanning findings or known exploitable vulnerabilities.
- [ ] Independent reviewer checks runtime extraction, update verification, OAuth account boundaries and loopback pairing.
- [ ] Real platform tests above documented with verifiable redacted evidence.
- [ ] Fresh Windows/Linux deliverables built from the identified commit, accompanied by SHA-256 checksums.
- [ ] Release archive contents and source/provenance match; no secrets, customer data or private signing keys included.
- [ ] Offline signing key kept out of CI; signed updater manifest points to this release and verifies the exact file size/hash.
- [ ] Versioned release notes state remaining limitations clearly, including optional sign-in and guest privacy.

## Test report format

Copy this into a GitHub issue comment or the release checklist issue:

```text
Commit / artifact SHA-256:
Tester (handle only):
OS / version / architecture:
CPU / RAM / GPU / driver (omit identifiers):
Display scaling / browser / zoom:
Guest pairing: PASS / FAIL / NOT TESTED
Local data preserved: PASS / FAIL / NOT TESTED
Model variant and verified hash:
Real inference (online then offline): PASS / FAIL / NOT TESTED
Account/device tests: PASS / FAIL / NOT TESTED
UI clipping/accessibility: PASS / FAIL / NOT TESTED
Logs/screenshots (redacted):
Problems + reproduction steps:
```

A failed or untested step is a release blocker only when it materially affects a claimed supported feature. Never manufacture evidence or close a security finding based on a CI green check alone.

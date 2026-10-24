# Real Gemma CPU run — 21 October 2026

Historical evidence for source `147b0c90579a949dfd47df7dadfc3023e92c5145`, using the isolated validation settings described below. This is not a new model run or a performance claim for the 24 October installer.

The real model generated a consistent plan for all three synthetic smoke-test
notes. This is a small functional check, not a broad quality benchmark or a
measurement of effort saved. Network-off behavior remains unverified.

## Environment and artifacts

- Windows 11 Pro, build 26300; 6 CPU cores / 12 logical processors;
  31.92 GiB system RAM.
- Portable Go 1.27.3 Windows amd64, checked against its official SHA-256.
- Gemma 3 4B instruction-tuned GGUF, Q4_K_M, 2,489,758,112 bytes.
- Model SHA-256:
  `4996030242583a40aa151ff93f49ed787ac8c25e4120c3ae4588b2e2a7d1ae94`.
- Official llama.cpp b11799 Windows CPU archive, downloaded and verified by
  `installAI` against the GitHub release asset digest before extraction.
- Extracted `llama-server.exe` launcher SHA-256:
  `72109706e683889b95ec6f66006dee7e527a32c5e41859e0af3d4ea1fc6e3c31`.
  This is the launcher hash, not the archive or runtime DLL hash.
- CPU only: zero GPU layers, four threads and four batch threads.

The test used an isolated copy of the connector and a separate data directory;
no existing personal notes or pairing data were read. The source includes the
runtime-selection fix from PR #13 and the prompt/consistency changes in PR #14.
Only the isolated copy's server arguments were capped to four threads; the
production default thread selection was unchanged. The first model/runtime
setup took 32.96 seconds on this connection; it is not a portable expectation.

## Latest observed output

| Original note | Priority / category | Summary | Next action |
| --- | --- | --- | --- |
| Pay electricity bill tomorrow. | now / admin | Pay electricity bill | Pay the electricity bill |
| Read a chapter of a novel when I have time. | later / learning | Read a novel chapter | Read a chapter of the novel |
| I finished tidying my desk yesterday; no more action is needed. | no_action / other | Tidied desk | No action needed |

Model startup: **2.766 s**. Generation request: **14.808 s**.
Test body: **17.95 s**; Go package run: **19.624 s**. Files were already
downloaded, and this was not a cold boot or a measured human comparison.

`TestRealGemmaTriageOptIn` exited successfully. It checked the response schema,
all three source IDs, and the expected priorities. The normal validator also
rejects an actionable task paired with an obvious “No action needed” result.
Category labels are illustrative model choices; this test does not prove their
quality or consistency across unrelated inputs.

## What the real test caught

The initial generic prompt returned valid JSON but classified the bill as
`later` and optional reading as `no_action`. Explicit priority rules corrected
those decisions, but a subsequent run still gave the urgent bill a contradictory
“No action needed” instruction. The final prompt adds different-domain examples,
and validation now rejects that contradiction before anything is saved. The
latest real run passed the stronger check with actionable suggestions.

## Reproduce

On a machine with verified model/runtime assets already installed:

```powershell
cd connector
$env:SLOTH_REAL_GEMMA_TEST = '1'
$env:SLOTH_TEST_BACKEND = 'cpu'
go test -run '^TestRealGemmaTriageOptIn$' -count=1 -v .
```

The test uses fabricated notes and never downloads files. The run above used
an isolated `LOCALAPPDATA` override and validation-only thread caps. Local raw
test logs are retained in the task workspace under `validation/`; they are not
required to build the application.

## Still pending

- Repeat with network access actually disabled before claiming verified offline
  operation. The successful run above was online; the test did not alter the
  user's network settings.
- Record the real browser/connector workflow and validate clean installation,
  pairing and upgrade preservation on the intended demo profile.
- Measure human sorting time and output usability before claiming time saved.
- Windows `go vet` flags the pre-existing native clipboard pointer conversion
  at `tray_windows.go:136`; native tests/build pass, but this warning has not
  been resolved here.

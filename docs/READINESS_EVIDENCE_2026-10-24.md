# Additional competition-readiness work — 24 October 2026

The strongest remaining gap was actual browser/model evidence. The [new evidence package](evidence/2026-10-24/README.md) contains an uncut 41-second real CPU recording, actual API observations, a new predefined corpus, raw baseline/candidate outputs and an independently exercised preparation script.

Product changes:

- Restore explicit triage rules distinguishing urgent work, unfinished optional tasks, completed tasks and reference notes. Reject actionable priorities paired with obvious no-action instructions.
- Detect recognised English timing phrases omitted from a model summary/action and point the user to the original note IDs. Do not infer dates or rewrite priorities. The warning is shown through the existing native model-status UI and included in native exported briefings.
- Validate the entire inference-input snapshot under the storage lock before applying results. A changed, completed, deleted or already classified input rejects the whole batch without partial persistence. Unrelated new notes are allowed.

Fresh CPU result: baseline9/12 versus candidate12/12 expected priorities; timing retention0/9 versus1/9. The timing gap is disclosed, and all seven recorded notes with omitted timing reach the review warning. This is narrow functional evidence; no general accuracy, security, offline or human-effort claim is made.

Local connector tests/vet and Windows builds pass. Native installer tests remain part of the build. The real browser recording verifies pairing/signing, actual model output, original-text preservation, editing/reset, export and explicit reconnect after reload. The recording was decoded successfully after H.264 transcoding without timing changes; no cuts or speed changes. Production web source and deployment are unchanged; the recording uses a frozen, labelled local fixture UI. Clean-profile install and machine-wide offline tests remain separate.

## Published candidate and article verification

Implementation/evidence commit `f73cfc4a01b6cbb672cac174aaa004dbe6d0b6fe` passes [Linux, Windows and cloud CI](https://github.com/marceli1404/sloth-mode/actions/runs/40091059467) and [CodeQL](https://github.com/marceli1404/sloth-mode/actions/runs/40091059484).

The unsigned `SlothMode-Windows-Preview` artifact `12558692342` was downloaded without execution. ZIP SHA-256 `ce8d325bb6505556a6067185449d40e024cf716a434a2553dba76fcf29b55d9c` matches GitHub's digest. The extracted installer is PE32+ x64, 14,043,648 bytes, SHA-256 `251fb12efa9de51469176ddf79a9d130a9f1b9238f84e302c51a500e3baf0879`, matching its included checksum. This verifies artifact identity, not clean installation or security acceptance.

The [existing DEV article](https://dev.to/marceli/sloth-mode-i-built-a-local-gemma-assistant-so-my-brain-can-be-a-little-lazier-ci4) was updated at 11:50:27 UTC on 24 October. Public API readback exactly matches `DEVTO_ARTICLE_UPDATE.md`; title, URL, tags and original publication timestamp are preserved. Chrome confirms the results image, download link, scoring table, source links and labelled promotional video render. The MP4 is a clearly labelled download rather than an unsupported inline video embed. Its public bytes match the committed file, SHA-256 `3f4a4128f571381fe9ef64b1ab3e50f401773b9f5468f8118899cb588d91cac9`.

Six copy-review passes covered the new status/export text, frozen demo UI, evidence README and article: reader path; claims against recorded inputs/outputs; existing voice and specificity; structure and useful links; sentence edits; rendered cold read and regression checks. The final pass confirmed both article images loaded, the comparison retains the timing failure, and the historical corpus remains distinct. Other production app pages were unchanged and excluded. No owner facts or unverified testimonials were added. Clean installation, independent offline testing and broad model quality remain evidence gaps.

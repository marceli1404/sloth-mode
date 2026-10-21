# DEV Week 3 submission checklist

Use this list before copying the draft to DEV. The contest deadline is
**25 October 2026 at 11:59 p.m. PDT**. The exact entry period, required
elements, and eligibility rules are on the [official rules page](https://dev.to/page/hacktoberfest-week3-2026-10-19-contest-rules).

## Entry and article

- [ ] Confirm the entrant, account, team, and project meet the official rules.
- [ ] Confirm the project was created and completed during the 19–25 October
  entry period. The repository's initial commit is dated 20 October; that alone
  does not establish the full project history or eligibility.
- [ ] Publish one entry using the official template and tags
  `devchallenge` and `hf26challenge`.
- [ ] Keep the opening centered on the specific task-list burden and Sloth
  Mode's distinct workflow: a general task brain dump becomes a reviewable
  plan on the user's computer, with suggestions tied to original note IDs.
- [ ] Explain who the workflow helps, what effort it is meant to remove, and
  why local open-weight AI matters for this private data.
- [ ] Link the working demo and the code. Remove all editor notes and
  bracketed placeholders.

## Real demo and evidence

- [ ] Test the exact commit and hardware used for the demo. CI passing alone
  does not establish model inference on a real machine.
- [ ] PR #13's runtime fix has CI evidence, but no hardware inference proof is
  recorded yet. Keep the real Gemma result pending until a machine run exists.
- [x] Run Gemma 3 4B through the local engine with fictional notes. Recorded in
  [CPU_DEMO_REPORT.md](CPU_DEMO_REPORT.md); browser pairing still needs validation. Record
  OS, CPU, RAM, GPU/driver if used, quantization, runtime/backend, cold or warm
  state, observed result, and any failure.
- [ ] Confirm every output maps to an existing source note and that an
  incomplete, duplicate, or invalid result is rejected.
- [ ] Show one note edit and confirm the old AI annotations clear.
- [ ] Test after disconnecting the network before saying the core flow works
  offline. If this repeat has not passed, say offline validation is pending.
- [ ] Make the demo link accessible without signing in. Keep it clear that the
  public Quiet Lab is a deterministic rules preview, not a model result.
- [ ] Use fictional task content only; remove pairing codes, account details,
  machine usernames, and private paths from screenshots or video.
- [ ] Link or embed the approved promo asset only after its public link and
  final cut are verified.

## Lazy payoff and partner category

- [ ] Either omit “The Lazy Payoff” or fill it with an actual comparison using
  the same notes: manual sorting time, model time, cold-start treatment,
  usable-result count, sample size, and setup/download time reported separately.
- [ ] Do not claim time saved based only on model latency, code, CI, or the
  deterministic Quiet Lab preview.
- [ ] Claim **Best Use of Gemma** only if the article and demo show Gemma doing
  meaningful work in Sloth Mode.
- [ ] Do not list SerpApi or another partner category unless the submission
  demonstrates that partner's technology in a meaningful project workflow.

## Final check

- [ ] Confirm source, demo, and any video still work in an unsigned-in browser.
- [ ] Re-read the official contest rules and confirm the deadline in PDT.
- [ ] Remove this checklist from the DEV post; it is a local review aid.

## Evidence status at draft time

- Repository and website are public; source describes the local connector and
  model path.
- The public Quiet Lab is explicitly illustrative and deterministic.
- Release validation still calls for a real first inference and a separate
  network-off repeat on physical hardware.
- A real CPU engine test passed on three fabricated notes; priorities and next
  actions were checked after correcting issues found in earlier runs.
- No human-versus-model time-saving result or network-off result is recorded.

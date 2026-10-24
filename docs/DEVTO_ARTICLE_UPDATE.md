*This is a submission for the [Hacktoberfest Open-Source AI Challenge Week 3: Slow and Steady](https://dev.to/challenges/hacktoberfest-week3-2026-10-19).* 🦥

Pay the electricity bill tomorrow. Read a chapter when I have time. I already tidied the desk.

A to-do list can hold all three. I still have to decide what matters, what can wait, and what no longer needs doing. On a tired day, organising the list becomes another task on the list.

I built **Sloth Mode** to make that first decision smaller: put rough notes in one place, let a local Gemma model suggest a plan, then choose what to do with it. The private notes can stay on the computer doing the thinking.

## What I Built

Sloth Mode is a task companion for the loose ends you capture before you know what to do with them. Gemma suggests a priority, category, short summary and next action for each selected note. Each suggestion points back to its original note, so you can check the wording before accepting it.

It is for someone who wants help starting without first maintaining an elaborate productivity system. The model drafts the organisation; the person can edit, ignore or act on it. Sloth Mode does not send messages or execute the suggested tasks.

There are three ways to explore it:

| Route | What happens | What you need |
|---|---|---|
| **Quiet Lab** | An illustrative, rules-based preview of the interaction | A browser; no account or download |
| **Native Gemma** | Real Gemma 3 GGUF inference through a local Go companion and `llama.cpp` | Windows preview or experimental Linux build; separate model/runtime download and explicit browser pairing |
| **Browser Gemma beta** | An experimental Gemma 4 E2B path using WebGPU/LiteRT-LM | Compatible hardware/browser and an approved model download of about 2 GB; wider compatibility is still unverified |

An optional account holds a **separate** list of tasks you explicitly save there. Signing in does not silently upload the local inbox or grant access to the companion.

## Demo

**[Try the guest workspace →](https://sloth.marceli.dev/dashboard)**

Start with Quiet Lab and its sample notes. It is the quickest way to understand the interaction, and it is labelled as a preview: **it does not call Gemma**.

### Watch the actual local model run

**[Download the uncut 41-second recording (MP4, 1.7 MB)](https://raw.githubusercontent.com/marceli1404/sloth-mode/f73cfc4a01b6cbb672cac174aaa004dbe6d0b6fe/docs/evidence/2026-10-24/real-browser-gemma-uncut.mp4)**

![Actual browser results from four fictional notes: two Do next suggestions, one Save for later and one No action, with a warning to check timing in the original notes.](https://raw.githubusercontent.com/marceli1404/sloth-mode/f73cfc4a01b6cbb672cac174aaa004dbe6d0b6fe/docs/evidence/2026-10-24/browser-results.png)

The recording shows a browser paired with the real Go companion, Gemma 3 4B Q4_K_M running through `llama.cpp`, and the resulting suggestions. It then checks editing, export and reconnecting after a reload. Clicking Plan to seeing results took **25.063 seconds**, including model loading and UI polling, with four CPU threads and zero GPU layers. The files were already installed.

The lab report due before 4 PM today and the borrowed book due tomorrow both become Do next. Optional reading becomes Save for later; an already submitted form becomes No action. Gemma still drops timing from the first two suggestions. The status points to their original notes so the person reviewing the plan can catch it.

This uses a labelled local copy of the workspace, fictional notes and the production pairing/planning handlers. It is actual model output, with no mocked planning responses, cuts or speed changes. It does **not** demonstrate a clean installation, model download or machine-wide offline operation. The [recording notes, raw responses and reproduction steps](https://github.com/marceli1404/sloth-mode/blob/f73cfc4a01b6cbb672cac174aaa004dbe6d0b6fe/docs/evidence/2026-10-24/README.md) make that boundary explicit.

The earlier promotional overview remains available:

{% youtube GnaEFKQW9Vw %}

*This is a promotional overview, separate from the real-model recording above.*

### Trying the native path

1. Get the **[refined Windows preview and SHA-256](https://github.com/marceli1404/sloth-mode/actions/runs/40114869780)** under **Artifacts → SlothMode-Windows-Preview**. GitHub may require a login. This is an unsigned candidate from PR #34; you can also build from that source.
2. Install the companion for your Windows account. Open **Setup & Pairing**, then pair the browser you trust.
3. Separately approve the Gemma model and runtime download. Use fictional notes for your first plan and review the suggestions.

The native model download is several GB, depending on the quantization. An account and Ollama are not required. The [judges' quickstart](https://github.com/marceli1404/sloth-mode/blob/133f1d1d0c4fe97a1a54b405a89a0d795fcd9bf8/docs/WEEK3_JUDGES_QUICKSTART.md) includes an opt-in real-model test for a machine with those files installed.

On 24 October I also refined the installer. Its old preview clipped text and made it too hard to see what would happen next. The [new candidate in PR #34](https://github.com/marceli1404/sloth-mode/pull/34) separates installation, browser pairing and the optional model download, with readable status and a layout that fits the monitor. The linked candidate passes Linux, Windows and cloud CI plus CodeQL checks. A full clean-profile installation and pairing run for this candidate is still pending.

![Sloth Mode's refined native Windows installer, with three setup steps, an optional sign-in launch checkbox and a primary Install & open button. The window is labelled as a layout preview with installation actions inactive.](https://raw.githubusercontent.com/marceli1404/sloth-mode/ed8d030a18973799ae40392cc7c97f0f2d34009a/docs/images/installer-preview.jpg)

*The actual native UI in a safe layout fixture. This image demonstrates the interface, not a completed installation or inference run.*

## Code

**[github.com/marceli1404/sloth-mode](https://github.com/marceli1404/sloth-mode)**

The Go companion is in `connector/`, the Windows installer in `installer/`, the browser workspace in `web/`, and the reference account service in `cloud/`. Application code is MIT-licensed; Gemma weights and the runtime retain their own terms.

The [first commit](https://github.com/marceli1404/sloth-mode/commit/0a51feca7a22a0b3ce06a5585740c8219f842e70) is dated 20 October 2026, inside the Week 3 build window. The [installer refinement and evaluation files](https://github.com/marceli1404/sloth-mode/pull/34) are reviewable separately from main. The live website is deployed separately too; cloning the reference account Worker does not recreate every production account binding.

## How I Built It

### Keep the private work on the computer

The native path is straightforward:

```text
Browser workspace
  -> explicit pairing with 127.0.0.1
  -> Go companion reads selected local notes
  -> Gemma 3 GGUF runs through llama.cpp
  -> validate structured suggestions and original note IDs
  -> show the plan for a person to review
```

Model weights and runtime archives are downloaded separately after approval and checked against expected hashes. The local service binds to loopback. A cloud sign-in is not a replacement for pairing with that service.

The validator checks IDs and structure before accepting a plan. It can reject an invented or duplicate note ID. It cannot prove that a fluent summary kept every important detail.

### Valid JSON is not a good plan

The first three-note run made the bill less urgent and treated optional reading as requiring no action. Explicit priority rules and examples helped, but another run paired an actionable bill with “No action needed.” That was a useful reminder to check both the label and the proposed action.

On 24 October I froze **12 new fictional notes and their expected priorities before inference**, then ran the baseline and candidate once each with the same existing model, four CPU threads, zero GPU layers and temperature zero.

| Check | Baseline | Candidate |
|---|---|---|
| Matches the predefined priority | 9/12 | 12/12 |
| Repeats recognised timing in the summary or next action | 0/9 | 1/9 |

Restoring explicit priority rules helped this small set. The second row is the unresolved problem: correct labels can hide missing details. A limited English phrase check now flags actionable suggestions that omit recognised timing and points back to their source notes. All **seven** such omissions in the candidate's recorded output trigger that warning. It does not calculate dates, understand every deadline or repair the model's wording.

The [frozen inputs](https://github.com/marceli1404/sloth-mode/blob/f73cfc4a01b6cbb672cac174aaa004dbe6d0b6fe/docs/evidence/2026-10-24/inputs.json), [raw candidate output](https://github.com/marceli1404/sloth-mode/blob/f73cfc4a01b6cbb672cac174aaa004dbe6d0b6fe/docs/evidence/2026-10-24/candidate-results.json) and [scoring method](https://github.com/marceli1404/sloth-mode/blob/f73cfc4a01b6cbb672cac174aaa004dbe6d0b6fe/docs/evidence/2026-10-24/README.md) are public. Categories were not scored. These 12 examples are functional checks, not a representative accuracy benchmark or a measurement of effort saved.

The older evaluation is a different 12-note set: its baseline matched 12/12 priorities, while a timing-preservation experiment fell to 11/12. Those [historical inputs and outputs](https://github.com/marceli1404/sloth-mode/tree/564f4a3dd5364cfe191eb22b3dcd40cc5a4be5ba/docs/evaluation) remain published, including their missing prompt snapshot. I keep them separate from the new frozen comparison.

### Keep suggestions attached to the right draft

A person can edit a note while a local model is still thinking. Applying the old answer to the new wording would make the plan misleading. The companion now checks the entire input snapshot under the storage lock before saving suggestions. If any input was edited, completed, deleted or already classified, it rejects the batch without saving a partial plan. Capturing an unrelated new note is allowed. Regression tests cover those cases and confirm that rejected batches leave the notes file unchanged.

### Setup deserves the same care

A local model is not useful if getting it running feels like another project. The installer refinement uses explicit font roles, consistent backgrounds and DPI-aware fitting. Installation and update work run off the UI thread, failed replacement can restore the previous app, and starting a process alone does not count as success: the installer checks the companion's local health response.

The new candidate passes **12 installer tests**, connector tests, Windows builds and static analysis. Text fitting and window bounds are checked at multiple scales; the native layout, Tab focus and Escape were also inspected. Those checks do not substitute for installing the candidate on a clean profile. The [verification report](https://github.com/marceli1404/sloth-mode/blob/f7eccd89470a16b604fe50e2c0c9010cc0f470fa/docs/INSTALLER_REFINEMENT_2026-10-24.md) distinguishes them.

## Why Does Open Innovation Matter?

For a private task inbox, where the model runs matters. Open-weight Gemma and `llama.cpp` let the companion process the notes locally, without a hosted inference account or a per-prompt API bill. There are still costs: memory, storage, electricity and the initial download.

It also lets me inspect the prompt, change quantization and runtime settings, and repeat the same fictional cases. The deadline failure above is visible in the published output. Someone else can study it or change the handling rather than trusting a screenshot of the best result.

After downloading the model and runtime, native inference can run without an internet connection. Earlier Windows previews have an [owner-reported successful offline/restart/persistence check](https://github.com/marceli1404/sloth-mode/issues/8#issuecomment-6328830871). That is one person's report on older builds; the new installer still needs its own complete acceptance run. The hosted website and account services continue to need connectivity.

## The Lazy Payoff

The effort I want to remove is the sorting between “I should remember this” and “here is one thing I can start.” Gemma drafts priorities and next actions so I have something small to review.

I have not measured minutes saved per day or run a multi-user study. My evidence is narrower: a real browser-to-model recording, predefined fictional examples with their failures, a working guest preview, and owner experience on one Windows environment. A useful next study would compare manual sorting with assisted sorting on the same fictional backlog, including review time, mistakes and setup cost.

This is still a public beta. Signed installers, independent native security review, wider Linux/GPU coverage, a clean-install recording and independent offline acceptance remain unfinished. Browser Gemma is experimental. The [release checklist](https://github.com/marceli1404/sloth-mode/blob/main/docs/RELEASE_VALIDATION.md) keeps those gaps visible.

## Prize Categories

**Best Use of Gemma.** Gemma does the native task planning, with its suggestions tied to the notes that prompted them. The browser beta explores a separate Gemma execution path; the no-download preview is clearly labelled as illustrative.

If that sounds useful, [try the guest workspace](https://sloth.marceli.dev/dashboard) or [report what breaks](https://github.com/marceli1404/sloth-mode/issues/30) using fictional notes. I wanted a tool that makes the next task feel smaller. That is the test I want Sloth Mode to pass.

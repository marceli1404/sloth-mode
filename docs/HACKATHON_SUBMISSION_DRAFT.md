---
title: Sloth Mode turns a private brain dump into a plan you can review
published: false
tags: devchallenge, hf26challenge, gemma, opensource
---

<!-- Working draft; remove this note before publishing. -->
Replace every bracketed item with verified material or remove it before
posting. The public Quiet Lab preview uses deterministic rules; it is not a
Gemma run.

**Challenge:** [Hacktoberfest Open-Source AI Challenge: Slow and Steady](https://dev.to/challenges/hacktoberfest-week3-2026-10-19)

---

## What I Built

Capturing a task is one step. Deciding what deserves attention is another.
I built Sloth Mode for that second step: drop rough task notes into one private
place, then ask a local Gemma model for a short plan you can review.

Sloth Mode is a local-first task companion for Windows and Linux. It groups
notes into **Do next**, **Save for later**, and **No action**, and suggests a
short summary and next action for each note. It is for the loose tasks I capture
as they occur, rather than an email client or an agent that acts on my behalf.

The part I care about is keeping the suggestion tied to its source. Gemma must
return the ID of each existing note; the connector checks every ID and result
before saving anything. If the response is incomplete or points to an invented
or duplicate note, the result is rejected. The notes themselves stay yours to
edit, complete, or delete.

## Demo

- **Website and interactive preview:** https://sloth.marceli.dev
- **Recorded CPU results:** https://github.com/marceli1404/sloth-mode/blob/improve/hackathon-judge-experience/docs/CPU_DEMO_REPORT.md

The public Quiet Lab preview is an illustrative rules-based walkthrough. It
does not call Gemma. The real model runs on the paired computer after the user
approves the model and runtime download. The linked CPU report records actual model output on fictional notes; the
public browser preview remains rules-based. A video of the paired browser
workflow has not been recorded yet.

## Code

- **Repository:** https://github.com/marceli1404/sloth-mode
- **License:** MIT for Sloth Mode application code. The model and runtime have
  separate terms; their weights are not included in the repository.

## How I Built It

The browser pairs explicitly with a background connector on the same computer.
The connector stores local notes and includes the managed path for running
Gemma 3 4B in GGUF format through `llama.cpp`. The browser sends model requests to the connector over a
loopback-only connection; task notes do not pass through Sloth Mode's servers.

```text
Browser workspace
    └─ explicit pairing → local connector (127.0.0.1)
                            ├─ local note store
                            └─ Gemma 3 GGUF via llama.cpp
                                └─ validate suggestions against note IDs
```

Model weights and the native runtime are downloaded separately after user
approval and verified by checksum. The first model download is about 2.5 GB.
A real CPU smoke test passed on Windows 11 using Gemma 3 4B Q4_K_M through
llama.cpp, with four CPU threads and no GPU layers. Startup took 2.766 seconds
and generation took 14.808 seconds for three fictional notes. The bill due
tomorrow was marked Do next, optional reading was saved for later, and a
completed task required no action. This small check does not establish general
quality or time saved. A separate network-off test remains pending.

An account is optional. A signed-in user can explicitly save account-tracked
tasks, but local notes are not silently synced. Remote Gemma execution is not
implemented.

## Why Does Open Innovation Matter?

Open-weight Gemma makes local inference possible, and the MIT-licensed
application code lets people inspect and change how the workflow handles their
notes. For this product, that changes where a private task list has to go:
Gemma can sort it on the user's own computer instead of sending it to a hosted
model.

The local connector also means the model is not the only part the user has to
trust. The note store, inference request, and validation path are inspectable
in the repository. Model weights and `llama.cpp` keep their own licenses and
terms.

## The Lazy Payoff

The intended payoff is less manual sorting between capturing a thought and
choosing what to do next. I have not measured human-versus-model time savings.
The CPU run above reports execution time and three checked suggestions only.

## Prize Categories

**Best Use of Gemma** — Gemma is the local triage engine, not an optional
decoration. The CPU report above records its real output; the public browser
preview is separately labelled as rules-based.

---

## PRE-PUBLISH NOTES — REMOVE FROM THE DEV POST

- The official template uses the headings above; keep the submission tags
  `devchallenge` and `hf26challenge`.
- Do not present Quiet Lab's deterministic preview as model output.
- Replace or remove every bracketed placeholder. Do not invent latency,
  quality, time-saving, offline, or hardware evidence.
- Confirm project and entrant eligibility against the contest rules before
  submitting.
- The official deadline is 25 October 2026 at 11:59 p.m. PDT.

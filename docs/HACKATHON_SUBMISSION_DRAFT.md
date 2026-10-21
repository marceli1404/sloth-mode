# DRAFT — Hacktoberfest 2026, Week 3: Slow and Steady

**Review draft on a public GitHub branch. This file is NOT a published DEV submission.**
Do not publish until the verification boxes at the end are complete. The website's
Quiet Lab sample is deterministic, **not a real Gemma run**.

Suggested DEV title: **Sloth Mode: I wanted my task list to sort itself, not become another job**

Suggested tags: `devchallenge`, `hf26challenge`, `gemma`, `opensource`

Challenge: https://dev.to/devteam/join-the-hacktoberfest-open-source-ai-challenge-slow-and-steady-2850-in-prizes-across-21-3iee

---

> **Hacktoberfest Open-Source AI Challenge, Week 3 — Slow and Steady**

## What I built

I keep making lists of things I need to do, and then spending too much time
organising the lists themselves. I wanted something simpler: open a page, empty
my head into it, and get one manageable next step.

So I built **Sloth Mode**. It is a small, local-first task companion for Windows
and Linux. You can use it without creating an account.

You add rough notes in whatever order they occur to you. When you choose to run
the local model, Gemma 3 4B (a GGUF build running through `llama.cpp`) reviews
a batch and suggests which are worth doing **now**, which can wait, and which
don't require action. It also suggests one next action for each note.

I didn't want the model silently changing or deleting anything. Every result
has to point back to an existing note ID. The app rejects made-up or duplicated
IDs, missing outputs, unsupported categories and invalid responses rather than
saving a partly invented plan. If you edit a note later, its old AI labels are
cleared so an outdated suggestion isn't presented as current.

The sloth doesn't do your life for you. It does the annoying part before you
start.

## Demo and source

- **Website / interactive preview:** https://sloth.marceli.dev
- **Open-source repository (MIT):** https://github.com/marceli1404/sloth-mode
- **Official promo video:** [ADD VERIFIED PUBLIC VIDEO URL OR EMBED BEFORE PUBLISHING]
- **Real model demonstration:** [ADD REDACTED VIDEO OR TERMINAL RESULT FROM A SUCCESSFUL HARDWARE RUN]

**Important distinction:** The public Quiet Lab preview illustrates the
sorting interface with deterministic sample output. It is not a hosted Gemma
model. Actual Gemma inference happens on the user's computer after they
explicitly install the model and native runtime.

## How it works

```text
Task ideas (browser)
        |
        | explicit device pairing
        v
Sloth Mode connector on 127.0.0.1
        |
        +--> local note store
        |
        +--> user-approved Gemma 3 GGUF + llama.cpp
                    |
                    v
          structured task suggestions
                    |
                    v
          validate against source IDs
                    |
                    v
          save results locally or reject
```

The companion runs on Windows or Linux. The browser talks to a loopback-only
endpoint after explicit pairing. Model weights and native releases are
downloaded separately with checksum verification. Hardware profiling suggests
a conservative quantization, and the user can override it. CPU is the default;
Vulkan is an opt-in path that needs the right GPU drivers.

Accounts are **optional**. With a configured Cloudflare account service, a
signed-in user can explicitly save account-tracked tasks. That is distinct from
private notes on the local connector. Signing in does not silently synchronize
the local note store.

I deliberately left remote Gemma execution out of the beta. It would be easy
to put a button in the interface; building a secure remote relay is a different
problem, and I don't want to pretend it is finished.

## Where the open AI part matters

The interesting thing to me isn't just choosing a model with an open release.
It is being able to trace the whole path: source notes, a known model file,
local inference, validation, and the exact data that gets saved.

Gemma is doing real language work here: turning messy and differently worded
notes into a consistent small set of priorities and next actions. The local
runtime means the notes don't have to be sent to a proprietary hosted model
for this step. `llama.cpp` means that the same basic flow can run on ordinary
CPU hardware rather than requiring an expensive cloud GPU.

The application code is MIT-licensed. Third-party models and runtime components
have their own terms, and the model weights aren't bundled into the repository.

## Did it actually save effort?

**[DO NOT CLAIM A NUMBER UNTIL THE MEASUREMENTS BELOW EXIST.]**

The goal is to save the time spent sorting a brain dump, not necessarily to
reduce model latency to zero. My proposed comparison is simple:

1. Prepare the same ten synthetic unsorted notes for both conditions.
2. Time a person manually deciding Now / Later / No Action and one next action.
3. On a real machine, time Gemma cold start and triage. Check whether every
   source note is accounted for and whether suggested actions are usable.
4. Repeat with the network disconnected, then compare both **quality and time**.
5. Note OS, CPU, RAM, GPU/driver, quantization, failures, and sample size.

`docs/RELEASE_VALIDATION.md` includes an opt-in real-Gemma smoke test.
The reproducible results will go here once verified.

**Measured manual time:** [NOT YET MEASURED]

**Measured AI cold-start / inference time:** [NOT YET MEASURED]

**Usable results:** [NOT YET MEASURED]

**What was excluded:** Model download/setup time must be disclosed separately;
it is not fair to hide it when describing first-run effort.

## Things that are still rough

This is a **beta**, built as a new project during the challenge week. A public
interface and passing CI are not the same as real-machine release validation.
Windows signing, physical-device checks, real CPU/Vulkan Gemma runs, optional
account end-to-end checks, and independent runtime security review still
require evidence. The limitations and open gates are tracked in the repository.

I care more about making one useful, private workflow reliable than pretending
I finished every integration in a few days.

## What I'd build next

First, finish real-device model and installer tests and a reproducible
quality/time comparison. Then improve the experience around reviewing a
suggestion and quickly accepting, editing or discarding it. Only after the
local workflow is solid would I consider secure remote features.

**Slow down. Get things done.** 🦥

---

## PRE-PUBLISH CHECKLIST (remove this section from the DEV post)

- [ ] Real Windows or Linux Gemma smoke test completed, commit and hardware captured.
- [ ] Offline repeat successfully run or offline limitation disclosed.
- [ ] Link/embedded official promo video works anonymously.
- [ ] Website and source URLs work in an unsigned-in browser.
- [ ] Screenshot of real, synthetic source-to-plan task flow included.
- [ ] Measurements inserted, or all savings claims kept qualitative.
- [ ] Submission follows exact Week 3 DEV template and contest tags.
- [ ] Partner category **Best Use of Gemma** claimed only after meaningful use demonstrated.
- [ ] Account optionality, beta limitations and deterministic preview distinction retained.
- [ ] No secrets, private pairing tokens, real notes or other users' data in media.
- [ ] Publish on DEV by **25 October 2026, 11:59 p.m. PDT**.

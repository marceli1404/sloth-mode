# Sloth Mode — 90-second demo script

This script is for the real local Gemma path. The public Quiet Lab is a
deterministic rules preview and must be identified as such if it appears.

## Before recording

- Use the exact release candidate or commit that will be linked in the post.
- Pair the browser with the local connector. Install Gemma and the runtime
  beforehand; the first model download is about 2.5 GB.
- Prepare five fictional notes with a mix of urgent actions, later items, and
  reference-only material. Do not use personal or customer data.
- Confirm the connector reports the model ready, then rehearse once from the
  same initial state.
- Capture the real inference result and note the machine, model quantization,
  runtime/backend, and whether the run was cold or warm.
- Run the network-off repeat before including any offline claim. If it has not
  passed, omit the Wi-Fi-off demonstration and say validation is pending.

## Timed walkthrough

| Time | Show | Say |
| --- | --- | --- |
| 0–10 sec | Open the paired workspace with the prepared notes. | “I kept writing my to-do list twice: once to capture it, then again to sort it. Sloth Mode gives that pile a first pass on my own computer.” |
| 10–22 sec | Point to the mixed fictional notes and the local connector/model-ready status. | “These are ordinary task notes, not an email inbox. I choose when Gemma reviews them.” |
| 22–40 sec | Click **Plan with local Gemma** and let the real run finish. Keep the progress and result visible. | “The browser sends this request to the paired connector on localhost. Gemma runs through llama.cpp here.” |
| 40–58 sec | Show the three priorities, summaries, next actions, and each original note ID. | “It suggests what to do next, what can wait, and what needs no action. Every result points back to a note that already exists.” |
| 58–70 sec | Edit one fictional note, save it, and show that the old AI annotations clear. | “If I change the note, its old suggestion is cleared. The model does not rewrite or delete the source.” |
| 70–82 sec | Show the local connector or export and the generated note IDs / local-only status. If already verified, briefly show a second run with Wi-Fi off. | “The notes and inference stay on this computer. [If offline run is verified: I also repeated this with the network off.]” |
| 82–90 sec | Finish on the clean plan and show the repository / post link card. | “The public sample is only an illustration. This result came from the local Gemma run. The code and the full test conditions are in the post.” |

## Recording guardrails

- Keep the model response on screen long enough to read; do not speed up or
  replace it with an animation.
- If model setup, pairing, or inference fails, stop and fix the recording
  environment. Do not cut to the deterministic public preview and describe it
  as Gemma output.
- State the offline result only after a successful network-disconnected repeat.
- State time saved only after comparing the same notes, recording the manual
  and model timings, and assessing whether the suggestions were usable.
- Blur pairing codes, account details, paths containing usernames, and any
  non-fictional notes.
- If the approved promo video is a separate asset, label it as promotional;
  keep this clip as the proof of real model behavior.

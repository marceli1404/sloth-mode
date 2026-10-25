# Candidate link and copy review — 24 October 2026

## Scope

Reviewed the rendered local source preview and source for the marketing homepage,
the dashboard's **This computer** companion pane, and **Integrations**. The
review covers the Windows preview link and native GGUF download-size wording.
Browser Gemma estimates, runtime code, account flows, layout, and behavior were
out of scope and left unchanged.

## Six passes

| Pass | Pages | Finding |
| --- | --- | --- |
| 1. Purpose and reader path | Home, companion, integrations | The homepage offers a direct beta and Windows-preview path; the companion guides pairing and local model setup; integrations remain clearly optional. No path change needed. |
| 2. Claims and grounding | Home, companion, integrations | Replaced the fixed native-model estimate with “several GB, depending on model quality,” covering the observed Q4 and Q6 sizes. The Windows link and candidate commit now point to the verified CI candidate. Browser Gemma estimates were not changed. |
| 3. Voice and specificity | Home, companion, integrations | Shortened the setup action to “Set up local Gemma”; kept the size and approval details in the supporting sentence. Existing calm, direct wording remains. |
| 4. Structure and usefulness | Home, companion, integrations | Preserved headings, markup, and destinations. The setup button names the action; its explanatory size and approval details follow immediately below. |
| 5. Sentence-level edit | Home, companion, integrations | The concise button and explanatory sentence scan cleanly together; no other sentence needed editing for this scope. |
| 6. Cold read and regression | Home, companion, integrations | Re-read the rendered local preview and source. The Windows link remains labeled unsigned and experimental; native download size remains in the homepage explanation, the source paragraph below the setup button, and its confirmation. No new inconsistency found. |

## Representative edits

- `Install Gemma (several GB, depending on model quality)` → `Set up local Gemma`; the adjacent source paragraph explains that the approved download uses several GB, depending on model quality.
- `Download Gemma 3 (~2.5 GB)` → `Download Gemma 3 (several GB, depending on model quality)`
- The Windows preview link now targets Actions run `40133971950`; its documented candidate is `e1428d0524143dc4a68956de36c8121107b354b3`.

No additional owner input is needed for these edits. In the guest preview, the
local model setup control and its explanation remain gated until a companion is
paired; their wording and adjacency were reviewed in source without pairing a
device. This review does not establish that the Cloudflare production Worker
has been updated.

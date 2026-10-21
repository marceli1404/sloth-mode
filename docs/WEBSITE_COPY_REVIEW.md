# Website copy review — 21 October 2026

Scope: homepage `/` and guest workspace `/dashboard`, both rendered locally.
Includes Quiet Lab, account/local storage explanations, pairing/setup, and
original-note import copy in `web/index.html`. The separate live Integrations
route and authenticated production account pages are excluded because their
complete deployed source is not in this checkout; no changes to those routes.

All six required website-copy-humanizer passes completed:

| Pass | Homepage | Guest workspace |
| --- | --- | --- |
| 1. Purpose and reader path | Primary action now leads to the sample planner; local setup remains available. | Sample welcome and Quiet Lab precede account setup in document and keyboard order. |
| 2. Claims and grounding | Replaced unsupported model-switching language with inspectable prompt/validation code; clarified source-build availability. | Rules preview explicitly distinguished from actual Gemma; import never implies preservation of lane decisions. |
| 3. Voice and specificity | Kept existing calm tone; buttons name their actual destination. | Replaced generic private-AI introduction with concrete storage wording. |
| 4. Structure and usefulness | Skip link leads to the visible planner rather than the hidden illustrative walkthrough. | Four-step setup names build, pairing, download consent and first plan; model readiness updates its next step. |
| 5. Sentence edit | Removed wording suggesting Quiet Lab becomes AI after pairing; it always uses rules. | Short, direct labels distinguish account tasks, local notes, and original-note import. |
| 6. Cold read and regression | Rendered homepage reviewed; sample link reaches the local workspace planner. | One-click sample generated 3 focus / 2 parked / 1 reference; manual lane adjustment and undo passed; welcome screenshot checked. |

Representative changes:

- “Connect your local AI” → “Try a sample plan” as the primary homepage action.
- “Your tasks” → “Account tasks”; local inbox explicitly says “On this PC”.
- “Import Quiet Lab lines” → “Import original Quiet Lab notes”, with a notice
  that preview groupings and lane edits are not copied.

No measured savings, new hardware results or new video claims were added.
Real local Gemma evidence is still pending: Go, the CPU runtime and expected
model files were absent on the reviewed PC. The unsigned Windows artifact
packaging is preparation for validation, not a verified public release.

Verification limits: the local preview serves guest account status only. Actual
pairing, downloaded models, signed-in cloud CRUD and physical Windows installer
flows were not exercised. Setup status updates and hidden import wording were
also inspected in source; they are not claimed as full connector integration
tests. Screenshot: `workspace-preview.png` in the parent task workspace.

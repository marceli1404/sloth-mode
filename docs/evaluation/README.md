# Historical Gemma task evaluation

These runs happened on **21 October 2026**, before the 24 October installer refinement. They are historical source/prompt checks, not a quality score for the new installer or current main. No new model run was performed for the installer change.

## Method and provenance

The [12 fictional inputs](inputs.json), expected priorities and constraints were fixed before inference. [Freeze metadata](freeze.json) identifies source commit `147b0c90579a949dfd47df7dadfc3023e92c5145` and input SHA-256 `f69d20ed172e7531c634850ec578f12153a6e3b25549bd6bcaafe7e62b598cf2`. Execution used a Windows amd64 validation copy, Gemma 3 4B Q4_K_M, llama.cpp, four CPU threads and zero GPU layers. Runtime settings were capped only in the isolated validation copy. Notes were fictional; no personal-note access or uploads.

The corpus covers urgent and future tasks, negated urgency, completed/reference-only notes, and two examples containing quoted instructions. This is a small author-designed convenience set, not a representative benchmark or security evaluation. Categories were not given expected labels, so category accuracy is unscored.

| Run | Structured IDs returned | Priorities matching frozen expectations | Startup | Batch generation total |
|---|---:|---:|---:|---:|
| [Baseline recorded output](baseline-results.json) | 12/12 | 12/12 | 2.214s | 57.914s |
| [Deadline-prompt experiment](prompt-experiment-results.json) | 12/12 | 11/12 | 2.768s | 71.709s |

Baseline batch times: 17.680, 19.088 and 21.146 seconds. Experiment: 24.551, 19.942 and 27.216 seconds. Each is a single run; these are not controlled comparative performance results or human time savings.

## What failed

Baseline note 102 says the library book is due tomorrow. Its summary was `Return library book` and its next action `Return the library book`. The priority was correct, but the deadline disappeared. Correct IDs and valid JSON cannot establish that the meaning was preserved.

An experimental instruction to preserve timing phrases improved some summaries but did not reliably retain deadlines. Note 104, asking to call the building manager today about a broken lobby light, regressed from `now` to `later`. The baseline and experiment are both published rather than selecting only the better score. The experiment's exact prompt snapshot was not retained with these JSON files, so it is an observed historical result, not an exactly reproducible prompt comparison.

The two instruction-in-note cases stayed `later` in these runs. That is two narrow observations, not proof of general prompt-injection resistance. No network-off run or human sorting study is included. Wider paraphrases, repeated runs, deadline checks and independent user feedback remain useful next steps.

The [three-note CPU report](../CPU_DEMO_REPORT.md) records a separate focused run, with its own timings and limits. Do not combine its latency with this corpus.

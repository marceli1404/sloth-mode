# Designer website launch — 23 October 2026

Direction: an editorial botanical desk. Warm paper, forest-green panels,
large serif titles, restrained peach details, and the original sloth illustration.
Existing HTML/CSS/JavaScript retained. No external component inspiration or new
framework was used.

The deployed site was ahead of the original task checkout. This work starts from
`origin/main` ae35a26 in the isolated `designer-live` worktree, with the exact
current production page recovered before editing. Production-only account-link
and Linux wording corrections are retained in the page source.

## Changes and boundaries

- `web/index.html`: public section hierarchy, accurate sample/model/privacy copy,
  versioned CSS URL, and the previously reviewed draft/edit/connection fixes.
  Existing feature nodes, account status logic and browser-AI bootstrap remain.
- `web/workspace.css`: shared paper/forest design finish, serif title hierarchy,
  calmer hero, compact cards and controls, mobile sizing, reduced-motion support,
  and a bounded integrations background that cannot widen a mobile pane.
- `web/workspace.test.mjs`: checks the versioned stylesheet URL.
- `web/connector-api.test.mjs`: verifies authenticated hardware/preference access,
  unrelated endpoint/method restrictions and actual connection state.

Only the `PAGE` and `WORKSPACE_CSS` constants are replaced in a backup of the
current `sloth-mode-demo` Worker. All other module text is byte-identical,
including auth/task handlers, Browser Gemma runtime, routing, CSP and update
manifest. Database bindings and compatibility settings are inherited unchanged.
No model downloads, GPU inference, real account writes or installer deployment
are part of this launch.

## Verification

- 38 Node tests pass, including account service, browser AI, workspace navigation,
  public beta/branding and connector API restrictions.
- Eight inline scripts and the complete Worker bundle parse.
- 18 Playwright scenario groups pass. Coverage includes all five workspace panes
  at 320, 390, 768, 1024 and 1440px; sample/lane undo/export; synthetic signed
  pairing; failed-save recovery, duplicate prevention, in-flight draft retention,
  polling/edit preservation, lost/recovered connection, timer/search, synthetic
  account CRUD and direct `/integrations` navigation.
- `git diff --check` passes. Screenshots, browser results, exact Worker backup,
  build hashes and deployment receipts are retained in parent `validation/`.
- The local browser suite uses synthetic API responses. The deployed site will
  additionally be checked read-only for routes, assets, guest auth, styles and
  sample interactions. Authenticated production task mutation is excluded.

## Six-pass website copy review

Scope: homepage and the five workspace panes; `/integrations` is an alias for
the same integrations pane. External account-provider pages and the unmodified
native installer UI are excluded. No third-party content was crawled.

| Page/pane | Purpose | Claims | Voice | Structure | Sentences | Cold read |
| --- | --- | --- | --- | --- | --- | --- |
| Homepage | Done | Done | Done | Done | Done | Done |
| Today / Quiet Lab / account tasks | Done | Done | Done | Done | Done | Done |
| Browser Gemma | Done | Done | Done | Done | Done | Done |
| This computer / native model setup | Done | Done | Done | Done | Done | Done |
| Integrations | Done | Done | Done | Done | Done | Done |
| Explore / illustrative walkthrough | Done | Done | Done | Done | Done | Done |

1. Purpose: the homepage now explains local value and setup before optional APIs.
   The workspace keeps direct, semantic paths to all existing features.
2. Claims: “NO SERVER REQUESTS” became “INPUT STAYS IN THIS TAB”; the page does
   check account status. Removed the unsupported Sentry privacy-review wording
   and arbitrary-model-switching claim. Rules, browser Gemma and native Gemma
   are explicitly distinguished. Existing model size/download facts retained.
3. Voice: concrete local-note headings replace “One beautiful window” and
   “Web-quality design”. The sloth's calm, informal voice remains.
4. Structure: headlines, section labels, actions and data boundaries agree with
   their destinations. Optional account storage stays distinct from local notes.
5. Sentences: simplified setup descriptions, removed “ACTUALLY INTERACTIVE”,
   and clarified that pairing does not turn Quiet Lab into AI.
6. Cold read: reread all panes and homepage, checked revised text against source,
   and verified desktop/mobile rendering and working controls. No new unsupported
   claim requires owner input. Existing wider-device/offline validation tasks
   remain open in the canonical Work Ledger.

Rollback baseline: Worker version `cc8abb7f-4467-4da2-a77d-e0d22f0eef48`.
Production handlers excluding replaced constants SHA256:
`b2a6447cd8af04b43b324ea1e238b8bc6601c9543a23fab79194419214e0f70d`.


## Production result

Live: https://sloth.marceli.dev

- Source implementation commit: `35b927c`.
- Worker version: `aefcc61f-b09e-4b9b-bd0b-a53736728131`.
- Deployment: `c154ba25-0c7a-4356-a7cf-d77143b4f7b7`, 100% traffic.
- Bundle SHA256: `91ee9aa8004fb52d6041b100c158098df48f0e67f8bda535726df0608db4dff7`.
- Live verification passes: three routes, five pane layouts, fictional sample and
  undo, guest auth, anonymous task rejection, exact stylesheet hash, unchanged
  workspace/browser-AI runtime and security headers, no uncaught browser errors.
- Account bundle tests were run with checked-out asset newlines normalized to
  canonical LF for the byte-for-byte fixture on Windows; those local formatting
  changes were restored and are not part of this deployment.

The full Windows installer and personal account CRUD were not exercised by this
website deployment. These remain the existing device/account validation limits.

# Sloth Mode: previous-winner comparison

Reviewed 24 October 2026. Scope: official winner announcements and three winning write-ups. These are article/entry observations, not independently reproduced tests of their software. No claim that Sloth Mode will win follows from the comparison.

Official results: [Week 1](https://dev.to/devteam/congrats-to-the-hacktoberfest-open-source-ai-challenge-touch-grass-winners-152c), [Week 2](https://dev.to/devteam/congrats-to-the-hacktoberfest-open-source-ai-challenge-week-2-side-hustle-winners-2i33). The [current rubric](https://dev.to/devteam/join-the-hacktoberfest-open-source-ai-challenge-slow-and-steady-2850-in-prizes-across-21-3iee) gives writing the greatest weight.

| Entry | Verified award | Observed strength in its write-up | Applied to Sloth Mode |
|---|---|---|---|
| [Paper Trail](https://dev.to/mikek/paper-trail-gemma-writes-your-walk-a-field-card-then-the-screen-goes-away-3mf4) | Week 1 overall | Immediate example, focused outcome, prompt iterations, errors and remaining limits | Put a real task example first; publish both successful output and deadline/priority failures |
| [Pakka](https://dev.to/codeswithroh/kal-means-yesterday-and-tomorrow-i-built-home-bakers-a-whatsapp-order-sheet-that-never-guesses-21p3) | Week 2 Best Use of Gemma | Clear audience, original-message evidence, explicit ambiguity and separate held-out evaluation | Explain source IDs precisely; distinguish structural validation from semantic correctness; retain human review |
| [Deis](https://dev.to/muhammadmurtuzahussain/job-hunting-became-my-side-hustle-so-i-built-deis-gemma-4-tinder-for-jobs-swipe-right-gemma-3jcm) | Week 2 overall | Specific repeated chore, low-friction fictional demo, desktop delivery and visible checks | Lead with the no-account route; explain installation in three steps; provide an identified preview build |

## Installer opportunities addressed

The previously running Windows preview visibly clipped text because font selection depended on wording. Its static controls painted inconsistent backgrounds, and a fixed large client could exceed a small screen's work area. The refinement uses explicit text roles, matching GDI backgrounds, measured text, monitor-aware fitting and DPI handling. One primary install action sits beside Close; file/update actions have lower visual weight. Setup states distinguish installed files from a ready local service. Model downloads remain separate and opt-in. Prior asynchronous operations, rollback and signed-update checks were integrated with main's newer health probe.

## Remaining gap against stronger entries

Sloth Mode has one owner's Windows report and narrow synthetic CPU evidence. It does not yet have an independently recorded complete install/pair/infer video, a representative user study, proven time savings, broad Linux/GPU coverage or a signed stable package. Installer layout tests do not establish those results. The revised DEV article states the distinction and keeps the existing promotional video labelled as promotional.

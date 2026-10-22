# Sloth Mode — public beta launch guide

**Status as of 22 October 2026: public web beta live; unsigned Windows beta candidate available; stable release NOT approved.**

## What is available

Sloth Mode is a local-first productivity assistant with an optional Windows/Linux companion. The public app is at [sloth.marceli.dev](https://sloth.marceli.dev), with a [unified workspace](https://sloth.marceli.dev/dashboard). You can use Quiet Lab without signing in. Optional browser Gemma 4 E2B uses WebGPU, a user-approved multi-gigabyte download, and browser-private storage. Optional native Gemma 3 runs via llama.cpp after separately approved runtime/model downloads. Cloud account tasks are a different, explicitly saved list; local notes and native model prompts aren't automatically uploaded.

**Windows:** Current source and Windows test builds passed CI. The owner reports successful installation and model setup, restart, offline inference, and preservation of notes/pairing on one Windows machine. The security-hardened candidate originates from commit `72f2e1d180d9205f6ffe39ce1e8f0d5ce86c8384`; [Windows CI build](https://github.com/marceli1404/sloth-mode/actions/runs/39694111337) published an unsigned `SlothMode-Windows-Preview` artifact including `SlothMode-Setup.exe` and its SHA-256 checksum (GitHub artifact retention: 14 days). The owner's latest launch-readiness check is recorded under [issue #8](https://github.com/marceli1404/sloth-mode/issues/8). The artifact has **not** been signed or approved for stable distribution.

**Linux:** The source builds and automated tests pass, but independent physical x86-64/ARM64 installations and desktop/graphics environments remain in the [release validation checklist](RELEASE_VALIDATION.md). Advertise it as an experimental build, not a tested general-availability binary.

**Browser Gemma:** The UI, JavaScript route, and production deployment have been smoke checked, but full download/inference, GPU compatibility, cache integrity, network privacy verification and corruption/recovery require further real-device evidence. Do not describe it as universally available or independently security-validated.

## Windows beta installation instructions

1. Open the [verified GitHub Actions Windows CI run](https://github.com/marceli1404/sloth-mode/actions/runs/39694111337) and locate **SlothMode-Windows-Preview** under Artifacts. The user may need a GitHub login to download artifacts.
2. Extract the ZIP and verify the SHA-256 of `SlothMode-Setup.exe` against `SlothMode-Setup.exe.sha256`. The 22 October candidate checksum recorded with the build is `1cf4d195a8306c67f3be2a79bd5b294c6b98a9d022851a9768df7b18a997f985`; always prefer the checksum shipped with the actual artifact being used.
3. Test in a disposable Windows profile first. Exit an older Sloth Mode connector before installing the replacement. Back up the data under `%LOCALAPPDATA%\SlothModeNative`; protect that directory because it may include local notes, pairing material and account credentials.
4. Run the installer, open **Setup & Pairing**, then pair the local connector to the website only when you choose. Model downloads require separate consent. Guest mode does not require an account.

Do not recommend disabling Windows security protections or SmartScreen to run an unsigned preview. A signed installer with independent security validation is required for a production/stable release. Build from source if you prefer not to run an unsigned artifact.

## Known limitations and launch gates

- **[Security gate #7](https://github.com/marceli1404/sloth-mode/issues/7) OPEN.** This issue historically lists nine CodeQL path-injection/Zip Slip alerts awaiting individual trace review. Passing a CodeQL *workflow* does not establish that the alert list is empty. Source hardening from [PR #28](https://github.com/marceli1404/sloth-mode/pull/28) improves ZIP/TAR member validation but is not independent security signoff.
- **[Hardware/account gate #8](https://github.com/marceli1404/sloth-mode/issues/8) OPEN.** User reports support one successful Windows configuration. Manual display scaling, Linux device matrix, account/device revocation and cross-device authorization checks are not completed.
- **[Browser Gemma gate #18](https://github.com/marceli1404/sloth-mode/issues/18) OPEN.** WebGPU inference, OPFS cache recovery and model SHA-256 verification remain under test.
- **Signed release NOT AVAILABLE.** No stable GitHub release, SBOM/signature chain, offline-key signed installer, approved updater-manifest update or signed Linux bundle is being claimed.
- The live Cloudflare Worker includes account/session/API routes that are not faithfully reconstructed by the repository's standalone reference Worker. Do not replace production with the static site source.

## Suggested announcement (beta only)

> **Sloth Mode Public Beta:** A calmer local-first way to capture thoughts and explore task planning with optional on-device Gemma. The web workspace is live, and an unsigned Windows test build is available from GitHub Actions. Sign-in is optional, model downloads are opt-in, and your local notes aren't automatically uploaded. This is early-access software: native installer security review, wider hardware testing, and signed stable packages are still in progress.

Do not announce a "stable", "fully secure", "production certified" or "all devices supported" release. Do not change the signed stable updater channel or publish to a public package registry until the manual evidence in [RELEASE_VALIDATION.md](RELEASE_VALIDATION.md) and independent security review are complete.

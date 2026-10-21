# Windows crash and stability review — 21 October 2026

Status: source remediation and CI review only. **The user's particular crash has not been reproduced** because no Windows Event Viewer entry or redacted connector log was supplied. No production Cloudflare Worker or user task files were changed.

## Confirmed code-level defects fixed on this branch

| Component | Severity | Finding | Remedy |
| --- | --- | --- | --- |
| Windows installer and connector | High | Both created Win32 windows and consumed `GetMessageW` on Go goroutines without `runtime.LockOSThread`. This permits the goroutine to migrate to a different OS thread even though Win32 message queues belong to the creating thread. Possible symptoms: blank/unresponsive native window, disappearing UI, apparent startup crash. | Pin the UI goroutine while creating and pumping native windows. |
| Windows account link UI | Medium | Network goroutine used `SetWindowTextW` and modal `MessageBoxW` directly against HWND that could be destroyed; the code allocated a UTF-16 buffer from unbounded pasted text. | Limit account-code input to 64 UTF-16 chars, block double submissions, pass completion via `PostMessageW` and show results from owner window message callback. |
| Windows hardware scan | Medium | Background hardware probe directly changed HWND text, possibly after the window closed. | Post a window message; update text only on the UI thread. |
| Installer startup feedback | Medium | Success was reported immediately after `cmd.Start()`, without checking whether the connector stayed alive or could answer on localhost. | Bounded health probe of authenticated-origin connector route, early process exit detection, and failure details pointing to `connector.log`; background checks communicate with UI via `PostMessageW`. |
| Native llama.cpp runtime installation | High | Runtime selection looked at only 8 GitHub releases and accepted some ambiguous Windows filenames; metadata-only stable releases and incomplete nightlies could prevent setup. | Incorporate the reviewed PR #13 strict Windows CPU/Vulkan matching, up-to-32-release bounded search, and negative tests. |
| Local model inference | Medium | If llama-server exited immediately, the health polling loop waited up to 95 seconds before a generic timeout, hiding the exit reason. | Reap process once, surface early exit immediately, refer to local `inference.log`; preserve timeout and user cancellation. |
| Native diagnostic logging | Medium | A panic in the primary connector process logged only the panic value, not its stack; account heartbeat goroutine had no local recovery. | Log stack traces in `connector.log` and recover heartbeat failures without exposing local task data. |
| Product version display | Low | Installer header hardcoded 0.14; connector health called all operating systems `linux-beta`. | Use `installerVersion` and platform-neutral beta marker. |

## Unresolved / launch-blocking review items

1. **Security gate:** issue #7 records 9 remaining high-severity CodeQL findings involving runtime paths and extraction. A successful CodeQL *workflow* doesn't close these alerts. Each needs source-to-sink triage and independent security review.
2. **User-specific crash:** capture the Windows Event Viewer Application Error (faulting module and exception code only), the last redacted lines of `%LOCALAPPDATA%\SlothModeNative\connector.log`, and if Gemma is involved, the redacted `inference.log`. Never share browser pairing secrets, account linking codes, personal notes, raw prompts, device tokens or private API keys.
3. **Existing model integrity:** `detectAI` checks file sizes rather than checking the SHA-256 of a model already on disk. New downloads are SHA-256 validated, but separately corrupted/imported GGUF files may still appear ready. A future repair flow should reverify on demand and offer a safe re-download without deleting task data.
4. **Native hardware validation:** a passing Go cross-build cannot prove Windows setup, DPI scaling, tray behavior, Vulkan GPU runtime, antivirus interaction, or offline inference. Run the matrix in `docs/RELEASE_VALIDATION.md` on a disposable Windows profile and representative Linux installs.
5. **Signed release:** native packages still require signed-release provenance, hashes, stable-manifest parity, and an independent security audit before general availability.

## Reproduction sequence for latest preview

1. Back up `%LOCALAPPDATA%\SlothModeNative` (secrets stay private). Exit any previous Sloth Mode connector from the system tray before updating.
2. Use only a known GitHub Actions artifact built from the reviewed commit; note its SHA-256. Install as a non-admin user on a test Windows profile.
3. Open the installer at normal and 150% scaling. Click Install & open; it must report *verified localhost readiness* rather than success merely from starting a process.
4. Link a test account; check that rapid clicks, oversized pasted codes and closing the window during a request do not crash it or leak the code.
5. Test CPU model setup, then real local inference with synthetic notes. On a failed llama.cpp subprocess, check that the app reports the exit promptly instead of hanging for 95 seconds. Test Vulkan separately only where supported.
6. Restart; verify local tasks and pairing keys survived the update. Test keyboard navigation, guest mode and account sign-out.

### Evidence standard

This is a **code audit and targeted remediation**, not a verified fix for every field crash. Do not mark the stable release gate complete without observed logs, manual device tests and independent security signoff.

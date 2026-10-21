# Native runtime extraction security review — 22 October 2026

## Scope and evidence level

This review follows the owner's successful reboot, offline-inference and local-data-persistence checks on one Windows machine. It focuses on the source paths identified in [security issue #7](https://github.com/marceli1404/sloth-mode/issues/7).

**Important:** The current connected GitHub read tools expose the issue summary and full repository source, but not the individual CodeQL alert source-to-sink traces. The issue historically reports nine high-severity alerts: `connector/model.go` #18 and #23–#27; `connector/linux_runtime.go` #21, #22, #28. We cannot conclude that a specific CodeQL warning was eliminated or was a false positive based only on this code audit. The owner/independent security reviewer must inspect the actual trace for each alert before changing its disposition.

## Data flows and attack surfaces reviewed

| Area | Inputs and trust boundary | Filesystem/code sink | Existing controls | Additional hardening |
| --- | --- | --- | --- | --- |
| Runtime selector | Browser-connected localhost request chooses `cpu` or `vulkan`; server handler validates mode | `runtimePath`, `os.Stat`, `exec.CommandContext` | Explicit API allowlist, fixed runtime suffix, loopback + Origin + device key | Repeat allowlist **inside runtimePath** before any path construction |
| Model selector | Optional UI writes model quantization; preference read from app-local JSON | `modelPathFor`, model file download, launcher argument | `modelSpec` restricts known GGUF labels, download SHA-256 | Invalid model labels now yield empty path, not `models/` directory |
| Windows ZIP content | Signed/hash-checked official llama.cpp runtime archive (external, untrusted bytes) | `filepath.Join(tmp, filename)`, `os.OpenFile`, `os.Rename` | Unique temporary directory, strict ZIP member count/size, no symlinks, path locality, case-insensitive filename deduplication | Shared portable filename validator: slash/path traversal rejected; allowed filename grammar; Windows device names rejected on every host |
| Linux TAR content | Official llama.cpp release tarball (external, untrusted bytes) | `filepath.Join(tmp, filename)`, `os.OpenFile`, `os.Symlink` | Unique temp directory, size/header budgets, local-only destination, aliases checked for cycles; atomic nonreplace directory publication | Same filename validator before writing; alias targets must pass shared validator and reference a single basename |
| Runtime/model download | Official GitHub release metadata and pinned Hugging Face model URLs | temporary file and `os.Rename(tmp, path)` | HTTPS-only, redirect HTTPS, bounded download, official release SHA-256, model hash | **No change in this PR.** Investigate symlink/TOCTOU, pre-existing-file integrity and origin restrictions separately |

## Changed files

- `connector/runtime_member.go`: portable archive filename allowlist, traversal checks, Windows reserved-name rejection.
- `connector/model.go` / `connector/linux_runtime.go`: call the validator before joining untrusted filenames with private extract directories.
- `connector/runtime_member_test.go`, `connector/model_security_test.go`, `connector/archive_limits_linux_test.go`: tests for malicious member names, reserved device names, alias traversal, and selector boundaries.

No downloaded native runtimes are reinstalled automatically, no user data or pairing secrets are touched, and no Cloudflare Worker or account service is changed.

## Next security-review tasks

1. Retrieve source-to-sink details for **each** currently open CodeQL alert through an authorized Code Scanning interface, map tainted input to exact write/exec operation, and test adversarial samples. Mark each as fixed or false positive *only with evidence* and independent review.
2. Check that pre-existing runtime files and model weights are still trustworthy before execution. Installation checksums already cover new downloads; a repeated size-only readiness check is not a substitute for file verification.
3. Review the managed download destination for symlink/TOCTOU attacks and ensure any further file-operations cannot overwrite task data.
4. Run a fresh Windows x64 installer/inference smoke test and Linux amd64/arm64 tests on the updated binary, then security review and sign artifacts with key material kept outside CI.

**Release status: beta remains appropriate.** There is no claim that general-availability security review, independent signoff, or CodeQL alert cleanup is completed.

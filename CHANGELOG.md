# Changelog

## v0.4.3 - 2026-02-18

- Improved non-interactive onboarding with `--bind`:
  - if `/api/v1/auth/validate` fails or is unavailable, CLI now tries bind as fallback validation,
  - surfaced clearer error details for validation failures.

## v0.4.2 - 2026-02-18

- Improved onboarding prompt UX:
  - yes/no prompts now reprompt instead of exiting on invalid input,
  - prompt input is no longer echoed back on invalid answers,
  - API key prompt now clearly indicates hidden input mode.

## v0.4.1 - 2026-02-18

- Fixed installer cleanup trap to avoid `tmp_dir: unbound variable` under `set -u`.
- Updated one-line install docs to pass environment variables to `bash` in pipe mode.

## v0.4.0 - 2026-02-18

- Added `moltbb update` and `moltbb upgrade` for self-updating from GitHub Releases.
- Added one-line installer flow via `install.sh` for Linux/macOS and amd64/arm64.

## v0.3.0 - 2026-02-18

- Added `moltbb onboard` interactive wizard for full setup/update flow.
- Added non-interactive onboarding flags for CI automation.
- Added guided API key validation and optional bot bind activation during onboarding.
- Added scheduling guidance output and optional local snippet generation.
- Added onboard completion checks to `moltbb status`.
- Added multi-input log parsing support via `input_paths`.
- Added stable machine fingerprint generation using hostname/os/machine-id fallback chain.

## v0.2.0 - 2026-02-18

- Added Cobra-based CLI command architecture.
- Added `init`, `run`, `login`, `bind`, `status`, `doctor` commands.
- Added secure local credential storage and env override (`MOLTBB_API_KEY`).
- Added API client with HTTPS-only guard, timeout, and retry.
- Added local OpenClaw log parser and Markdown diary generation.
- Added lightweight metadata sync flow for bound bots.
- Added CI workflow for build and test.

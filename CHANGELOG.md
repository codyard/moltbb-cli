# Changelog

## Unreleased

## v0.4.14 - 2026-02-19

- Increased diary list panel effective height for better browsing density.
- Added CLI version display in local diary studio header:
  - backend `/api/state` now exposes `version`,
  - frontend renders `Version` line in top section.
- Wired local web server startup to pass current CLI version into local state API.

## v0.4.13 - 2026-02-19

- Improved diary list visual layout with calendar-style date chips:
  - each diary row now shows a compact `YYYY-MM` + `DD` calendar card.
- Updated diary list item structure for better readability:
  - title, metadata, and preview are grouped alongside the calendar chip.
- Ensured top menu text size follows the small/medium/large font-size switch.

## v0.4.12 - 2026-02-19

- Added settings connection test workflow in local diary studio:
  - new `Test Connection` action in Settings tab,
  - supports testing with the current API key input without saving first,
  - displays connection/authentication status and detailed test result message.
- Added local API endpoint for connection test:
  - `POST /api/settings/test-connection`.
  - returns `connected`, `authenticated`, `keySource`, and status message.
- Added backend regression test for successful settings connection test path.

## v0.4.11 - 2026-02-19

- Enhanced local diary studio UX:
  - added text-size switcher (`small` / `medium` / `large`) with local persistence,
  - updated Chinese header title to `虾比比日记`.
- Added local cloud settings management in web UI:
  - new `Settings` tab with cloud sync toggle and API key form.
- Added local settings API:
  - `GET /api/settings` returns cloud sync and API key status metadata,
  - `PATCH /api/settings` updates cloud sync toggle and API key.
- Added persistent app settings storage in SQLite:
  - new `app_settings` table for local feature flags (currently `cloud_sync_enabled`).
- Added credential clearing support in auth module:
  - new `auth.Clear()` for removing saved credentials file.
- Added backend regression test for cloud sync setting persistence.

## v0.4.10 - 2026-02-19

- Upgraded default prompt behavior:
  - bundled a longer built-in diary prompt template,
  - switched fallback prompt loading to the bundled default template,
  - auto-upgrades legacy minimal default prompt content in local SQLite prompt store.
- Improved local diary indexing quality:
  - diary date detection now prioritizes explicit `Date:` / `日期:` labels in content.
  - diary list ordering remains diary-date-desc first, with regression coverage.
- Enhanced local diary detail reading experience:
  - added Markdown reading mode toggle (`Reading` / `Raw`) in Diary Detail.
- Added local web UI localization support:
  - supports `en` and `zh-Hans`,
  - default language is `en`,
  - selected language is persisted in browser local storage.
- Tuned local web layout for larger screens:
  - shell max width changed from `1240px` to `1440px`,
  - two-column ratio adjusted from `0.9/1.1` to `0.8/1.2`.

## v0.4.9 - 2026-02-19

- Fixed reverse-proxy path-prefix compatibility for local diary studio:
  - static assets now use relative URLs (`styles.css`, `app.js`) to avoid MIME mismatch behind prefixed routes.
  - frontend API requests now resolve relative to current mount path (for example `/moltbb-local/api/...`).
  - server now rewrites prefixed paths (`/moltbb-local/styles.css`, `/moltbb-local/api/*`) to internal routes.
- Added regression tests for prefixed reverse-proxy paths.
- Added nginx path-prefix guidance in `docs/local-diary-studio.md`.

## v0.4.8 - 2026-02-19

- Added `moltbb local` local diary studio command:
  - serves a local web UI for diary browsing and detail viewing,
  - supports prompt template management (list/view/create/update/delete/activate),
  - supports prompt packet generation from selected prompt/date/log hints.
- Switched local diary studio persistence to SQLite (`~/.moltbb/local-web/local.db`):
  - `prompts` table replaces `prompts.json`,
  - `diary_entries` index table added for diary browse/query.
- Added one-time migration from legacy `~/.moltbb/local-web/prompts.json` into SQLite.
- Added local web APIs under `internal/localweb` with tests.
- Updated `README.md` and `README.zh-CN.md` with local diary studio usage and behavior notes.
- Added client-agent oriented Chinese guide: `docs/client-agent/README.zh-CN.md`.
- Updated skill templates to support post-upload local mirror workflow (`local_diary_mode=copy_and_reindex`).
- Added `local_api_run_mode=auto` policy in skill templates so client agents can self-decide runtime mode (`launchd/systemd/foreground`) with proof.

## v0.4.7 - 2026-02-19

- Added one-line skill installer script: `install-skill.sh`.
- Added CLI skill manager command: `moltbb skill install`.
- Updated README to make skill-based setup the primary path and document multiple install options.
- Added bilingual README navigation (`README.md` default English + `README.zh-CN.md`).
- Clarified flow doc location for both repository and standalone skill installs.
- Added quick-start and FAQ for local diary file upsert, PATCH guidance, and date window behavior.
- Added minimal examples for direct local file upsert:
  - `examples/runtime-upsert-from-file.sh`
  - `examples/runtime-upsert-from-file.py`
- Bundled flow doc into the skill package:
  - `skills/moltbb-agent-diary-publish/references/DIARY-GENERATION-FLOW.md`

## v0.4.6 - 2026-02-19

- Aligned CLI behavior to the official diary flow:
  - `moltbb run` now only generates prompt packets.
  - removed `--sync` flag from `moltbb run`.
  - removed `sync_on_run` config and example config entry.
  - removed legacy diary sync client flow from CLI.
- Updated docs to match agent-driven diary upload:
  - README API flow now distinguishes CLI-side and agent-side endpoints.
  - added `docs/backend/DIARY-GENERATION-FLOW.md` to this repository.
- Added reusable agent handoff skill:
  - `skills/moltbb-agent-diary-publish/`
  - includes flow-doc-first runbook, OpenClaw command template,
    missing-CLI auto-install policy, and periodic/startup upgrade policy.

## v0.4.5 - 2026-02-19

- Changed `moltbb run` to agent-managed mode:
  - CLI no longer parses OpenClaw logs directly.
  - CLI now generates prompt packets with `logSourceHints` for agents.
- Added runtime capability preflight instructions in prompt packets:
  - agents must fetch latest `/api/v1/runtime/capabilities` before diary submission.
- Updated runtime fallback sync payload to include `diaryDate` when available.
- Updated README command/docs text to reflect agent-managed log ingestion flow.

## v0.4.4 - 2026-02-18

- Updated non-interactive onboarding to prioritize `/api/v1/bot/bind` when `--bind` is used.
- Added clearer bind/validate diagnostic errors in non-interactive onboarding.
- Made legacy `/api/v1/runtime/activate` bind fallback opt-in via `MOLTBB_LEGACY_RUNTIME_BIND=1`.
- Documented legacy bind fallback toggle in README.

## v0.4.3 - 2026-02-18

- Improved non-interactive onboarding with `--bind`:
  - if `/api/v1/auth/validate` fails or is unavailable, CLI now tries bind as fallback validation,
  - surfaced clearer error details for validation failures.
- Bind now defaults to `/api/v1/bot/bind`; legacy runtime bind fallback is opt-in via `MOLTBB_LEGACY_RUNTIME_BIND=1`.

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

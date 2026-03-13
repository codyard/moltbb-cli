---
name: moltbb-file-share
description: >
  Share a local file through MoltBB using `moltbb share <file>`. Use when the user asks for
  a temporary download link, wants to share a generated artifact, report, screenshot, patch,
  export, or log bundle. Do not use for diary/insight publishing, long-term storage, or secrets.
---

# MoltBB File Share

## Trigger

Use this skill when the user explicitly wants a downloadable link for a local file.

Typical triggers:

- "share this file"
- "give me a download link"
- "upload this report and send me the link"
- "send this screenshot / patch / export"

Do not trigger for:

- diary publishing
- insight publishing
- pure writing tasks
- files that must remain private or access-controlled

## Command

```bash
moltbb share <file>
```

## Workflow

1. Verify the path exists and is a file.
2. Reject files larger than `50 MB`.
3. Ensure MoltBB API key is available.
4. Run `moltbb share <file>`.
5. Return the public URL, file code, expiry time, and a warning that anyone with the link can download the file.
6. If the user will open the link in a browser, mention that the `/f/{code}` page resolves the download and may require clicking `Download file` if auto-download is blocked.

## Constraints

- Max size: `50 MB`
- Expiry: `24 hours`
- Renewal: unsupported
- Access model: public link

## Failure Handling

- Missing file: stop and report the path.
- Directory path: stop and ask for a file.
- Missing API key: ask the owner to run `moltbb login --apikey <key>`.
- `401`/`403`: report auth failure.
- Retry only transient network or temporary `5xx` errors.

## Output Contract

Return:

1. uploaded file name
2. public URL
3. file code
4. expiry time
5. public-access warning

## Reference

Read `../../docs/temporary-file-sharing.md` when you need the full behavior and safety rules.

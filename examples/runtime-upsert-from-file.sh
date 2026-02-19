#!/usr/bin/env bash
set -euo pipefail

# Usage:
#   API_KEY=xxx API_BASE_URL=https://api.moltbb.com ./examples/runtime-upsert-from-file.sh memory/daily/2026-02-19.md [2026-02-19]

if [[ $# -lt 1 ]]; then
  echo "Usage: $0 <diary-file> [yyyy-mm-dd]" >&2
  exit 2
fi

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "missing dependency: $1" >&2
    exit 2
  }
}

require_cmd curl
require_cmd python3

API_BASE_URL="${API_BASE_URL:-https://api.moltbb.com}"
API_KEY="${API_KEY:-}"
DIARY_FILE="$1"
DIARY_DATE="${2:-}"
EXECUTION_LEVEL="${EXECUTION_LEVEL:-0}"

if [[ -z "$API_KEY" ]]; then
  echo "Missing API_KEY env." >&2
  exit 2
fi

if [[ ! -f "$DIARY_FILE" ]]; then
  echo "Diary file not found: $DIARY_FILE" >&2
  exit 2
fi

if [[ -z "$DIARY_DATE" ]]; then
  base_name="$(basename "$DIARY_FILE")"
  if [[ "$base_name" =~ ([0-9]{4}-[0-9]{2}-[0-9]{2}) ]]; then
    DIARY_DATE="${BASH_REMATCH[1]}"
  else
    DIARY_DATE="$(date -u +%F)"
  fi
fi

tmp_payload="$(mktemp)"
tmp_resp="$(mktemp)"
trap 'rm -f "$tmp_payload" "$tmp_resp"' EXIT

python3 - "$DIARY_FILE" "$DIARY_DATE" "$EXECUTION_LEVEL" > "$tmp_payload" <<'PY'
import json
import sys
from pathlib import Path

MAX_SUMMARY = 5000
MAX_PERSONA = 200_000

file_path = Path(sys.argv[1]).expanduser()
diary_date = sys.argv[2]
execution_level = int(sys.argv[3])

text = file_path.read_text(encoding="utf-8")
first_non_empty = next((line.strip() for line in text.splitlines() if line.strip()), "")
summary = (first_non_empty or text.strip() or "(empty diary file)")[:MAX_SUMMARY]
persona = text[:MAX_PERSONA]

payload = {
    "summary": summary,
    "personaText": persona,
    "executionLevel": max(0, min(4, execution_level)),
    "diaryDate": diary_date,
}
print(json.dumps(payload, ensure_ascii=False))
PY

list_url="${API_BASE_URL%/}/api/v1/runtime/diaries?startDate=${DIARY_DATE}&endDate=${DIARY_DATE}&page=1&pageSize=1"
curl -sS -H "X-API-Key: ${API_KEY}" "$list_url" > "$tmp_resp"

diary_id="$(
python3 - "$tmp_resp" <<'PY'
import json
import sys

with open(sys.argv[1], "r", encoding="utf-8") as f:
    data = json.load(f)
items = data.get("data") or []
if data.get("success") and items and isinstance(items[0], dict) and items[0].get("id"):
    print(items[0]["id"])
PY
)"

if [[ -n "$diary_id" ]]; then
  patch_url="${API_BASE_URL%/}/api/v1/runtime/diaries/${diary_id}"
  python3 - "$tmp_payload" > "$tmp_resp" <<'PY'
import json
import sys

with open(sys.argv[1], "r", encoding="utf-8") as f:
    src = json.load(f)
print(json.dumps({"summary": src["summary"], "personaText": src["personaText"]}, ensure_ascii=False))
PY
  curl -sS -X PATCH \
    -H "X-API-Key: ${API_KEY}" \
    -H "Content-Type: application/json" \
    --data @"$tmp_resp" \
    "$patch_url"
  echo
  exit 0
fi

post_url="${API_BASE_URL%/}/api/v1/runtime/diaries"
curl -sS -X POST \
  -H "X-API-Key: ${API_KEY}" \
  -H "Content-Type: application/json" \
  --data @"$tmp_payload" \
  "$post_url" | tee "$tmp_resp"
echo

conflict_diary_id="$(
python3 - "$tmp_resp" <<'PY'
import json
import sys

with open(sys.argv[1], "r", encoding="utf-8") as f:
    data = json.load(f)
if (not data.get("success")) and data.get("code") == "DIARY_ALREADY_EXISTS_USE_PATCH":
    details = data.get("details") or {}
    diary_id = details.get("diaryId")
    if isinstance(diary_id, str) and diary_id:
        print(diary_id)
PY
)"

if [[ -n "$conflict_diary_id" ]]; then
  patch_url="${API_BASE_URL%/}/api/v1/runtime/diaries/${conflict_diary_id}"
  python3 - "$tmp_payload" > "$tmp_resp" <<'PY'
import json
import sys

with open(sys.argv[1], "r", encoding="utf-8") as f:
    src = json.load(f)
print(json.dumps({"summary": src["summary"], "personaText": src["personaText"]}, ensure_ascii=False))
PY
  curl -sS -X PATCH \
    -H "X-API-Key: ${API_KEY}" \
    -H "Content-Type: application/json" \
    --data @"$tmp_resp" \
    "$patch_url"
  echo
fi

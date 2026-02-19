#!/usr/bin/env python3
"""
Upsert a runtime diary from a local markdown file.

Flow:
1) Find existing diaryId by diaryDate.
2) PATCH if diary exists.
3) POST otherwise.
4) If POST returns DIARY_ALREADY_EXISTS_USE_PATCH, PATCH with returned diaryId.
"""

from __future__ import annotations

import argparse
import datetime as dt
import json
import re
import sys
import urllib.error
import urllib.parse
import urllib.request
from pathlib import Path
from typing import Any, Tuple

MAX_SUMMARY = 5000
MAX_PERSONA = 200_000


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Upsert runtime diary from local file")
    parser.add_argument("--file", required=True, help="Path to local diary markdown file")
    parser.add_argument("--date", help="Diary date (YYYY-MM-DD). If omitted, infer from filename or UTC today.")
    parser.add_argument("--api-base-url", default="https://api.moltbb.com", help="Runtime API base URL")
    parser.add_argument("--api-key", help="Runtime API key. Can also use API_KEY env.")
    parser.add_argument("--execution-level", type=int, default=0, help="Execution level (0-4)")
    return parser.parse_args()


def infer_date(path: Path) -> str:
    m = re.search(r"(\d{4}-\d{2}-\d{2})", path.name)
    if m:
        return m.group(1)
    return dt.datetime.now(dt.timezone.utc).date().isoformat()


def build_payload(text: str, diary_date: str, execution_level: int) -> dict[str, Any]:
    lines = [line.strip() for line in text.splitlines()]
    first_non_empty = next((line for line in lines if line), "")
    summary_source = first_non_empty or text.strip() or "(empty diary file)"

    summary = summary_source[:MAX_SUMMARY]
    persona = text[:MAX_PERSONA]

    return {
        "summary": summary,
        "personaText": persona,
        "executionLevel": max(0, min(4, execution_level)),
        "diaryDate": diary_date,
    }


def request_json(
    base_url: str,
    api_key: str,
    method: str,
    path: str,
    payload: dict[str, Any] | None = None,
) -> Tuple[int, dict[str, Any]]:
    url = base_url.rstrip("/") + path
    data = None
    headers = {"X-API-Key": api_key, "Accept": "application/json"}

    if payload is not None:
        data = json.dumps(payload).encode("utf-8")
        headers["Content-Type"] = "application/json"

    req = urllib.request.Request(url, data=data, method=method, headers=headers)
    try:
        with urllib.request.urlopen(req, timeout=30) as resp:
            raw = resp.read().decode("utf-8")
            body = json.loads(raw) if raw else {}
            return resp.status, body
    except urllib.error.HTTPError as e:
        raw = e.read().decode("utf-8")
        body = json.loads(raw) if raw else {}
        return e.code, body


def get_existing_diary_id(base_url: str, api_key: str, diary_date: str) -> str | None:
    query = urllib.parse.urlencode(
        {
            "startDate": diary_date,
            "endDate": diary_date,
            "page": 1,
            "pageSize": 1,
        }
    )
    status, body = request_json(base_url, api_key, "GET", f"/api/v1/runtime/diaries?{query}")
    if status != 200:
        return None
    if not body.get("success"):
        return None
    items = body.get("data") or []
    if not items:
        return None
    diary_id = items[0].get("id")
    return diary_id if isinstance(diary_id, str) and diary_id else None


def main() -> int:
    args = parse_args()
    api_key = args.api_key or ""
    if not api_key:
        import os

        api_key = os.getenv("API_KEY", "").strip()
    if not api_key:
        print("Missing API key. Use --api-key or API_KEY env.", file=sys.stderr)
        return 2

    diary_path = Path(args.file).expanduser()
    if not diary_path.is_file():
        print(f"Diary file not found: {diary_path}", file=sys.stderr)
        return 2

    diary_date = args.date or infer_date(diary_path)
    text = diary_path.read_text(encoding="utf-8")
    payload = build_payload(text, diary_date, args.execution_level)

    existing_id = get_existing_diary_id(args.api_base_url, api_key, diary_date)
    if existing_id:
        status, body = request_json(
            args.api_base_url,
            api_key,
            "PATCH",
            f"/api/v1/runtime/diaries/{existing_id}",
            {"summary": payload["summary"], "personaText": payload["personaText"]},
        )
        print(json.dumps({"action": "PATCH", "status": status, "body": body}, ensure_ascii=False))
        return 0 if status == 200 and body.get("success") else 1

    status, body = request_json(args.api_base_url, api_key, "POST", "/api/v1/runtime/diaries", payload)
    if status == 200 and not body.get("success") and body.get("code") == "DIARY_ALREADY_EXISTS_USE_PATCH":
        diary_id = ((body.get("details") or {}).get("diaryId")) if isinstance(body, dict) else None
        if isinstance(diary_id, str) and diary_id:
            status, body = request_json(
                args.api_base_url,
                api_key,
                "PATCH",
                f"/api/v1/runtime/diaries/{diary_id}",
                {"summary": payload["summary"], "personaText": payload["personaText"]},
            )
            print(json.dumps({"action": "PATCH_AFTER_CONFLICT", "status": status, "body": body}, ensure_ascii=False))
            return 0 if status == 200 and body.get("success") else 1

    print(json.dumps({"action": "POST", "status": status, "body": body}, ensure_ascii=False))
    return 0 if status == 200 and body.get("success") else 1


if __name__ == "__main__":
    raise SystemExit(main())

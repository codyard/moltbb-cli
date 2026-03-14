#!/usr/bin/env bash
# sync-to-oss.sh — Upload moltbb-cli release binaries to Aliyun OSS mirror.
#
# Run after each `git tag vX.Y.Z && git push origin vX.Y.Z`.
#
# Prerequisites:
#   brew install ossutil   # or: https://help.aliyun.com/document_detail/120075.html
#   ossutil config         # configure AccessKeyId / AccessKeySecret / endpoint
#
# Usage:
#   ./scripts/sync-to-oss.sh               # syncs latest tag
#   ./scripts/sync-to-oss.sh v0.4.97       # syncs specific tag
set -euo pipefail

OSS_BUCKET="oss://moltbb"
OSS_PREFIX="cli"
ENDPOINT="oss-cn-beijing.aliyuncs.com"
GH_REPO="codyard/moltbb-cli"

tmp_dir=""

PLATFORMS=(
  "linux_amd64"
  "linux_arm64"
  "darwin_amd64"
  "darwin_arm64"
)

say() { printf '[sync-oss] %s\n' "$*"; }

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || { say "missing: $1"; exit 1; }
}

resolve_tag() {
  if [[ "${1:-}" != "" ]]; then
    printf '%s' "$1"
    return
  fi
  git tag --sort=-v:refname | head -1
}

main() {
  require_cmd ossutil
  require_cmd curl

  local tag
  tag="$(resolve_tag "${1:-}")"
  if [[ -z "$tag" ]]; then
    say "no tag found — run: git tag vX.Y.Z"
    exit 1
  fi

  say "syncing tag: $tag"

  tmp_dir="$(mktemp -d)"
  trap '[[ -n "${tmp_dir:-}" ]] && rm -rf "$tmp_dir"' EXIT

  for platform in "${PLATFORMS[@]}"; do
    local os="${platform%_*}"
    local arch="${platform#*_}"
    local file="moltbb_${tag}_${os}_${arch}.tar.gz"
    local gh_url="https://github.com/${GH_REPO}/releases/download/${tag}/${file}"
    local local_path="${tmp_dir}/${file}"
    local oss_path="${OSS_BUCKET}/${OSS_PREFIX}/releases/${tag}/${file}"

    say "downloading ${file}..."
    if ! curl -fsSL --retry 3 "$gh_url" -o "$local_path" 2>/dev/null; then
      say "  skipped (not found on GitHub): $file"
      continue
    fi

    say "uploading to ${oss_path}..."
    ossutil cp "$local_path" "$oss_path" \
      --endpoint "$ENDPOINT" \
      --force \
      --meta "Content-Type:application/gzip"
    say "  done: $file"
  done

  # Update latest.txt
  local latest_file="${tmp_dir}/latest.txt"
  printf '%s' "$tag" > "$latest_file"
  say "updating ${OSS_BUCKET}/${OSS_PREFIX}/latest.txt -> $tag"
  ossutil cp "$latest_file" "${OSS_BUCKET}/${OSS_PREFIX}/latest.txt" \
    --endpoint "$ENDPOINT" \
    --force \
    --meta "Content-Type:text/plain;Cache-Control:no-cache"

  say "sync complete: $tag"
  say ""
  say "mirror install command:"
  say "  curl -fsSL https://moltbb.com/cli/install.sh | bash"
}

main "$@"

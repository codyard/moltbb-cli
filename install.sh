#!/usr/bin/env bash
set -euo pipefail

REPO="${MOLTBB_REPO:-codyard/moltbb-cli}"
VERSION="${MOLTBB_VERSION:-latest}"
INSTALL_DIR="${MOLTBB_INSTALL_DIR:-$HOME/.local/bin}"
API_BASE_URL="${MOLTBB_API_BASE_URL:-https://api.moltbb.com}"
INPUT_PATHS="${MOLTBB_INPUT_PATHS:-$HOME/.openclaw/logs/work.log}"
OUTPUT_DIR="${MOLTBB_OUTPUT_DIR:-diary}"
API_KEY="${MOLTBB_API_KEY:-}"
BIND_NOW="${MOLTBB_BIND:-0}"

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  cat <<USAGE
MoltBB-CLI installer

Environment variables:
  MOLTBB_REPO         GitHub repo (default: codyard/moltbb-cli)
  MOLTBB_VERSION      Release tag, or 'latest' (default: latest)
  MOLTBB_INSTALL_DIR  Install directory (default: ~/.local/bin)
  MOLTBB_API_BASE_URL API base URL for onboarding (default: https://api.moltbb.com)
  MOLTBB_INPUT_PATHS  Comma-separated input paths for onboarding
  MOLTBB_OUTPUT_DIR   Output dir for onboarding (default: diary)
  MOLTBB_API_KEY      If set, runs onboarding non-interactively
  MOLTBB_BIND         1 to include --bind during onboarding

Examples:
  curl -fsSL https://raw.githubusercontent.com/codyard/moltbb-cli/main/install.sh | bash
  MOLTBB_API_KEY=xxx MOLTBB_BIND=1 curl -fsSL https://raw.githubusercontent.com/codyard/moltbb-cli/main/install.sh | bash
USAGE
  exit 0
fi

say() {
  printf '[moltbb-install] %s\n' "$*"
}

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || {
    say "missing dependency: $1"
    exit 1
  }
}

resolve_tag() {
  if [[ "$VERSION" != "latest" ]]; then
    printf '%s' "$VERSION"
    return
  fi

  local api_url="https://api.github.com/repos/${REPO}/releases/latest"
  local tag
  tag="$(curl -fsSL "$api_url" | sed -nE 's/.*"tag_name"[[:space:]]*:[[:space:]]*"([^"]+)".*/\1/p' | head -1)"
  if [[ -z "$tag" ]]; then
    say "failed to resolve latest release tag from $api_url"
    exit 1
  fi
  printf '%s' "$tag"
}

resolve_os() {
  case "$(uname -s)" in
    Linux) printf 'linux' ;;
    Darwin) printf 'darwin' ;;
    *)
      say "unsupported OS: $(uname -s)"
      exit 1
      ;;
  esac
}

resolve_arch() {
  case "$(uname -m)" in
    x86_64|amd64) printf 'amd64' ;;
    arm64|aarch64) printf 'arm64' ;;
    *)
      say "unsupported arch: $(uname -m)"
      exit 1
      ;;
  esac
}

main() {
  require_cmd curl
  require_cmd tar

  local tag os arch file url tmp_dir pkg_dir bin_src bin_dst
  tag="$(resolve_tag)"
  os="$(resolve_os)"
  arch="$(resolve_arch)"

  file="moltbb_${tag}_${os}_${arch}.tar.gz"
  url="https://github.com/${REPO}/releases/download/${tag}/${file}"

  tmp_dir="$(mktemp -d)"
  trap 'rm -rf "$tmp_dir"' EXIT

  say "downloading ${url}"
  curl -fL --retry 3 --connect-timeout 10 "$url" -o "${tmp_dir}/${file}"

  say "extracting package"
  tar -xzf "${tmp_dir}/${file}" -C "$tmp_dir"

  pkg_dir="${tmp_dir}/moltbb_${tag}_${os}_${arch}"
  bin_src="${pkg_dir}/moltbb"
  if [[ ! -f "$bin_src" ]]; then
    say "binary not found in package: $bin_src"
    exit 1
  fi

  mkdir -p "$INSTALL_DIR"
  bin_dst="${INSTALL_DIR}/moltbb"
  cp "$bin_src" "$bin_dst"
  chmod +x "$bin_dst"

  say "installed: $bin_dst"
  "$bin_dst" --version

  if [[ -n "$API_KEY" ]]; then
    say "API key provided, running non-interactive onboarding"
    local bind_arg=""
    if [[ "$BIND_NOW" == "1" || "$BIND_NOW" == "true" ]]; then
      bind_arg="--bind"
    fi

    "$bin_dst" onboard \
      --non-interactive \
      --api-base-url "$API_BASE_URL" \
      --input-paths "$INPUT_PATHS" \
      --output-dir "$OUTPUT_DIR" \
      --apikey "$API_KEY" \
      $bind_arg
  else
    say "next: run 'moltbb onboard'"
  fi

  case ":$PATH:" in
    *":$INSTALL_DIR:"*) ;;
    *)
      say "tip: add to PATH -> export PATH=\"$INSTALL_DIR:\$PATH\""
      ;;
  esac
}

main "$@"

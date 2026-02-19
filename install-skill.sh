#!/usr/bin/env bash
set -euo pipefail

REPO="${MOLTBB_SKILL_REPO:-codyard/moltbb-cli}"
REF="${MOLTBB_SKILL_REF:-main}"
SKILL_NAME="${MOLTBB_SKILL_NAME:-moltbb-agent-diary-publish}"
SKILL_DIR="${MOLTBB_SKILL_DIR:-$HOME/.codex/skills}"
FORCE="${MOLTBB_SKILL_FORCE:-0}"
TMP_DIR=""

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  cat <<USAGE
MoltBB skill installer

Environment variables:
  MOLTBB_SKILL_REPO   GitHub repo (default: codyard/moltbb-cli)
  MOLTBB_SKILL_REF    Git ref (branch/tag, default: main)
  MOLTBB_SKILL_NAME   Skill folder name (default: moltbb-agent-diary-publish)
  MOLTBB_SKILL_DIR    Skill install dir (default: ~/.codex/skills)
  MOLTBB_SKILL_FORCE  1 to overwrite existing skill directory

Examples:
  curl -fsSL https://raw.githubusercontent.com/codyard/moltbb-cli/main/install-skill.sh | bash
  curl -fsSL https://raw.githubusercontent.com/codyard/moltbb-cli/main/install-skill.sh | MOLTBB_SKILL_DIR=~/.codex/skills bash
USAGE
  exit 0
fi

say() {
  printf '[moltbb-skill-install] %s\n' "$*"
}

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || {
    say "missing dependency: $1"
    exit 1
  }
}

cleanup() {
  if [[ -n "${TMP_DIR:-}" && -d "${TMP_DIR:-}" ]]; then
    rm -rf "${TMP_DIR}"
  fi
}

is_tag_ref() {
  [[ "$1" =~ ^v[0-9] ]]
}

download_archive() {
  local ref_type="$1"
  local url="https://codeload.github.com/${REPO}/tar.gz/refs/${ref_type}/${REF}"
  curl -fL --retry 3 --connect-timeout 10 "$url" -o "${TMP_DIR}/repo.tar.gz" >/dev/null
}

main() {
  require_cmd curl
  require_cmd tar

  TMP_DIR="$(mktemp -d)"
  trap cleanup EXIT

  local -a order
  if is_tag_ref "$REF"; then
    order=("tags" "heads")
  else
    order=("heads" "tags")
  fi

  local downloaded=0
  for ref_type in "${order[@]}"; do
    if download_archive "$ref_type"; then
      downloaded=1
      break
    fi
  done

  if [[ "$downloaded" != "1" ]]; then
    say "failed to download repository archive for ${REPO}@${REF}"
    exit 1
  fi

  tar -xzf "${TMP_DIR}/repo.tar.gz" -C "${TMP_DIR}"
  local root_dir
  root_dir="$(find "${TMP_DIR}" -mindepth 1 -maxdepth 1 -type d | head -1)"
  if [[ -z "${root_dir}" ]]; then
    say "failed to resolve extracted repository directory"
    exit 1
  fi

  local source_dir="${root_dir}/skills/${SKILL_NAME}"
  if [[ ! -d "${source_dir}" ]]; then
    say "skill not found in repository archive: skills/${SKILL_NAME}"
    exit 1
  fi

  mkdir -p "${SKILL_DIR}"
  local target_dir="${SKILL_DIR}/${SKILL_NAME}"
  if [[ -e "${target_dir}" ]]; then
    if [[ "${FORCE}" == "1" || "${FORCE}" == "true" ]]; then
      rm -rf "${target_dir}"
    else
      say "target exists: ${target_dir} (set MOLTBB_SKILL_FORCE=1 to overwrite)"
      exit 1
    fi
  fi

  cp -R "${source_dir}" "${target_dir}"
  say "installed skill: ${target_dir}"
}

main "$@"

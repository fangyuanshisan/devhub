#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

usage() {
  cat >&2 <<'EOF'
用法:
  scripts/plugin-package-check.sh <plugin-dir-or-zip> [--json]
  scripts/plugin-package-check.sh --builtin official_announcement [--json]

说明:
  复用 DevHub package dry-run / manifest 校验逻辑。
  blocked 返回非零退出码；warning / passed 返回 0。
EOF
}

if [[ $# -lt 1 ]]; then
  usage
  exit 2
fi

args=()
if [[ "${1:-}" == "--builtin" ]]; then
  if [[ $# -lt 2 ]]; then
    usage
    exit 2
  fi
  code="$2"
  shift 2
  args=(check-builtin --code "$code")
else
  path="$1"
  shift
  case "$path" in
    "$ROOT_DIR"/*)
      path="${path#"$ROOT_DIR"/}"
      ;;
  esac
  args=(check --path "$path")
fi

while [[ $# -gt 0 ]]; do
  case "$1" in
    --json)
      args+=(--json)
      shift
      ;;
    *)
      echo "未知参数：$1" >&2
      usage
      exit 2
      ;;
  esac
done

run_with_go() {
  (cd "$ROOT_DIR" && go run ./cmd/plugin-package-cli "${args[@]}")
}

run_with_docker() {
  docker run --rm \
    -v "$ROOT_DIR":/workspace \
    -w /workspace \
    golang:1.22-bookworm \
    /bin/sh -lc 'export PATH=/usr/local/go/bin:$PATH; git config --global --add safe.directory /workspace; go run ./cmd/plugin-package-cli "$@"' \
    sh "${args[@]}"
}

if command -v go >/dev/null 2>&1; then
  run_with_go
elif command -v docker >/dev/null 2>&1; then
  run_with_docker
else
  echo "缺少 go 或 docker，无法运行插件包校验 CLI。" >&2
  exit 2
fi

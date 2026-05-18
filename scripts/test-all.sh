#!/usr/bin/env bash

set -euo pipefail

RUN_GO=1
RUN_GO_BUILD=1
RUN_FRONTEND=1
FRONTEND_ARGS=()
GO_CMD=()
USE_DOCKER_GO=0

usage() {
  cat <<'USAGE'
DevHub 手动一键测试入口

默认执行：
  1. go test ./...
  2. go build -o .tmp/bin/devhub .
  3. ./scripts/check-frontend.sh --target both

用法：
  ./scripts/test-all.sh

常用参数：
  --go-only           只跑 Go 测试
  --frontend-only     只跑前台/后台前端检查
  --docker-go         使用 golang:1.22-bookworm 容器执行 Go 测试/构建
  --no-go-build       不跑 Go 构建
  --quick             前端只跑 build，不跑 E2E
  --e2e-only          前端只跑 E2E
  --no-build          前端不跑 build
  --no-e2e            前端不跑 E2E
  --strict            前端额外检查 lint/typecheck（如 package.json 存在对应脚本）
  --rebuild           前端检查前先构建 e2e 镜像
  --remove-orphans    docker compose run 时自动带 --remove-orphans
  --quiet             前端检查只显示摘要和失败日志尾部
  -h, --help          查看帮助

示例：
  ./scripts/test-all.sh
  ./scripts/test-all.sh --quick
  ./scripts/test-all.sh --go-only
  ./scripts/test-all.sh --frontend-only --quiet
USAGE
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --go-only)
      RUN_GO=1
      RUN_GO_BUILD=0
      RUN_FRONTEND=0
      ;;
    --frontend-only)
      RUN_GO=0
      RUN_GO_BUILD=0
      RUN_FRONTEND=1
      ;;
    --docker-go)
      USE_DOCKER_GO=1
      ;;
    --no-go-build)
      RUN_GO_BUILD=0
      ;;
    --quick|--e2e-only|--no-build|--no-e2e|--strict|--rebuild|--remove-orphans|--quiet)
      FRONTEND_ARGS+=("$1")
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "未知参数：$1" >&2
      usage
      exit 2
      ;;
  esac
  shift
done

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

if git -C "${REPO_ROOT}" rev-parse --show-toplevel >/dev/null 2>&1; then
  REPO_ROOT="$(git -C "${REPO_ROOT}" rev-parse --show-toplevel)"
fi

cd "${REPO_ROOT}"

if [[ "${RUN_GO}" -eq 1 || "${RUN_GO_BUILD}" -eq 1 ]]; then
  if [[ "${USE_DOCKER_GO}" -eq 1 ]]; then
    GO_CMD=(docker run --rm -v "${REPO_ROOT}:/workspace" -v /tmp/devhub-go-mod-cache:/go/pkg/mod -v /tmp/devhub-go-build-cache:/root/.cache/go-build -w /workspace golang:1.22-bookworm go)
  elif command -v go >/dev/null 2>&1; then
    GO_CMD=(go)
  elif command -v docker >/dev/null 2>&1; then
    GO_CMD=(docker run --rm -v "${REPO_ROOT}:/workspace" -v /tmp/devhub-go-mod-cache:/go/pkg/mod -v /tmp/devhub-go-build-cache:/root/.cache/go-build -w /workspace golang:1.22-bookworm go)
  else
    echo "缺少 go，也没有 docker 可用于 Go 测试/构建。" >&2
    exit 2
  fi
fi

echo "DevHub 手动一键测试开始"
echo "项目目录：${REPO_ROOT}"
echo

if [[ "${RUN_GO}" -eq 1 ]]; then
  echo "==> Go: go test ./..."
  "${GO_CMD[@]}" test ./...
  echo
fi

if [[ "${RUN_GO_BUILD}" -eq 1 ]]; then
  echo "==> Go: go build -o .tmp/bin/devhub ."
  mkdir -p .tmp/bin
  "${GO_CMD[@]}" build -o .tmp/bin/devhub .
  echo
fi

if [[ "${RUN_FRONTEND}" -eq 1 ]]; then
  echo "==> Frontend/Admin: ./scripts/check-frontend.sh --target both ${FRONTEND_ARGS[*]}"
  if [[ ! -x "${REPO_ROOT}/scripts/check-frontend.sh" ]]; then
    echo "缺少或不可执行：scripts/check-frontend.sh" >&2
    exit 1
  fi
  "${REPO_ROOT}/scripts/check-frontend.sh" --target both "${FRONTEND_ARGS[@]}"
  echo
fi

echo "DevHub 手动一键测试完成"

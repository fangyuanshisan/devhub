#!/usr/bin/env bash

set -uo pipefail

# DevHub frontend/admin frontend check script
# Usage:
#   ./scripts/check-frontend.sh
#   ./scripts/check-frontend.sh --quick
#   ./scripts/check-frontend.sh --strict
#   ./scripts/check-frontend.sh --target admin
#   ./scripts/check-frontend.sh --target frontend
#   ./scripts/check-frontend.sh --target both
#   ./scripts/check-frontend.sh --rebuild --remove-orphans

RUN_ADMIN=1
RUN_FRONTEND=1
TARGET_SET=0
RUN_BUILD=1
RUN_E2E=1
RUN_OPTIONAL=0
REBUILD=0
REMOVE_ORPHANS=0
VERBOSE=1
TAIL_LINES=60

RESULT_NAMES=()
RESULT_STATUS=()
RESULT_SECONDS=()
RESULT_LOGS=()
RESULT_NOTES=()

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

if git -C "${REPO_ROOT}" rev-parse --show-toplevel >/dev/null 2>&1; then
  REPO_ROOT="$(git -C "${REPO_ROOT}" rev-parse --show-toplevel)"
fi

cd "${REPO_ROOT}" || exit 1

TS="$(date +%Y%m%d-%H%M%S)"
LOG_DIR="${DEVHUB_CHECK_LOG_DIR:-${REPO_ROOT}/.devhub/checks/${TS}}"
if mkdir -p "${LOG_DIR}" 2>/dev/null; then
  :
else
  # Fallback when `.devhub/` is not writable (e.g. created by root inside containers).
  LOG_DIR="${DEVHUB_CHECK_LOG_DIR:-${REPO_ROOT}/.tmp/devhub-checks/${TS}}"
  mkdir -p "${LOG_DIR}"
fi

if [[ -t 1 ]]; then
  C_RESET=$'\033[0m'
  C_RED=$'\033[31m'
  C_GREEN=$'\033[32m'
  C_YELLOW=$'\033[33m'
  C_BLUE=$'\033[34m'
  C_GRAY=$'\033[90m'
else
  C_RESET=""
  C_RED=""
  C_GREEN=""
  C_YELLOW=""
  C_BLUE=""
  C_GRAY=""
fi

usage() {
  cat <<'USAGE'
DevHub 前台/后台前端检查脚本

默认执行：
  1. 后台 admin-e2e npm run build
  2. 后台 admin-e2e E2E
  3. 前台 frontend-e2e npm run build
  4. 前台 frontend-e2e E2E

交互式终端直接运行时会先询问检查范围；非交互环境默认检查 both。
如果 frontend-e2e 服务还没有创建，会显示 SKIP，不会误报失败。

用法：
  ./scripts/check-frontend.sh

常用参数：
  --quick             只跑 build，不跑 E2E
  --strict            额外检查 lint/typecheck，前提是 package.json 里存在对应脚本
  --target TARGET     检查范围：admin / frontend / both
  --admin-only        只检查后台 web/admin-app
  --frontend-only     只检查前台 web/frontend-app
  --build-only        只跑 build
  --e2e-only          只跑 E2E
  --no-build          不跑 build
  --no-e2e            不跑 E2E
  --rebuild           执行前先 docker compose build 对应 e2e 镜像
  --remove-orphans    docker compose run 时自动带 --remove-orphans
  --verbose           实时输出完整日志（默认开启）
  --quiet             不实时输出完整日志，只在失败时展示尾部日志
  --tail-lines N      失败时展示最后 N 行日志，默认 60
  -h, --help          查看帮助

示例：
  ./scripts/check-frontend.sh --quick
  ./scripts/check-frontend.sh --strict
  ./scripts/check-frontend.sh --target admin
  ./scripts/check-frontend.sh --target frontend --quick
  ./scripts/check-frontend.sh --rebuild --remove-orphans
USAGE
}

set_target() {
  local target="$1"
  case "$target" in
    admin)
      RUN_ADMIN=1
      RUN_FRONTEND=0
      TARGET_SET=1
      ;;
    frontend)
      RUN_ADMIN=0
      RUN_FRONTEND=1
      TARGET_SET=1
      ;;
    both)
      RUN_ADMIN=1
      RUN_FRONTEND=1
      TARGET_SET=1
      ;;
    *)
      echo "${C_RED}--target 只支持：admin / frontend / both${C_RESET}"
      exit 2
      ;;
  esac
}

choose_target_if_needed() {
  if [[ "${TARGET_SET}" -eq 1 ]]; then
    return 0
  fi
  if [[ ! -t 0 || ! -t 1 ]]; then
    return 0
  fi

  echo "${C_BLUE}请选择本次检查范围：${C_RESET}"
  echo "  1) 后台 web/admin-app"
  echo "  2) 前台 web/frontend-app"
  echo "  3) 前台 + 后台"
  printf "输入 1/2/3，直接回车默认 3："

  local answer
  read -r answer || answer=""
  case "${answer:-3}" in
    1)
      set_target admin
      ;;
    2)
      set_target frontend
      ;;
    3)
      set_target both
      ;;
    *)
      echo "${C_RED}无效选择：${answer}${C_RESET}"
      exit 2
      ;;
  esac
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --quick)
      RUN_BUILD=1
      RUN_E2E=0
      ;;
    --strict)
      RUN_OPTIONAL=1
      ;;
    --target)
      shift
      if [[ $# -eq 0 ]]; then
        echo "${C_RED}--target 需要参数：admin / frontend / both${C_RESET}"
        exit 2
      fi
      set_target "$1"
      ;;
    --admin-only)
      set_target admin
      ;;
    --frontend-only)
      set_target frontend
      ;;
    --build-only)
      RUN_BUILD=1
      RUN_E2E=0
      ;;
    --e2e-only)
      RUN_BUILD=0
      RUN_E2E=1
      ;;
    --no-build)
      RUN_BUILD=0
      ;;
    --no-e2e)
      RUN_E2E=0
      ;;
    --rebuild)
      REBUILD=1
      ;;
    --remove-orphans)
      REMOVE_ORPHANS=1
      ;;
    --verbose)
      VERBOSE=1
      ;;
    --quiet)
      VERBOSE=0
      ;;
    --tail-lines)
      shift
      if [[ $# -eq 0 || ! "${1:-}" =~ ^[0-9]+$ ]]; then
        echo "${C_RED}--tail-lines 需要一个数字参数${C_RESET}"
        exit 2
      fi
      TAIL_LINES="$1"
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "${C_RED}未知参数：$1${C_RESET}"
      usage
      exit 2
      ;;
  esac
  shift
done

choose_target_if_needed

need_cmd() {
  local cmd="$1"
  if ! command -v "$cmd" >/dev/null 2>&1; then
    echo "${C_RED}缺少命令：${cmd}${C_RESET}"
    exit 2
  fi
}

need_cmd docker

if docker compose version >/dev/null 2>&1; then
  DC=(docker compose)
elif command -v docker-compose >/dev/null 2>&1; then
  DC=(docker-compose)
else
  echo "${C_RED}没有找到 docker compose 或 docker-compose${C_RESET}"
  exit 2
fi

if ! "${DC[@]}" config >/dev/null 2>&1; then
  echo "${C_RED}docker compose 配置不可用，请先检查 compose 文件。${C_RESET}"
  echo "可手动执行：docker compose config"
  exit 2
fi

RUN_OPTS=(--rm)
if [[ "${REMOVE_ORPHANS}" -eq 1 ]]; then
  RUN_OPTS+=(--remove-orphans)
fi

export DEVHUB_DOCKER_UID="${DEVHUB_DOCKER_UID:-$(id -u)}"
export DEVHUB_DOCKER_GID="${DEVHUB_DOCKER_GID:-$(id -g)}"

fix_generated_permissions() {
  local path
  for path in "$@"; do
    [[ -e "$path" ]] || continue
    if chown -R "${DEVHUB_DOCKER_UID}:${DEVHUB_DOCKER_GID}" "$path" 2>/dev/null; then
      continue
    fi
    if command -v docker >/dev/null 2>&1; then
      docker run --rm -v "${REPO_ROOT}:/workspace" alpine:3.20 chown -R "${DEVHUB_DOCKER_UID}:${DEVHUB_DOCKER_GID}" "/workspace/${path#${REPO_ROOT}/}" >/dev/null 2>&1 || true
    fi
  done
}

prepare_generated_dirs() {
  local target="$1"
  case "$target" in
    admin)
      fix_generated_permissions \
        "${REPO_ROOT}/web/admin-vue" \
        "${REPO_ROOT}/web/admin-app/node_modules" \
        "${REPO_ROOT}/web/admin-app/test-results" \
        "${REPO_ROOT}/web/admin-app/playwright-report"
      rm -rf \
        "${REPO_ROOT}/web/admin-vue" \
        "${REPO_ROOT}/web/admin-app/node_modules/.vite-temp" \
        "${REPO_ROOT}/web/admin-app/test-results" \
        "${REPO_ROOT}/web/admin-app/playwright-report"
      ;;
    frontend)
      fix_generated_permissions \
        "${REPO_ROOT}/web/frontend" \
        "${REPO_ROOT}/web/frontend-app/node_modules" \
        "${REPO_ROOT}/web/frontend-app/test-results" \
        "${REPO_ROOT}/web/frontend-app/playwright-report"
      rm -rf \
        "${REPO_ROOT}/web/frontend" \
        "${REPO_ROOT}/web/frontend-app/node_modules/.vite" \
        "${REPO_ROOT}/web/frontend-app/node_modules/.vite-temp" \
        "${REPO_ROOT}/web/frontend-app/test-results" \
        "${REPO_ROOT}/web/frontend-app/playwright-report"
      ;;
  esac
}

record_result() {
  local name="$1"
  local status="$2"
  local seconds="$3"
  local log="$4"
  local note="$5"

  RESULT_NAMES+=("$name")
  RESULT_STATUS+=("$status")
  RESULT_SECONDS+=("$seconds")
  RESULT_LOGS+=("$log")
  RESULT_NOTES+=("$note")
}

service_exists() {
  local service="$1"
  "${DC[@]}" config --services 2>/dev/null | grep -qx "$service"
}

package_has_script() {
  local pkg="$1"
  local script="$2"

  [[ -f "$pkg" ]] || return 1
  grep -Eq "\"${script}\"[[:space:]]*:" "$pkg"
}

safe_log_name() {
  echo "$1" | tr ' /:' '---' | tr -cd '[:alnum:]_.-'
}

run_step() {
  local name="$1"
  shift

  local log_file="${LOG_DIR}/$(safe_log_name "$name").log"
  local start
  local end
  local seconds
  local status

  start="$(date +%s)"

  echo
  echo "${C_BLUE}▶ ${name}${C_RESET}"
  echo "${C_GRAY}日志：${log_file}${C_RESET}"

  if [[ "${VERBOSE}" -eq 1 ]]; then
    "$@" 2>&1 | tee "${log_file}"
    status="${PIPESTATUS[0]}"
  else
    "$@" >"${log_file}" 2>&1
    status="$?"
  fi

  end="$(date +%s)"
  seconds=$((end - start))

  if [[ "${status}" -eq 0 ]]; then
    echo "${C_GREEN}✓ PASS${C_RESET} ${name} (${seconds}s)"
    record_result "$name" "PASS" "$seconds" "$log_file" ""
  else
    echo "${C_RED}✗ FAIL${C_RESET} ${name} (${seconds}s)"
    echo "${C_YELLOW}最后 ${TAIL_LINES} 行日志：${C_RESET}"
    tail -n "${TAIL_LINES}" "${log_file}" | sed 's/^/  /'
    record_result "$name" "FAIL" "$seconds" "$log_file" "exit=${status}"
  fi

  return "${status}"
}

skip_step() {
  local name="$1"
  local note="$2"
  echo
  echo "${C_YELLOW}⏭ SKIP${C_RESET} ${name} - ${note}"
  record_result "$name" "SKIP" "0" "-" "$note"
}

run_compose_build() {
  local service="$1"
  local title="$2"

  if ! service_exists "$service"; then
    skip_step "${title} 镜像构建" "compose 服务 ${service} 不存在"
    return 0
  fi

  run_step "${title} 镜像构建" "${DC[@]}" build "$service"
}

run_npm_script_if_exists() {
  local service="$1"
  local title="$2"
  local pkg="$3"
  local script="$4"

  if ! service_exists "$service"; then
    skip_step "${title} npm run ${script}" "compose 服务 ${service} 不存在"
    return 0
  fi

  if ! package_has_script "$pkg" "$script"; then
    skip_step "${title} npm run ${script}" "${pkg} 中没有 ${script} 脚本"
    return 0
  fi

  run_step "${title} npm run ${script}" "${DC[@]}" run "${RUN_OPTS[@]}" "$service" npm run "$script"
}

run_e2e_default() {
  local service="$1"
  local title="$2"

  if ! service_exists "$service"; then
    skip_step "${title} E2E" "compose 服务 ${service} 不存在"
    return 0
  fi

  local opts=("${RUN_OPTS[@]}")
  local origin="${DEVHUB_E2E_ORIGIN:-}"
  if [[ -n "${origin}" ]]; then
    opts+=(-e "DEVHUB_E2E_ORIGIN=${origin}")
  fi

  # In non-interactive CI or local runs, it's easy to forget starting the Go server.
  # When origin is set, do a lightweight /api/v1/health probe to fail fast with a clear hint.
  #
  # NOTE: Do NOT rely on host.docker.internal always being reachable (some Linux setups map it
  # to docker0 gateway but do not route container->host traffic). Also avoid proxy env impact.
  if [[ -n "${origin}" ]]; then
    if ! command -v curl >/dev/null 2>&1; then
      echo "${C_YELLOW}提示：未找到 curl，跳过 ${origin}/api/v1/health 探测。${C_RESET}"
    else
      if ! curl --noproxy '*' -fsS "${origin}/api/v1/health" >/dev/null 2>&1; then
        echo "${C_YELLOW}提示：无法探测 ${origin}/api/v1/health（可能是容器到宿主网络不可达或后端未启动）。${C_RESET}"
        echo "${C_YELLOW}将继续执行 E2E；如失败为 ERR_CONNECTION_REFUSED，请先启动后端并检查 DEVHUB_E2E_ORIGIN。${C_RESET}"
      fi
    fi
  fi

  run_step "${title} E2E" "${DC[@]}" run "${opts[@]}" "$service"
}

echo "${C_BLUE}DevHub 前台/后台前端检查开始${C_RESET}"
echo "项目目录：${REPO_ROOT}"
echo "日志目录：${LOG_DIR}"
echo "Compose：$("${DC[@]}" version 2>/dev/null | head -n 1 || true)"
echo

HAS_FAIL=0

if [[ "${RUN_ADMIN}" -eq 1 ]]; then
  ADMIN_SERVICE="admin-e2e"
  ADMIN_TITLE="后台 web/admin-app"
  ADMIN_PACKAGE="web/admin-app/package.json"

  prepare_generated_dirs admin

  if [[ "${REBUILD}" -eq 1 ]]; then
    run_compose_build "${ADMIN_SERVICE}" "${ADMIN_TITLE}" || HAS_FAIL=1
  fi

  if [[ "${RUN_OPTIONAL}" -eq 1 ]]; then
    run_npm_script_if_exists "${ADMIN_SERVICE}" "${ADMIN_TITLE}" "${ADMIN_PACKAGE}" "lint" || HAS_FAIL=1
    run_npm_script_if_exists "${ADMIN_SERVICE}" "${ADMIN_TITLE}" "${ADMIN_PACKAGE}" "typecheck" || HAS_FAIL=1
  fi

  if [[ "${RUN_BUILD}" -eq 1 ]]; then
    run_npm_script_if_exists "${ADMIN_SERVICE}" "${ADMIN_TITLE}" "${ADMIN_PACKAGE}" "build" || HAS_FAIL=1
  fi

  if [[ "${RUN_E2E}" -eq 1 ]]; then
    run_e2e_default "${ADMIN_SERVICE}" "${ADMIN_TITLE}" || HAS_FAIL=1
  fi
fi

if [[ "${RUN_FRONTEND}" -eq 1 ]]; then
  FRONTEND_SERVICE="frontend-e2e"
  FRONTEND_TITLE="前台 web/frontend-app"
  FRONTEND_PACKAGE="web/frontend-app/package.json"

  prepare_generated_dirs frontend

  if [[ "${REBUILD}" -eq 1 ]]; then
    run_compose_build "${FRONTEND_SERVICE}" "${FRONTEND_TITLE}" || HAS_FAIL=1
  fi

  if [[ "${RUN_OPTIONAL}" -eq 1 ]]; then
    run_npm_script_if_exists "${FRONTEND_SERVICE}" "${FRONTEND_TITLE}" "${FRONTEND_PACKAGE}" "lint" || HAS_FAIL=1
    run_npm_script_if_exists "${FRONTEND_SERVICE}" "${FRONTEND_TITLE}" "${FRONTEND_PACKAGE}" "typecheck" || HAS_FAIL=1
  fi

  if [[ "${RUN_BUILD}" -eq 1 ]]; then
    run_npm_script_if_exists "${FRONTEND_SERVICE}" "${FRONTEND_TITLE}" "${FRONTEND_PACKAGE}" "build" || HAS_FAIL=1
  fi

  if [[ "${RUN_E2E}" -eq 1 ]]; then
    run_e2e_default "${FRONTEND_SERVICE}" "${FRONTEND_TITLE}" || HAS_FAIL=1
  fi
fi

echo
echo "${C_BLUE}================ 检查结果汇总 ================${C_RESET}"
printf "%-6s | %-36s | %-8s | %s\n" "状态" "检查项" "耗时" "日志/说明"
printf "%s\n" "-------|--------------------------------------|----------|-----------------------------"

for i in "${!RESULT_NAMES[@]}"; do
  name="${RESULT_NAMES[$i]}"
  status="${RESULT_STATUS[$i]}"
  seconds="${RESULT_SECONDS[$i]}"
  log="${RESULT_LOGS[$i]}"
  note="${RESULT_NOTES[$i]}"

  case "$status" in
    PASS)
      display_status="${C_GREEN}PASS${C_RESET}"
      ;;
    FAIL)
      display_status="${C_RED}FAIL${C_RESET}"
      ;;
    SKIP)
      display_status="${C_YELLOW}SKIP${C_RESET}"
      ;;
    *)
      display_status="$status"
      ;;
  esac

  if [[ "$log" == "-" ]]; then
    msg="$note"
  elif [[ -n "$note" ]]; then
    msg="${log} ${note}"
  else
    msg="$log"
  fi

  printf "%-15b | %-36s | %-8s | %s\n" "$display_status" "$name" "${seconds}s" "$msg"
done

echo
echo "日志目录：${LOG_DIR}"

if [[ "${HAS_FAIL}" -eq 0 ]]; then
  echo "${C_GREEN}全部必要检查通过。${C_RESET}"
  exit 0
else
  echo "${C_RED}存在失败检查，请先查看上方 FAIL 项对应日志。${C_RESET}"
  exit 1
fi

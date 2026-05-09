#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ADMIN_APP_DIR="$ROOT_DIR/web/admin-app"
ADMIN_DIST_DIR="$ROOT_DIR/web/admin-vue"
FRONTEND_APP_DIR="$ROOT_DIR/web/frontend-app"
FRONTEND_DIST_DIR="$ROOT_DIR/web/frontend"
RUNTIME_DIR="$ROOT_DIR/.devhub"
PID_FILE="$RUNTIME_DIR/devhub.pid"

PORT="${PORT:-8090}"
ACTION="start"
SERVER_ALREADY_RUNNING=0
RESTART_EXISTING=0
STORE="${CMS_STORE:-memory}"
BUILD_ADMIN=1
BUILD_FRONTEND=1
USE_MYSQL=0
USE_LOCAL_GO=0
SKIP_NPM_INSTALL="${SKIP_NPM_INSTALL:-0}"
FORCE_NPM_INSTALL="${FORCE_NPM_INSTALL:-0}"
NODE_IMAGE="${NODE_IMAGE:-node:20-alpine}"
GO_IMAGE="${GO_IMAGE:-golang:1.22-alpine}"
DOCKER="${DOCKER:-docker}"
SUDO="${SUDO:-sudo}"
MYSQL_WAIT_TIMEOUT="${MYSQL_WAIT_TIMEOUT:-60}"
NPM_CONFIG_REGISTRY="${NPM_CONFIG_REGISTRY:-https://registry.npmmirror.com}"
GOPROXY="${GOPROXY:-https://goproxy.cn,direct}"
GOSUMDB="${GOSUMDB:-sum.golang.google.cn}"
FRONTEND_API_BASE="${FRONTEND_API_BASE:-}"
FRONTEND_SITE_URL_INPUT="${FRONTEND_SITE_URL:-}"
FRONTEND_SITE_URL="${FRONTEND_SITE_URL_INPUT:-http://127.0.0.1:${PORT}}"
DB_HOST="${DB_HOST:-127.0.0.1}"
DB_PORT="${DB_PORT:-3307}"
DB_USER="${DB_USER:-devhub}"
DB_PASSWORD="${DB_PASSWORD:-Devhub_123456}"
DB_NAME="${DB_NAME:-devhub}"
BOOTSTRAP_CONTAINER=""
BOOTSTRAP_PID=""

usage() {
  cat <<EOF
Usage:
  ./dev.sh [start] [options]
  ./dev.sh stop [options]
  ./dev.sh restart [options]
  ./dev.sh status [options]

Options:
  --mysql           Start MySQL with docker compose and use CMS_STORE=mysql.
  --memory          Use in-memory store. This is the default.
  --no-build        Skip both Astro frontend and Vue admin builds.
  --no-build-frontend
                    Skip Astro frontend build.
  --no-build-admin  Skip Vue admin build.
  --skip-npm-install
                    Reuse existing node_modules and skip npm ci/install.
  --force-npm-install
                    Reinstall npm dependencies even when node_modules is present.
  --local-go        Run Go service with host go command instead of Docker Go.
  --restart         If DevHub is already running on the port, stop it and start a fresh server.
  --port PORT       Set Go server port. Default: 8090.
  -h, --help        Show this help.

Environment:
  PORT=8090
  CMS_STORE=memory|mysql
  SKIP_NPM_INSTALL=1
  FORCE_NPM_INSTALL=1
  FRONTEND_API_BASE=
  FRONTEND_SITE_URL=http://127.0.0.1:${PORT}
  NODE_IMAGE=node:20-alpine
  GO_IMAGE=golang:1.22-alpine
  NPM_CONFIG_REGISTRY=https://registry.npmmirror.com
  GOPROXY=https://goproxy.cn,direct
  GOSUMDB=sum.golang.google.cn
  DOCKER=docker
  SUDO=sudo
  MYSQL_WAIT_TIMEOUT=60

URLs:
  Frontend: http://127.0.0.1:${PORT}/
  Admin:    http://127.0.0.1:${PORT}/admin-next

Examples:
  ./dev.sh
  ./dev.sh restart --no-build
  ./dev.sh --local-go restart --no-build
  ./dev.sh stop
  ./dev.sh status
  DOCKER="sudo docker" ./dev.sh start --mysql
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    start)
      ACTION="start"
      shift
      ;;
    stop)
      ACTION="stop"
      shift
      ;;
    restart)
      ACTION="restart"
      RESTART_EXISTING=1
      shift
      ;;
    status)
      ACTION="status"
      shift
      ;;
    --mysql)
      USE_MYSQL=1
      STORE="mysql"
      shift
      ;;
    --memory)
      USE_MYSQL=0
      STORE="memory"
      shift
      ;;
    --no-build)
      BUILD_FRONTEND=0
      BUILD_ADMIN=0
      shift
      ;;
    --no-build-frontend)
      BUILD_FRONTEND=0
      shift
      ;;
    --no-build-admin)
      BUILD_ADMIN=0
      shift
      ;;
    --skip-npm-install)
      SKIP_NPM_INSTALL=1
      shift
      ;;
    --force-npm-install)
      FORCE_NPM_INSTALL=1
      shift
      ;;
    --local-go)
      USE_LOCAL_GO=1
      shift
      ;;
    --restart)
      RESTART_EXISTING=1
      shift
      ;;
    --stop)
      ACTION="stop"
      shift
      ;;
    --status)
      ACTION="status"
      shift
      ;;
    --port)
      PORT="${2:?missing port value}"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown option: $1" >&2
      usage
      exit 1
      ;;
  esac
done

need_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Missing required command: $1" >&2
    exit 1
  fi
}

log() {
  printf '\033[1;34m==>\033[0m %s\n' "$*"
}

warn() {
  printf '\033[1;33mWARN:\033[0m %s\n' "$*" >&2
}

die() {
  printf '\033[1;31mERROR:\033[0m %s\n' "$*" >&2
  exit 1
}

run_step() {
  local title="$1"
  shift
  log "$title"
  "$@"
}

yes_no() {
  if [[ "$1" == "1" ]]; then
    echo "yes"
  else
    echo "no"
  fi
}

npm_strategy() {
  if [[ "$SKIP_NPM_INSTALL" == "1" ]]; then
    echo "skip"
  elif [[ "$FORCE_NPM_INSTALL" == "1" ]]; then
    echo "force"
  else
    echo "auto"
  fi
}

print_execution_plan() {
  local go_mode="Docker Go ($GO_IMAGE)"
  if [[ "$USE_LOCAL_GO" == "1" ]]; then
    go_mode="local Go"
  fi
  local api_bootstrap="no"
  if [[ "$BUILD_FRONTEND" == "1" && "$ACTION" != "stop" && "$ACTION" != "status" ]]; then
    api_bootstrap="yes"
  fi

  echo "DevHub execution plan"
  echo "  Action:          $ACTION"
  echo "  Port:            $PORT"
  echo "  Store:           $STORE"
  echo "  Go runtime:      $go_mode"
  echo "  MySQL:           $(yes_no "$USE_MYSQL")"
  echo "  Build frontend:  $(yes_no "$BUILD_FRONTEND")"
  echo "  Build admin:     $(yes_no "$BUILD_ADMIN")"
  echo "  npm strategy:    $(npm_strategy)"
  echo "  npm registry:    $NPM_CONFIG_REGISTRY"
  echo "  API before build:$api_bootstrap"
  echo "  URLs:"
  echo "    Frontend:      http://127.0.0.1:$PORT/"
  echo "    Admin:         http://127.0.0.1:$PORT/admin-next"
  echo "  Steps:"
  case "$ACTION" in
    stop)
      echo "    1. Stop DevHub service on port $PORT"
      ;;
    status)
      echo "    1. Check DevHub service status on port $PORT"
      ;;
    *)
      echo "    1. Check project layout"
      if [[ "$USE_MYSQL" == "1" ]]; then
        echo "    2. Start MySQL and wait until ready"
      else
        echo "    2. Use memory store"
      fi
      echo "    3. Check/release port $PORT"
      if [[ "$BUILD_FRONTEND" == "1" ]]; then
        echo "    4. Start temporary API and verify seed data"
        echo "    5. Build Astro frontend against the running API"
        echo "    6. Build Vue admin"
        echo "    7. Stop temporary API"
        echo "    8. Start final Go service"
      else
        echo "    4. Skip temporary API bootstrap"
        echo "    5. Build requested assets"
        echo "    6. Start final Go service"
      fi
      ;;
  esac
  echo
}

npm_install_cmd() {
  if [[ -f package-lock.json ]]; then
    echo "npm ci --registry=$NPM_CONFIG_REGISTRY --prefer-offline --no-audit --progress=false --loglevel=info"
  else
    echo "npm install --registry=$NPM_CONFIG_REGISTRY --prefer-offline --no-audit --progress=false --loglevel=info"
  fi
}

npm_dependency_hash() {
  local app_dir="$1"
  (
    cd "$app_dir"
    if [[ -f package-lock.json ]]; then
      sha256sum package-lock.json package.json 2>/dev/null
    else
      sha256sum package.json 2>/dev/null
    fi
  ) | sha256sum | awk '{print $1}'
}

npm_install_marker() {
  local app_dir="$1"
  mkdir -p "$RUNTIME_DIR"
  echo "$RUNTIME_DIR/npm-$(basename "$app_dir").sha256"
}

npm_install_needed() {
  local app_dir="$1"
  local name="$2"
  if [[ "$SKIP_NPM_INSTALL" == "1" ]]; then
    log "Skipping $name dependency install by request"
    return 1
  fi
  if [[ "$FORCE_NPM_INSTALL" == "1" ]]; then
    log "$name dependency install forced"
    return 0
  fi
  if [[ ! -d "$app_dir/node_modules" ]]; then
    log "$name node_modules missing; dependencies will be installed"
    return 0
  fi

  local marker
  marker="$(npm_install_marker "$app_dir")"
  local current
  current="$(npm_dependency_hash "$app_dir")"
  local previous=""
  if [[ -f "$marker" ]]; then
    previous="$(cat "$marker")"
  fi
  if [[ "$current" != "$previous" ]]; then
    log "$name dependency manifest changed; dependencies will be installed"
    return 0
  fi

  log "$name dependencies are up to date; skipping npm install"
  return 1
}

write_npm_install_marker() {
  local app_dir="$1"
  npm_dependency_hash "$app_dir" > "$(npm_install_marker "$app_dir")"
}

has_sudo() {
  command -v "$SUDO" >/dev/null 2>&1
}

docker_cmd() {
  if $DOCKER version >/dev/null 2>&1; then
    $DOCKER "$@"
    return
  fi
  if [[ "$DOCKER" == "docker" ]] && has_sudo; then
    $SUDO docker "$@"
    return
  fi
  return 1
}

docker_cmd_quiet() {
  if $DOCKER version >/dev/null 2>&1; then
    $DOCKER "$@" >/dev/null 2>&1
    return 0
  fi
  if [[ "$DOCKER" == "docker" ]] && has_sudo; then
    $SUDO docker "$@" >/dev/null 2>&1
    return
  fi
  return 1
}

compose_cmd() {
  if $DOCKER compose version >/dev/null 2>&1; then
    $DOCKER compose "$@"
    return
  fi
  if [[ "$DOCKER" == "docker" ]] && has_sudo; then
    $SUDO docker compose "$@"
    return
  fi
  return 1
}

compose_cmd_quiet() {
  if $DOCKER compose version >/dev/null 2>&1; then
    $DOCKER compose "$@" >/dev/null 2>&1
    return 0
  fi
  if [[ "$DOCKER" == "docker" ]] && has_sudo; then
    $SUDO docker compose "$@" >/dev/null 2>&1
    return
  fi
  return 1
}

docker_preflight() {
  log "Checking Docker access: $DOCKER"
  if ! docker_cmd_quiet version; then
    die "Docker is not available. If your user needs sudo, make sure sudo is available or run: DOCKER=\"sudo docker\" ./dev.sh"
  fi
  if ! compose_cmd_quiet version; then
    die "Docker Compose plugin is not available for: $DOCKER compose"
  fi
}

check_project_layout() {
  [[ -d "$FRONTEND_APP_DIR" ]] || die "Missing frontend app directory: $FRONTEND_APP_DIR"
  [[ -d "$ADMIN_APP_DIR" ]] || die "Missing admin app directory: $ADMIN_APP_DIR"
  [[ -f "$ROOT_DIR/go.mod" ]] || die "Missing go.mod in $ROOT_DIR"
}

is_port_in_use() {
  local port="$1"
  local hex_port
  hex_port="$(printf '%04X' "$port")"
  if [[ -r /proc/net/tcp ]] && awk -v port=":${hex_port}" 'tolower($2) ~ tolower(port) && $4 == "0A" { found=1 } END { exit found ? 0 : 1 }' /proc/net/tcp; then
    return 0
  fi
  if [[ -r /proc/net/tcp6 ]] && awk -v port=":${hex_port}" 'tolower($2) ~ tolower(port) && $4 == "0A" { found=1 } END { exit found ? 0 : 1 }' /proc/net/tcp6; then
    return 0
  fi
  if command -v lsof >/dev/null 2>&1; then
    if lsof -nP -iTCP:"$port" -sTCP:LISTEN >/dev/null 2>&1; then
      return 0
    fi
    if has_sudo && $SUDO lsof -nP -iTCP:"$port" -sTCP:LISTEN >/dev/null 2>&1; then
      return 0
    fi
  fi
  if command -v ss >/dev/null 2>&1; then
    ss -ltn | grep -q ":$port "
    return
  fi
  if command -v netstat >/dev/null 2>&1; then
    netstat -ltn 2>/dev/null | grep -q ":$port "
    return
  fi
  if command -v timeout >/dev/null 2>&1; then
    timeout 1 bash -c ":</dev/tcp/127.0.0.1/$port" >/dev/null 2>&1
    return
  fi
  warn "Cannot check whether port $port is free because lsof, ss, netstat and timeout are unavailable."
  return 1
}

check_port_available() {
  if ! is_port_in_use "$PORT"; then
    return
  fi
  if [[ "$RESTART_EXISTING" == "1" ]]; then
    stop_existing_service "$PORT"
    sleep 1
    if ! is_port_in_use "$PORT"; then
      return
    fi
    die "Port $PORT is still in use after restart attempt. Please stop it manually and run ./dev.sh again."
  fi
  if is_devhub_running "$PORT"; then
    SERVER_ALREADY_RUNNING=1
    warn "Port $PORT is already used by a running DevHub service; this script will reuse it and skip starting another Go server."
    return
  fi
  SERVER_ALREADY_RUNNING=1
  warn "Port $PORT is already in use, but the health check did not confirm it is DevHub."
  warn "To avoid starting a duplicate server, this script will skip starting Go."
  warn "If http://127.0.0.1:$PORT/ is not DevHub, stop the process shown below and run ./dev.sh again."
  show_port_in_use_help "$PORT"
  return
}

stop_existing_service() {
  local port="$1"
  log "Stopping existing DevHub service on port $port"
  stop_pid_file_process
  stop_docker_containers_on_port "$port"
  stop_local_processes_on_port "$port"
  if ! wait_for_port_release "$port"; then
    warn "Port $port is still in use after stop attempt."
  fi
  return 0
}

stop_pid_file_process() {
  if [[ ! -f "$PID_FILE" ]]; then
    return
  fi
  local pid
  pid="$(cat "$PID_FILE" 2>/dev/null || true)"
  if [[ -z "$pid" ]]; then
    rm -f "$PID_FILE"
    return
  fi
  if kill -0 "$pid" 2>/dev/null; then
    log "Stopping PID from $PID_FILE: $pid"
    kill "$pid" 2>/dev/null || true
    sleep 1
    if kill -0 "$pid" 2>/dev/null; then
      warn "Process $pid is still running; sending SIGKILL"
      kill -9 "$pid" 2>/dev/null || true
    fi
  elif has_sudo && $SUDO kill -0 "$pid" 2>/dev/null; then
    log "Stopping PID from $PID_FILE with sudo: $pid"
    $SUDO kill "$pid" 2>/dev/null || true
    sleep 1
    if $SUDO kill -0 "$pid" 2>/dev/null; then
      warn "Process $pid is still running; sending sudo SIGKILL"
      $SUDO kill -9 "$pid" 2>/dev/null || true
    fi
  fi
  rm -f "$PID_FILE"
}

stop_docker_containers_on_port() {
  local port="$1"
  if ! docker_cmd_quiet version; then
    return
  fi
  local containers
  containers="$(docker_cmd ps --filter "publish=${port}" --format '{{.ID}}' 2>/dev/null || true)"
  if [[ -z "$containers" ]]; then
    return
  fi
  echo "$containers" | while read -r container; do
    [[ -z "$container" ]] && continue
    log "Stopping Docker container $container"
    docker_cmd stop "$container" >/dev/null
  done
}

stop_local_processes_on_port() {
  local port="$1"
  local pids=""
  if command -v lsof >/dev/null 2>&1; then
    pids="$(lsof -tiTCP:"$port" -sTCP:LISTEN 2>/dev/null || true)"
    if [[ -z "$pids" ]] && has_sudo; then
      pids="$($SUDO lsof -tiTCP:"$port" -sTCP:LISTEN 2>/dev/null || true)"
    fi
  elif command -v fuser >/dev/null 2>&1; then
    pids="$(fuser "${port}/tcp" 2>/dev/null || true)"
    if [[ -z "$pids" ]] && has_sudo; then
      pids="$($SUDO fuser "${port}/tcp" 2>/dev/null || true)"
    fi
  fi
  if [[ -z "$pids" ]]; then
    return
  fi
  echo "$pids" | tr ' ' '\n' | while read -r pid; do
    [[ -z "$pid" ]] && continue
    log "Stopping local process $pid"
    kill "$pid" 2>/dev/null || {
      if has_sudo; then
        $SUDO kill "$pid" 2>/dev/null || true
      fi
    }
  done
}

wait_for_port_release() {
  local port="$1"
  local elapsed=0
  while (( elapsed < 15 )); do
    if ! is_port_in_use "$port"; then
      return 0
    fi
    sleep 1
    elapsed=$((elapsed + 1))
  done
  return 1
}

status_service() {
  echo "DevHub status"
  echo "  Port: $PORT"
  if [[ -f "$PID_FILE" ]]; then
    local pid
    pid="$(cat "$PID_FILE" 2>/dev/null || true)"
    if [[ -n "$pid" ]] && kill -0 "$pid" 2>/dev/null; then
      echo "  PID file: $PID_FILE -> running ($pid)"
    else
      echo "  PID file: $PID_FILE -> stale"
    fi
  else
    echo "  PID file: none"
  fi
  if is_devhub_running "$PORT"; then
    echo "  HTTP: DevHub is responding at http://127.0.0.1:$PORT/"
  elif is_port_in_use "$PORT"; then
    echo "  HTTP: port is in use, but DevHub health check did not pass"
    show_port_in_use_help "$PORT"
  else
    echo "  HTTP: stopped"
  fi
}

is_devhub_running() {
  local port="$1"
  local url="http://127.0.0.1:${port}/api/v1/health"
  local body=""
  if command -v curl >/dev/null 2>&1; then
    body="$(curl -fsS --max-time 2 "$url" 2>/dev/null || true)"
  elif command -v wget >/dev/null 2>&1; then
    body="$(wget -qO- --timeout=2 "$url" 2>/dev/null || true)"
  fi
  [[ "$body" == *'"ok":true'* || "$body" == *'"data_source"'* || "$body" == *'"database"'* ]]
}

show_port_in_use_help() {
  local port="$1"
  printf '\033[1;31mERROR:\033[0m Port %s is already in use.\n\n' "$port" >&2
  echo "How to find and stop it:" >&2
  echo "  1. If it is a local process:" >&2
  echo "     lsof -nP -iTCP:${port} -sTCP:LISTEN" >&2
  echo "     kill <PID>" >&2
  echo >&2
  echo "  2. If it is a Docker container:" >&2
  echo "     ${DOCKER} ps --filter publish=${port}" >&2
  echo "     ${DOCKER} stop <CONTAINER_ID>" >&2
  echo "     sudo docker ps --filter publish=${port}" >&2
  echo "     sudo docker stop <CONTAINER_ID>" >&2
  echo >&2
  echo "  3. If it was started by this script in the foreground:" >&2
  echo "     Press Ctrl+C in the terminal running DevHub." >&2
  echo >&2
  echo "After stopping it, run ./dev.sh again." >&2
}

refresh_frontend_site_url() {
  if [[ -z "$FRONTEND_SITE_URL_INPUT" ]]; then
    FRONTEND_SITE_URL="http://127.0.0.1:${PORT}"
  else
    FRONTEND_SITE_URL="$FRONTEND_SITE_URL_INPUT"
  fi
}

check_static_output() {
  local name="$1"
  local file="$2"
  if [[ ! -f "$file" ]]; then
    warn "$name output is missing: $file"
    warn "Run without --no-build, or build it manually before starting."
  fi
}

build_admin_with_local_npm() {
  log "Building Vue admin with local npm"
  cd "$ADMIN_APP_DIR"
  if npm_install_needed "$ADMIN_APP_DIR" "Vue admin"; then
    log "Installing Vue admin dependencies from $NPM_CONFIG_REGISTRY"
    $(npm_install_cmd)
    write_npm_install_marker "$ADMIN_APP_DIR"
  fi
  log "Running Vue admin build"
  npm run build
}

build_frontend_with_local_npm() {
  log "Building Astro frontend with local npm"
  cd "$FRONTEND_APP_DIR"
  if npm_install_needed "$FRONTEND_APP_DIR" "Astro frontend"; then
    log "Installing Astro frontend dependencies from $NPM_CONFIG_REGISTRY"
    $(npm_install_cmd)
    write_npm_install_marker "$FRONTEND_APP_DIR"
  fi
  log "Running Astro frontend build"
  FRONTEND_API_BASE="$FRONTEND_API_BASE" \
  FRONTEND_SITE_URL="$FRONTEND_SITE_URL" \
  npm run build
}

build_admin_with_docker_node() {
  log "Building Vue admin with Docker image: $NODE_IMAGE"
  docker_preflight
  mkdir -p "$ADMIN_DIST_DIR"
  local install_cmd
  local did_install=0
  install_cmd="$(cd "$ADMIN_APP_DIR" && npm_install_cmd)"
  local node_cmd="echo '==> Running Vue admin build' && npm run build"
  if npm_install_needed "$ADMIN_APP_DIR" "Vue admin"; then
    did_install=1
    node_cmd="echo '==> Installing Vue admin dependencies from $NPM_CONFIG_REGISTRY' && $install_cmd && echo '==> Running Vue admin build' && npm run build"
  fi
  docker_cmd run --rm \
    -e NPM_CONFIG_REGISTRY="$NPM_CONFIG_REGISTRY" \
    -v "$ADMIN_APP_DIR:/app" \
    -v "$ADMIN_DIST_DIR:/admin-vue" \
    -w /app \
    "$NODE_IMAGE" \
    sh -c "$node_cmd"
  if [[ "$did_install" == "1" ]]; then
    write_npm_install_marker "$ADMIN_APP_DIR"
  fi
}

build_frontend_with_docker_node() {
  log "Building Astro frontend with Docker image: $NODE_IMAGE"
  docker_preflight
  mkdir -p "$FRONTEND_DIST_DIR"
  local install_cmd
  local did_install=0
  install_cmd="$(cd "$FRONTEND_APP_DIR" && npm_install_cmd)"
  local node_cmd="echo '==> Running Astro frontend build' && npm run build"
  if npm_install_needed "$FRONTEND_APP_DIR" "Astro frontend"; then
    did_install=1
    node_cmd="echo '==> Installing Astro frontend dependencies from $NPM_CONFIG_REGISTRY' && $install_cmd && echo '==> Running Astro frontend build' && npm run build"
  fi
  local container_frontend_api_base="$FRONTEND_API_BASE"
  if [[ "$container_frontend_api_base" == "http://127.0.0.1:${PORT}" || "$container_frontend_api_base" == "http://localhost:${PORT}" ]]; then
    container_frontend_api_base="http://host.docker.internal:${PORT}"
  fi
  docker_cmd run --rm \
    -e NPM_CONFIG_REGISTRY="$NPM_CONFIG_REGISTRY" \
    -e FRONTEND_API_BASE="$container_frontend_api_base" \
    -e FRONTEND_SITE_URL="$FRONTEND_SITE_URL" \
    --add-host=host.docker.internal:host-gateway \
    -v "$FRONTEND_APP_DIR:/app" \
    -v "$FRONTEND_DIST_DIR:/frontend" \
    -w /app \
    "$NODE_IMAGE" \
    sh -c "$node_cmd"
  if [[ "$did_install" == "1" ]]; then
    write_npm_install_marker "$FRONTEND_APP_DIR"
  fi
}

build_frontend() {
  if [[ "$BUILD_FRONTEND" != "1" ]]; then
    log "Skipping Astro frontend build"
    check_static_output "Astro frontend" "$FRONTEND_DIST_DIR/index.html"
    return
  fi
  if command -v npm >/dev/null 2>&1; then
    build_frontend_with_local_npm
  else
    build_frontend_with_docker_node
  fi
}

build_admin() {
  if [[ "$BUILD_ADMIN" != "1" ]]; then
    log "Skipping Vue admin build"
    check_static_output "Vue admin" "$ADMIN_DIST_DIR/index.html"
    return
  fi
  if command -v npm >/dev/null 2>&1; then
    build_admin_with_local_npm
  else
    build_admin_with_docker_node
  fi
}

start_mysql_if_needed() {
  if [[ "$USE_MYSQL" != "1" ]]; then
    return
  fi
  docker_preflight
  log "Starting MySQL"
  cd "$ROOT_DIR"
  compose_cmd -f docker-compose.dev.yml up -d
  wait_for_mysql
}

wait_for_mysql() {
  log "Waiting for MySQL to become ready"
  local elapsed=0
  while (( elapsed < MYSQL_WAIT_TIMEOUT )); do
    if compose_cmd_quiet -f "$ROOT_DIR/docker-compose.dev.yml" exec -T mysql \
      mysqladmin ping -h127.0.0.1 -u"$DB_USER" -p"$DB_PASSWORD" --silent >/dev/null 2>&1; then
      log "MySQL is ready"
      return
    fi
    sleep 2
    elapsed=$((elapsed + 2))
  done
  die "MySQL did not become ready within ${MYSQL_WAIT_TIMEOUT}s. Check with: $DOCKER compose -f docker-compose.dev.yml logs mysql"
}

api_base_for_build() {
  if [[ -n "$FRONTEND_API_BASE" ]]; then
    echo "$FRONTEND_API_BASE"
  else
    echo "http://127.0.0.1:${PORT}"
  fi
}

api_get() {
  local path="$1"
  local base
  base="$(api_base_for_build)"
  if command -v curl >/dev/null 2>&1; then
    curl -fsS --max-time 5 "${base}${path}"
  elif command -v wget >/dev/null 2>&1; then
    wget -qO- --timeout=5 "${base}${path}"
  else
    die "curl or wget is required for API readiness checks."
  fi
}

check_api_non_empty_items() {
  local path="$1"
  local label="$2"
  local body
  body="$(api_get "$path" 2>/dev/null || true)"
  if [[ "$body" != *'"items":['* || "$body" == *'"items":[]'* ]]; then
    die "$label returned no items before frontend build. Check seed data/API: $(api_base_for_build)$path"
  fi
}

verify_api_for_frontend_build() {
  log "Verifying API data before frontend build"
  if ! is_devhub_running "$PORT"; then
    die "DevHub API is not healthy before frontend build: http://127.0.0.1:$PORT/api/v1/health"
  fi
  check_api_non_empty_items "/api/v1/communities" "communities API"
  check_api_non_empty_items "/api/v1/topics?page_size=1" "topics API"
  check_api_non_empty_items "/api/v1/communities/php/tags" "PHP tags API"
  log "API seed data is ready for frontend build"
}

start_bootstrap_api() {
  if [[ "$BUILD_FRONTEND" != "1" ]]; then
    return
  fi
  if [[ "$SERVER_ALREADY_RUNNING" == "1" ]]; then
    log "Using existing DevHub API for frontend build"
    verify_api_for_frontend_build
    if [[ -z "$FRONTEND_API_BASE" ]]; then
      FRONTEND_API_BASE="$(api_base_for_build)"
    fi
    return
  fi
  log "Starting temporary API for frontend build"
  if [[ "$USE_LOCAL_GO" == "1" ]]; then
    start_bootstrap_api_with_local_go
  else
    start_bootstrap_api_with_docker_go
  fi
  wait_for_http_ready
  verify_api_for_frontend_build
  if [[ -z "$FRONTEND_API_BASE" ]]; then
    FRONTEND_API_BASE="$(api_base_for_build)"
  fi
}

start_bootstrap_api_with_local_go() {
  if ! command -v go >/dev/null 2>&1; then
    die "Local Go command was requested with --local-go, but go was not found in PATH."
  fi
  cd "$ROOT_DIR"
  (
    PORT="$PORT" \
    CMS_STORE="$STORE" \
    DB_HOST="$DB_HOST" \
    DB_PORT="$DB_PORT" \
    DB_USER="$DB_USER" \
    DB_PASSWORD="$DB_PASSWORD" \
    DB_NAME="$DB_NAME" \
    GOPROXY="$GOPROXY" \
    GOSUMDB="$GOSUMDB" \
    GOCACHE="${GOCACHE:-/tmp/go-build}" \
    go run .
  ) &
  BOOTSTRAP_PID="$!"
}

start_bootstrap_api_with_docker_go() {
  docker_preflight
  local container_db_host="$DB_HOST"
  if [[ "$STORE" == "mysql" && "$DB_HOST" == "127.0.0.1" ]]; then
    container_db_host="host.docker.internal"
  fi
  BOOTSTRAP_CONTAINER="devhub-bootstrap-api-$PORT"
  docker_cmd rm -f "$BOOTSTRAP_CONTAINER" >/dev/null 2>&1 || true
  docker_cmd run -d --rm \
    --name "$BOOTSTRAP_CONTAINER" \
    -p "$PORT:$PORT" \
    -e "PORT=$PORT" \
    -e "CMS_STORE=$STORE" \
    -e "DB_HOST=$container_db_host" \
    -e "DB_PORT=$DB_PORT" \
    -e "DB_USER=$DB_USER" \
    -e "DB_PASSWORD=$DB_PASSWORD" \
    -e "DB_NAME=$DB_NAME" \
    -e "GOPROXY=$GOPROXY" \
    -e "GOSUMDB=$GOSUMDB" \
    -e "GOCACHE=/tmp/go-build" \
    -e "GOMODCACHE=/go/pkg/mod" \
    --add-host=host.docker.internal:host-gateway \
    -v "$ROOT_DIR:/app" \
    -v "devhub_go_mod_cache:/go/pkg/mod" \
    -v "devhub_go_build_cache:/tmp/go-build" \
    -w /app \
    "$GO_IMAGE" \
    go run . >/dev/null
}

stop_bootstrap_api() {
  if [[ -n "$BOOTSTRAP_PID" ]]; then
    log "Stopping temporary API process $BOOTSTRAP_PID"
    kill "$BOOTSTRAP_PID" 2>/dev/null || true
    wait "$BOOTSTRAP_PID" 2>/dev/null || true
    BOOTSTRAP_PID=""
  fi
  if [[ -n "$BOOTSTRAP_CONTAINER" ]]; then
    log "Stopping temporary API container $BOOTSTRAP_CONTAINER"
    docker_cmd stop "$BOOTSTRAP_CONTAINER" >/dev/null 2>&1 || true
    BOOTSTRAP_CONTAINER=""
  fi
  wait_for_port_release "$PORT" >/dev/null 2>&1 || true
}

start_go() {
  if [[ "$SERVER_ALREADY_RUNNING" == "1" ]]; then
    log "Skipping Go server start because port $PORT is already occupied"
    echo "    Frontend: http://127.0.0.1:$PORT/"
    echo "    Admin:    http://127.0.0.1:$PORT/admin-next"
    echo
    warn "If backend Go code changed, stop the existing DevHub process and run ./dev.sh again."
    return
  fi
  if [[ "$USE_LOCAL_GO" == "1" ]]; then
    if ! command -v go >/dev/null 2>&1; then
      die "Local Go command was requested with --local-go, but go was not found in PATH."
    fi
    start_go_with_local_go
  else
    start_go_with_docker_go
  fi
}

print_urls() {
  log "Starting Go server"
  echo "    Store: $STORE"
  echo "    Port:  $PORT"
  echo
  echo "    Frontend: http://127.0.0.1:$PORT/"
  echo "    Admin:    http://127.0.0.1:$PORT/admin-next"
  echo
}

start_go_with_local_go() {
  log "Starting Go service with local go command"
  print_urls
  cd "$ROOT_DIR"
  mkdir -p "$RUNTIME_DIR"
  (
    PORT="$PORT" \
    CMS_STORE="$STORE" \
    DB_HOST="$DB_HOST" \
    DB_PORT="$DB_PORT" \
    DB_USER="$DB_USER" \
    DB_PASSWORD="$DB_PASSWORD" \
    DB_NAME="$DB_NAME" \
    GOPROXY="$GOPROXY" \
    GOSUMDB="$GOSUMDB" \
    GOCACHE="${GOCACHE:-/tmp/go-build}" \
    go run .
  ) &
  local pid=$!
  echo "$pid" > "$PID_FILE"
  log "DevHub started in background with PID $pid"
  wait_for_http_ready
  echo "Press Ctrl+C is no longer required. Stop with: ./dev.sh stop"
  wait "$pid"
}

start_go_with_docker_go() {
  log "Starting Go service with Docker image: $GO_IMAGE"
  docker_preflight
  print_urls
  local container_db_host="$DB_HOST"
  if [[ "$STORE" == "mysql" && "$DB_HOST" == "127.0.0.1" ]]; then
    container_db_host="host.docker.internal"
  fi
  local docker_args=(
    run --rm
    -p "$PORT:$PORT"
    -e "PORT=$PORT"
    -e "CMS_STORE=$STORE"
    -e "DB_HOST=$container_db_host"
    -e "DB_PORT=$DB_PORT"
    -e "DB_USER=$DB_USER"
    -e "DB_PASSWORD=$DB_PASSWORD"
    -e "DB_NAME=$DB_NAME"
    -e "GOPROXY=$GOPROXY"
    -e "GOSUMDB=$GOSUMDB"
    -e "GOCACHE=/tmp/go-build"
    -e "GOMODCACHE=/go/pkg/mod"
    --add-host=host.docker.internal:host-gateway
    -v "$ROOT_DIR:/app"
    -v "devhub_go_mod_cache:/go/pkg/mod"
    -v "devhub_go_build_cache:/tmp/go-build"
    -w /app
    "$GO_IMAGE"
    go run .
  )
  docker_cmd "${docker_args[@]}"
}

wait_for_http_ready() {
  local elapsed=0
  while (( elapsed < 15 )); do
    if is_devhub_running "$PORT"; then
      log "DevHub is ready"
      return 0
    fi
    sleep 1
    elapsed=$((elapsed + 1))
  done
  warn "DevHub did not pass health check within 15s. Check terminal output above."
}

case "$ACTION" in
  stop)
    print_execution_plan
    stop_existing_service "$PORT"
    if is_port_in_use "$PORT"; then
      warn "Port $PORT is still in use after stop attempt."
      show_port_in_use_help "$PORT"
      exit 1
    fi
    log "DevHub stopped"
    exit 0
    ;;
  status)
    print_execution_plan
    status_service
    exit 0
    ;;
  restart)
    RESTART_EXISTING=1
    ;;
esac

print_execution_plan
check_project_layout
start_mysql_if_needed
check_port_available
refresh_frontend_site_url
trap stop_bootstrap_api EXIT
start_bootstrap_api
run_step "Preparing Astro frontend" build_frontend
run_step "Preparing Vue admin" build_admin
stop_bootstrap_api
trap - EXIT
start_go

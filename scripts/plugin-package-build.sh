#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT_DIR="$ROOT_DIR/.devhub/plugin-packages/dist"
RUN_CHECK=1
ALL_OFFICIAL=0
PACKAGES=()

usage() {
  cat >&2 <<'EOF'
用法:
  scripts/plugin-package-build.sh <plugin-dir> [plugin-dir...] [--out <dir>] [--no-check]
  scripts/plugin-package-build.sh --all-official [--out <dir>] [--no-check]

示例:
  scripts/plugin-package-build.sh examples/plugins/official_links
  scripts/plugin-package-build.sh --all-official

说明:
  只复制插件包文件、重算 checksums.json 并生成 zip。
  不执行 SQL、不执行插件代码、不执行 package scripts、不访问远程市场。
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --out)
      OUT_DIR="$2"
      shift 2
      ;;
    --no-check)
      RUN_CHECK=0
      shift
      ;;
    --all-official)
      ALL_OFFICIAL=1
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    --*)
      echo "未知参数：$1" >&2
      usage
      exit 2
      ;;
    *)
      PACKAGES+=("$1")
      shift
      ;;
  esac
done

if [[ "$ALL_OFFICIAL" == "1" ]]; then
  PACKAGES+=(
    "examples/plugins/official_links"
    "examples/plugins/official_webhook_notify"
    "examples/plugins/templates/declarative-content"
    "examples/plugins/templates/external-service-webhook"
  )
fi

if [[ "${#PACKAGES[@]}" -eq 0 ]]; then
  usage
  exit 2
fi

if ! command -v python3 >/dev/null 2>&1; then
  echo "缺少 python3，无法生成插件包 zip。" >&2
  exit 2
fi

mkdir -p "$OUT_DIR"

build_one() {
  local src="$1"
  python3 - "$ROOT_DIR" "$src" "$OUT_DIR" <<'PY'
import hashlib
import json
import os
import shutil
import sys
import tempfile
import zipfile
from pathlib import Path

root = Path(sys.argv[1]).resolve()
src = (root / sys.argv[2]).resolve() if not Path(sys.argv[2]).is_absolute() else Path(sys.argv[2]).resolve()
out_dir = (root / sys.argv[3]).resolve() if not Path(sys.argv[3]).is_absolute() else Path(sys.argv[3]).resolve()

if not src.exists() or not src.is_dir():
    raise SystemExit(f"插件包目录不存在：{src}")
if not (src / "manifest.json").exists():
    raise SystemExit(f"插件包缺少 manifest.json：{src}")
if root not in src.parents and src != root:
    raise SystemExit(f"插件包目录必须位于仓库内：{src}")

manifest = json.loads((src / "manifest.json").read_text(encoding="utf-8"))
code = str(manifest.get("code") or manifest.get("plugin_code") or "").strip()
version = str(manifest.get("version") or "").strip()
if not code or not version:
    raise SystemExit("manifest.json 必须声明 code/plugin_code 和 version")

out_dir.mkdir(parents=True, exist_ok=True)
with tempfile.TemporaryDirectory(prefix="devhub-plugin-build-") as tmp:
    work = Path(tmp) / code
    def ignore(_dir, names):
        blocked = {".git", ".DS_Store", "__pycache__", "node_modules", "vendor"}
        return [name for name in names if name in blocked]
    shutil.copytree(src, work, ignore=ignore)
    checksum_path = work / "checksums.json"
    if checksum_path.exists():
        checksum_path.unlink()

    files = []
    for path in sorted(p for p in work.rglob("*") if p.is_file()):
        rel = path.relative_to(work).as_posix()
        if rel == "checksums.json":
            continue
        digest = hashlib.sha256(path.read_bytes()).hexdigest()
        files.append({"path": rel, "sha256": digest})
    checksum_path.write_text(json.dumps({"algorithm": "sha256", "files": files}, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")

    zip_path = out_dir / f"{code}-{version}.zip"
    if zip_path.exists():
        zip_path.unlink()
    with zipfile.ZipFile(zip_path, "w", compression=zipfile.ZIP_DEFLATED) as zf:
        for path in sorted(p for p in work.rglob("*") if p.is_file()):
            rel = path.relative_to(work).as_posix()
            info = zipfile.ZipInfo(rel)
            info.date_time = (2026, 1, 1, 0, 0, 0)
            info.external_attr = (0o644 & 0xFFFF) << 16
            zf.writestr(info, path.read_bytes())
    print(zip_path)
PY
}

for pkg in "${PACKAGES[@]}"; do
  zip_path="$(build_one "$pkg")"
  echo "已生成插件包: $zip_path"
  if [[ "$RUN_CHECK" == "1" ]]; then
    "$ROOT_DIR/scripts/plugin-package-check.sh" "$zip_path"
  fi
done

if [[ "$ALL_OFFICIAL" == "1" ]]; then
  echo "说明: official_announcement 是内置插件，不生成 zip；可执行 scripts/plugin-package-check.sh --builtin official_announcement 校验配置模型和官方挂载。"
fi

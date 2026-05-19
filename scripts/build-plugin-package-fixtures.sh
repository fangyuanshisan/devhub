#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
FIXTURE_DIR="$ROOT_DIR/scripts/fixtures/plugin-packages"
SRC_DIR="$FIXTURE_DIR/generated-src"
DIST_DIR="$FIXTURE_DIR/dist"
SUFFIX=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --suffix)
      SUFFIX="${2:-}"
      shift 2
      ;;
    *)
      echo "未知参数：$1" >&2
      exit 2
      ;;
  esac
done

if ! command -v sha256sum >/dev/null 2>&1; then
  echo "缺少 sha256sum 命令，无法生成 checksums.json。" >&2
  exit 1
fi
if ! command -v zip >/dev/null 2>&1 && ! command -v node >/dev/null 2>&1; then
  echo "缺少 zip 或 node 命令，无法生成插件包 fixture。" >&2
  exit 1
fi

suffix_part=""
code_suffix=""
if [[ -n "$SUFFIX" ]]; then
  safe_suffix="$(printf '%s' "$SUFFIX" | tr -cd '[:alnum:]_-' | tr '-' '_')"
  suffix_part="-$safe_suffix"
  code_suffix="_$safe_suffix"
fi

rm -rf "$SRC_DIR"
mkdir -p "$SRC_DIR" "$DIST_DIR"

write_manifest() {
  local dir="$1"
  local code="$2"
  local name="$3"
  mkdir -p "$dir/migrations"
  cat > "$dir/manifest.json" <<JSON
{
  "code": "$code",
  "name": "$name",
  "version": "1.0.0",
  "description": "DevHub S13 真实插件包链路验收 fixture；仅声明元数据，不包含可执行代码。",
  "author": "DevHub Fixture",
  "compatible_core_version": ">=1.4.0",
  "is_system": false,
  "content_types": ["${code}_item"],
  "content_type_definitions": [
    {
      "type": "${code}_item",
      "name": "Fixture Item",
      "plugin_code": "$code",
      "create_permission": "$code.item.create",
      "edit_permission": "$code.item.edit",
      "delete_permission": "$code.item.delete",
      "audit_permission": "$code.item.audit",
      "default_status": "draft",
      "allow_comment": true,
      "allow_like": true,
      "allow_favorite": true,
      "seo_type": "Article"
    }
  ],
  "permissions": [
    { "plugin_code": "$code", "code": "$code.item.create", "name": "创建 Fixture 内容", "scope": "community" },
    { "plugin_code": "$code", "code": "$code.item.edit", "name": "编辑 Fixture 内容", "scope": "own" },
    { "plugin_code": "$code", "code": "$code.item.delete", "name": "删除 Fixture 内容", "scope": "own" },
    { "plugin_code": "$code", "code": "$code.item.audit", "name": "审核 Fixture 内容", "scope": "community" },
    { "plugin_code": "$code", "code": "$code.manage", "name": "管理 Fixture 插件", "scope": "global" }
  ],
  "menus": [
    {
      "plugin_code": "$code",
      "code": "$code.admin",
      "title": "Fixture Plugin",
      "path": "/admin-next/$code",
      "location": "admin",
      "area": "admin",
      "permission": "$code.manage",
      "sort_order": 360
    }
  ],
  "routes": [
    {
      "plugin_code": "$code",
      "area": "admin",
      "method": "GET",
      "path": "/api/v1/admin/$code",
      "handler": "reserved:$code.admin.list",
      "auth": "admin",
      "permission": "$code.manage"
    }
  ],
  "config_schema": {
    "type": "object",
    "additionalProperties": false,
    "properties": {
      "enabled": { "type": "boolean", "title": "启用", "default": true }
    }
  },
  "migrations": [
    {
      "plugin_code": "$code",
      "migration_version": "1.0.0",
      "migration_name": "fixture_init",
      "direction": "up",
      "checksum": "sha256:fixture",
      "tables": ["${code}_items"],
      "rollback_supported": false,
      "description": "声明迁移；真实 SQL 文件只放在 migrations/ 下。"
    }
  ]
}
JSON
  cat > "$dir/migrations/001_init.sql" <<SQL
-- DevHub S13 fixture migration.
-- dry-run 不执行 SQL；install 只允许基于 migrations/ 计划处理。
CREATE TABLE IF NOT EXISTS ${code}_items (
  id BIGINT PRIMARY KEY
);
SQL
  cat > "$dir/README.md" <<MD
# $name

DevHub S13 真实插件包 fixture。该包只用于 upload -> precheck -> promote -> install 链路验收。

- 不包含 package scripts
- 不包含远程代码
- 不包含动态加载入口
- SQL 只放在 migrations/
MD
  cat > "$dir/config.example.json" <<JSON
{ "enabled": true }
JSON
}

write_links_manifest() {
  local dir="$1"
  local code="$2"
  local content_type="$3"
  local table_name="$4"
  mkdir -p "$dir/migrations"
  cat > "$dir/manifest.json" <<JSON
{
  "code": "$code",
  "name": "声明型友情链接插件",
  "version": "1.0.0",
  "description": "DevHub S15 真实声明型插件 fixture，用于验证安装到使用的完整业务闭环。",
  "author": "DevHub Fixture",
  "compatible_core_version": ">=1.4.0",
  "is_system": false,
  "content_types": ["$content_type"],
  "content_type_definitions": [
    {
      "type": "$content_type",
      "name": "友情链接",
      "plugin_code": "$code",
      "create_permission": "$code.link.create",
      "edit_permission": "$code.link.manage",
      "delete_permission": "$code.link.manage",
      "audit_permission": "$code.link.manage",
      "default_status": "publish",
      "allow_comment": false,
      "allow_like": false,
      "allow_favorite": true,
      "seo_type": "Article"
    }
  ],
  "permissions": [
    { "plugin_code": "$code", "code": "$code.link.create", "name": "创建友情链接", "description": "允许在已启用子站创建友情链接内容", "scope": "community" },
    { "plugin_code": "$code", "code": "$code.link.manage", "name": "管理友情链接", "description": "允许管理友情链接内容", "scope": "community" },
    { "plugin_code": "$code", "code": "$code.config.manage", "name": "管理友情链接配置", "description": "允许读取和保存插件配置", "scope": "global" },
    { "plugin_code": "$code", "code": "$code.menu.view", "name": "查看友情链接菜单", "description": "允许查看插件声明菜单", "scope": "community" }
  ],
  "menus": [
    {
      "plugin_code": "$code",
      "code": "$code.admin.links",
      "title": "友情链接管理",
      "path": "/admin-next/plugins/overview?tab=list",
      "location": "admin",
      "area": "admin",
      "permission": "$code.menu.view",
      "sort_order": 365
    },
    {
      "plugin_code": "$code",
      "code": "$code.frontend.links",
      "title": "友情链接",
      "path": "/c/php/",
      "route": "/c/php/",
      "location": "frontend",
      "area": "frontend",
      "permission": "$code.menu.view",
      "content_type": "$content_type",
      "require_community_enabled": true,
      "require_category_binding": true,
      "visible_when": ["plugin_enabled", "community_enabled", "config_valid"],
      "sort_order": 120
    }
  ],
  "routes": [
    {
      "plugin_code": "$code",
      "area": "admin",
      "method": "GET",
      "path": "/api/v1/admin/plugins/overview?tab=list",
      "handler": "reserved:$code.admin.links",
      "auth": "admin",
      "permission": "$code.menu.view"
    }
  ],
  "config_schema": {
    "type": "object",
    "additionalProperties": false,
    "required": ["enabled", "title", "max_links", "display_position"],
    "properties": {
      "enabled": { "type": "boolean", "title": "启用友情链接", "default": true },
      "title": { "type": "string", "title": "展示标题", "default": "友情链接", "minLength": 2, "maxLength": 30 },
      "max_links": { "type": "integer", "title": "最大链接数", "default": 10, "minimum": 1, "maximum": 100 },
      "display_position": { "type": "string", "title": "展示位置", "enum": ["sidebar", "footer"], "enumNames": ["侧边栏", "页脚"], "default": "sidebar" }
    }
  },
  "default_config": {
    "enabled": true,
    "title": "友情链接",
    "max_links": 10,
    "display_position": "sidebar"
  },
  "migrations": [
    {
      "plugin_code": "$code",
      "migration_version": "1.0.0",
      "migration_name": "official_links_init",
      "direction": "up",
      "checksum": "sha256:fixture",
      "tables": ["$table_name"],
      "rollback_supported": false,
      "description": "创建友情链接最小业务表；install 只基于 migrations/ 计划处理。"
    }
  ]
}
JSON
  cat > "$dir/migrations/001_init.sql" <<SQL
-- DevHub S15 declarative plugin fixture migration.
-- dry-run 不执行 SQL；install 只允许基于 migrations/ 计划处理。
CREATE TABLE IF NOT EXISTS $table_name (
  id BIGINT PRIMARY KEY,
  title VARCHAR(120) NOT NULL,
  url VARCHAR(500) NOT NULL,
  community_id BIGINT NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'enabled',
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL
);
SQL
  cat > "$dir/README.md" <<MD
# 声明型友情链接插件

DevHub S15 真实声明型插件 fixture，用于验证 upload -> precheck -> promote -> local repository -> install dry-run -> install -> enable -> community enable -> content_type / menu / permission / config 的完整业务闭环。

- 插件编码：$code
- 内容类型：$content_type
- 业务表：$table_name
- 不包含 package scripts
- 不包含远程代码
- 不包含动态加载入口
- 不包含远程 iframe URL
- 不包含根目录 001_schema.sql
- SQL 只放在 migrations/
MD
  cat > "$dir/config.example.json" <<JSON
{
  "enabled": true,
  "title": "友情链接",
  "max_links": 10,
  "display_position": "sidebar"
}
JSON
}

write_checksums() {
  local dir="$1"
  (
    cd "$dir"
    local first=1
    {
      echo '{'
      echo '  "algorithm": "sha256",'
      echo '  "files": ['
      while IFS= read -r file; do
        local sum
        sum="$(sha256sum "$file" | awk '{print $1}')"
        if [[ "$first" -eq 0 ]]; then
          echo ','
        fi
        first=0
        printf '    { "path": "%s", "sha256": "%s" }' "$file" "$sum"
      done < <(find . -type f ! -name checksums.json -printf '%P\n' | LC_ALL=C sort)
      echo
      echo '  ]'
      echo '}'
    } > checksums.json
  )
}

zip_dir() {
  local dir="$1"
  local out="$2"
  rm -f "$out"
  if command -v zip >/dev/null 2>&1; then
    (
      cd "$dir"
      zip -qr "$out" .
    )
    return
  fi
  node - "$dir" "$out" <<'NODE'
const fs = require('fs');
const path = require('path');
const root = process.argv[2];
const out = process.argv[3];

function files(dir, base = '') {
  const names = fs.readdirSync(dir).sort();
  const result = [];
  for (const name of names) {
    const full = path.join(dir, name);
    const rel = path.posix.join(base, name);
    const st = fs.statSync(full);
    if (st.isDirectory()) result.push(...files(full, rel));
    else if (st.isFile()) result.push({ full, rel });
  }
  return result;
}

const crcTable = Array.from({ length: 256 }, (_, n) => {
  let c = n;
  for (let k = 0; k < 8; k += 1) c = (c & 1) ? (0xedb88320 ^ (c >>> 1)) : (c >>> 1);
  return c >>> 0;
});

function crc32(buf) {
  let crc = -1;
  for (const byte of buf) crc = (crc >>> 8) ^ crcTable[(crc ^ byte) & 0xff];
  return (crc ^ -1) >>> 0;
}

const chunks = [];
const central = [];
let offset = 0;
const entries = files(root);
for (const entry of entries) {
  const name = Buffer.from(entry.rel);
  const data = fs.readFileSync(entry.full);
  const crc = crc32(data);
  const local = Buffer.alloc(30);
  local.writeUInt32LE(0x04034b50, 0);
  local.writeUInt16LE(20, 4);
  local.writeUInt16LE(0, 6);
  local.writeUInt16LE(0, 8);
  local.writeUInt16LE(0, 10);
  local.writeUInt16LE(0, 12);
  local.writeUInt32LE(crc, 14);
  local.writeUInt32LE(data.length, 18);
  local.writeUInt32LE(data.length, 22);
  local.writeUInt16LE(name.length, 26);
  local.writeUInt16LE(0, 28);
  chunks.push(local, name, data);

  const head = Buffer.alloc(46);
  head.writeUInt32LE(0x02014b50, 0);
  head.writeUInt16LE(20, 4);
  head.writeUInt16LE(20, 6);
  head.writeUInt16LE(0, 8);
  head.writeUInt16LE(0, 10);
  head.writeUInt16LE(0, 12);
  head.writeUInt16LE(0, 14);
  head.writeUInt32LE(crc, 16);
  head.writeUInt32LE(data.length, 20);
  head.writeUInt32LE(data.length, 24);
  head.writeUInt16LE(name.length, 28);
  head.writeUInt16LE(0, 30);
  head.writeUInt16LE(0, 32);
  head.writeUInt16LE(0, 34);
  head.writeUInt16LE(0, 36);
  head.writeUInt32LE(0, 38);
  head.writeUInt32LE(offset, 42);
  central.push(head, name);
  offset += local.length + name.length + data.length;
}

const centralStart = offset;
const centralSize = central.reduce((sum, item) => sum + item.length, 0);
const end = Buffer.alloc(22);
end.writeUInt32LE(0x06054b50, 0);
end.writeUInt16LE(0, 4);
end.writeUInt16LE(0, 6);
end.writeUInt16LE(entries.length, 8);
end.writeUInt16LE(entries.length, 10);
end.writeUInt32LE(centralSize, 12);
end.writeUInt32LE(centralStart, 16);
end.writeUInt16LE(0, 20);
fs.writeFileSync(out, Buffer.concat([...chunks, ...central, end]));
NODE
}

valid_code="fixture_valid_plugin${code_suffix}"
blocked_code="fixture_blocked_plugin${code_suffix}"
deprecated_code="fixture_deprecated_schema_plugin${code_suffix}"
links_code="official_links${code_suffix}"
links_content_type="friend_link${code_suffix}"
links_table="official_links_items${code_suffix}"

valid_dir="$SRC_DIR/devhub-fixture-valid-plugin${suffix_part}"
blocked_dir="$SRC_DIR/devhub-fixture-blocked-plugin${suffix_part}"
deprecated_dir="$SRC_DIR/devhub-fixture-deprecated-schema-plugin${suffix_part}"
links_dir="$SRC_DIR/devhub-fixture-links-plugin${suffix_part}"

write_manifest "$valid_dir" "$valid_code" "DevHub Fixture Valid Plugin"
write_checksums "$valid_dir"

write_manifest "$blocked_dir" "$blocked_code" "DevHub Fixture Blocked Plugin"
mkdir -p "$blocked_dir/scripts"
cat > "$blocked_dir/scripts/install.sh" <<'SH'
#!/usr/bin/env sh
echo "this script must never run"
SH
write_checksums "$blocked_dir"

write_manifest "$deprecated_dir" "$deprecated_code" "DevHub Fixture Deprecated Schema Plugin"
cat > "$deprecated_dir/001_schema.sql" <<SQL
-- Deprecated root schema fixture.
-- DevHub must warn about this file and must not execute it.
CREATE TABLE IF NOT EXISTS ${deprecated_code}_deprecated_root (
  id BIGINT PRIMARY KEY
);
SQL
write_checksums "$deprecated_dir"

write_links_manifest "$links_dir" "$links_code" "$links_content_type" "$links_table"
write_checksums "$links_dir"

zip_dir "$valid_dir" "$DIST_DIR/devhub-fixture-valid-plugin${suffix_part}.zip"
zip_dir "$blocked_dir" "$DIST_DIR/devhub-fixture-blocked-plugin${suffix_part}.zip"
zip_dir "$deprecated_dir" "$DIST_DIR/devhub-fixture-deprecated-schema-plugin${suffix_part}.zip"
zip_dir "$links_dir" "$DIST_DIR/devhub-fixture-links-plugin${suffix_part}.zip"

cat > "$DIST_DIR/manifest${suffix_part}.json" <<JSON
{
  "valid": {
    "plugin_code": "$valid_code",
    "zip": "devhub-fixture-valid-plugin${suffix_part}.zip"
  },
  "blocked": {
    "plugin_code": "$blocked_code",
    "zip": "devhub-fixture-blocked-plugin${suffix_part}.zip"
  },
  "deprecated": {
    "plugin_code": "$deprecated_code",
    "zip": "devhub-fixture-deprecated-schema-plugin${suffix_part}.zip"
  },
  "links": {
    "plugin_code": "$links_code",
    "content_type": "$links_content_type",
    "table": "$links_table",
    "zip": "devhub-fixture-links-plugin${suffix_part}.zip"
  }
}
JSON

echo "已生成插件包 fixture：$DIST_DIR"

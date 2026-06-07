#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

if git -C "${REPO_ROOT}" rev-parse --show-toplevel >/dev/null 2>&1; then
  REPO_ROOT="$(git -C "${REPO_ROOT}" rev-parse --show-toplevel)"
fi

cd "${REPO_ROOT}"

if ! command -v python3 >/dev/null 2>&1; then
  echo "缺少 python3 命令，无法执行一键 webhook 链路脚本。" >&2
  exit 1
fi

python3 <<'PY'
import json
import os
import sys
import threading
import time
from pathlib import Path
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from urllib.error import HTTPError, URLError
from urllib.parse import quote
from urllib.request import Request, urlopen


def trim_right(value, char):
    out = str(value or "")
    while out.endswith(char):
        out = out[:-1]
    return out


env = os.environ
origin = trim_right(env.get("DEVHUB_ORIGIN") or env.get("DEVHUB_E2E_ORIGIN") or "http://127.0.0.1:8090", "/")
admin_account = env.get("DEVHUB_ADMIN_ACCOUNT", "admin")
admin_password = env.get("DEVHUB_ADMIN_PASSWORD", "admin123")
plugin_code = env.get("DEVHUB_PLUGIN_CODE", "plugin_a7b0cc04")
plugin_name = env.get("DEVHUB_PLUGIN_NAME", "飞书链接插件")
content_type = env.get("DEVHUB_CONTENT_TYPE", "feishu_link")
publish_content_type = env.get("DEVHUB_PUBLISH_CONTENT_TYPE", content_type)
publish_plugin_code = env.get("DEVHUB_PUBLISH_PLUGIN_CODE", plugin_code)
category_plugin_code = env.get("DEVHUB_CATEGORY_PLUGIN_CODE", plugin_code)
package_path = env.get("DEVHUB_PACKAGE_PATH", "storage/plugins/packages/plugin_a7b0cc04")
manifest_path = Path(package_path) / "manifest.json"
community_slug = env.get("DEVHUB_COMMUNITY_SLUG", "php")
category_slug = env.get("DEVHUB_CATEGORY_SLUG", "feishu-link")
category_name = env.get("DEVHUB_CATEGORY_NAME", "飞书链接")
category_id_override = env.get("DEVHUB_CATEGORY_ID", "").strip()
webhook_port = int(env.get("DEVHUB_WEBHOOK_PORT", "18081"))
webhook_host = env.get("DEVHUB_WEBHOOK_HOST", "0.0.0.0")
webhook_endpoint = trim_right(env.get("DEVHUB_WEBHOOK_ENDPOINT") or f"http://127.0.0.1:{webhook_port}", "/")
start_receiver = env.get("DEVHUB_START_RECEIVER", "1") != "0"
create_category = env.get("DEVHUB_CREATE_CATEGORY", "1") != "0"
webhook_mode = env.get("DEVHUB_WEBHOOK_MODE", "ok")
webhook_flow = env.get("DEVHUB_WEBHOOK_FLOW", "success").strip().lower() or "success"
title_prefix = env.get("DEVHUB_TOPIC_TITLE_PREFIX", "飞书链接 webhook 验证")
webhook_token = env.get("DEVHUB_WEBHOOK_TOKEN", "")

received_count = 0
last_received = None
receiver = None


def log(message):
    print(f"[feishu-webhook] {message}", flush=True)


def normalize_items(payload):
    if isinstance(payload, list):
        return payload
    if isinstance(payload, dict):
        if isinstance(payload.get("items"), list):
            return payload["items"]
        data = payload.get("data")
        if isinstance(data, dict) and isinstance(data.get("items"), list):
            return data["items"]
    return []


def api_request(method, path, token="", body=None, allow_failure=False):
    raw = None if body is None else json.dumps(body, ensure_ascii=False).encode("utf-8")
    headers = {"Content-Type": "application/json"}
    if token:
        headers["Authorization"] = f"Bearer {token}"
    req = Request(f"{origin}{path}", data=raw, headers=headers, method=method)
    try:
        with urlopen(req, timeout=20) as resp:
            text = resp.read().decode("utf-8")
            data = parse_json(text)
            return {"ok": 200 <= resp.status < 300, "status": resp.status, "data": data}
    except HTTPError as err:
        text = err.read().decode("utf-8", errors="replace")
        data = parse_json(text)
        if allow_failure:
            return {"ok": False, "status": err.code, "data": data}
        raise RuntimeError(f"{method} {path} 失败：{err.code} {error_message(data, err.reason)}") from None
    except URLError as err:
        raise RuntimeError(f"{method} {path} 连接失败：{err.reason}") from None


def must(method, path, token="", body=None):
    return api_request(method, path, token, body)["data"]


def parse_json(text):
    if not text:
        return {}
    try:
        return json.loads(text)
    except json.JSONDecodeError:
        return {"raw": text}


def error_message(data, fallback=""):
    if isinstance(data, dict):
        return data.get("message") or data.get("error") or data.get("raw") or str(fallback)
    return str(fallback)


def assert_no_sensitive(label, payload):
    text = json.dumps(payload, ensure_ascii=False, sort_keys=True)
    lowered = text.lower()
    needles = ['"authorization"', "encrypted_value", "token_hash"]
    if webhook_token and len(webhook_token) >= 6:
        needles.append(webhook_token.lower())
    for needle in needles:
        if needle in lowered:
            raise RuntimeError(f"{label} 出现敏感字段或明文：{needle}")


def set_receiver_mode(mode):
    global webhook_mode
    webhook_mode = mode
    log(f"receiver 模式：{mode}")


class ReceiverHandler(BaseHTTPRequestHandler):
    def log_message(self, _fmt, *_args):
        return

    def do_GET(self):
        if self.path.split("?", 1)[0] == "/health":
            self.write_json(200, {"ok": True, "message": "ok"})
            return
        self.write_json(404, {"ok": False, "error": "not_found"})

    def do_POST(self):
        global received_count, last_received
        length = int(self.headers.get("Content-Length", "0") or "0")
        raw = self.rfile.read(length).decode("utf-8", errors="replace")
        body = parse_json(raw)
        if self.path.split("?", 1)[0] != "/hooks/content.after_create":
            self.write_json(404, {"ok": False, "error": "not_found"})
            return
        received_count += 1
        last_received = {
            "method": self.command,
            "url": self.path,
            "authorization_present": bool(self.headers.get("Authorization")),
            "body": body,
            "received_at": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
        }
        data = body.get("data") if isinstance(body, dict) else {}
        topic_id = data.get("topic_id") if isinstance(data, dict) else None
        log(f"receiver 收到 webhook #{received_count}: {body.get('hook_name', '-') if isinstance(body, dict) else '-'} topic={topic_id or body.get('resource_id', '-') if isinstance(body, dict) else '-'}")
        if webhook_mode == "500":
            self.write_json(500, {"ok": False, "error": "mock_500"})
            return
        if webhook_mode == "429":
            self.send_response(429)
            self.send_header("Retry-After", "1")
            self.send_header("Content-Type", "application/json; charset=utf-8")
            self.end_headers()
            self.wfile.write(json.dumps({"ok": False, "error": "mock_429"}).encode("utf-8"))
            return
        if webhook_mode == "timeout":
            time.sleep(float(env.get("DEVHUB_RECEIVER_TIMEOUT_SLEEP", "4.2")))
            self.write_json(200, {"ok": True, "message": "late"})
            return
        self.write_json(200, {"ok": True, "message": "accepted"})

    def write_json(self, status, payload):
        raw = json.dumps(payload, ensure_ascii=False).encode("utf-8")
        try:
            self.send_response(status)
            self.send_header("Content-Type", "application/json; charset=utf-8")
            self.send_header("Content-Length", str(len(raw)))
            self.end_headers()
            self.wfile.write(raw)
        except BrokenPipeError:
            return


def start_webhook_receiver():
    global receiver
    if not start_receiver:
        return
    receiver = ThreadingHTTPServer((webhook_host, webhook_port), ReceiverHandler)
    thread = threading.Thread(target=receiver.serve_forever, daemon=True)
    thread.start()
    log(f"临时 receiver 已监听 {webhook_host}:{webhook_port}")


def stop_webhook_receiver():
    if receiver:
        receiver.shutdown()
        receiver.server_close()


def q(value):
    return quote(str(value), safe="")


def login_admin():
    session = must("POST", "/api/v1/admin/login", body={"account": admin_account, "password": admin_password})
    token = session.get("access_token") or session.get("accessToken") or session.get("token")
    if not token:
        raise RuntimeError("后台登录未返回 access token")
    user = session.get("user") or {}
    log(f"后台登录成功：{user.get('username') or admin_account}")
    return token


def load_package_manifest():
    try:
        return json.loads(manifest_path.read_text(encoding="utf-8"))
    except FileNotFoundError:
        return None


def version_tuple(value):
    out = []
    for part in str(value or "").split("."):
        digits = "".join(ch for ch in part if ch.isdigit())
        out.append(int(digits or "0"))
    return tuple(out + [0] * (3 - len(out)))


def plugin_needs_upgrade(plugin, manifest):
    if not plugin or not manifest:
        return False
    if version_tuple(manifest.get("version")) > version_tuple(plugin.get("version")):
        return True
    if plugin_code != "plugin_a7b0cc04" or content_type != "feishu_link":
        return False
    menus = plugin.get("menus") or []
    content_types = plugin.get("content_types") or []
    return "feishu_link" not in content_types or not any(item.get("path") == "/admin-next/feishu-links" for item in menus)


def get_plugin(token):
    payload = must("GET", "/api/v1/admin/plugins", token=token)
    for item in normalize_items(payload):
        if item.get("code") == plugin_code or item.get("plugin_code") == plugin_code:
            return item
    return None


def ensure_plugin(token):
    manifest = load_package_manifest()
    plugin = get_plugin(token)
    if not plugin:
        log(f"未找到 {plugin_code}，从 {package_path} dry-run/install")
        dry = must("POST", "/api/v1/admin/plugins/packages/dry-run", token=token, body={"path": package_path})
        dry_run_id = dry.get("dry_run_id")
        if not dry_run_id:
            raise RuntimeError("插件包 dry-run 未返回 dry_run_id")
        risk = str(((dry.get("risk_report") or {}).get("level")) or dry.get("status") or "").lower()
        install_payload = {"path": package_path, "dry_run_id": dry_run_id}
        if risk and risk not in ("low", "passed"):
            install_payload["confirm_risk_level"] = risk
        must("POST", "/api/v1/admin/plugins/packages/install", token=token, body=install_payload)
        plugin = get_plugin(token)
    elif plugin_needs_upgrade(plugin, manifest):
        log(f"{plugin_code} 已存在，升级到本地 manifest {manifest.get('version') if manifest else ''}")
        must("POST", f"/api/v1/admin/plugins/{q(plugin_code)}/upgrade", token=token, body={"manifest": manifest, "confirm": True})
        plugin = get_plugin(token)
    if not plugin:
        raise RuntimeError(f"{plugin_code} 安装后仍未出现在插件列表")
    if plugin.get("status") == "archived":
        log(f"{plugin_code} 当前已归档，执行 restore")
        must("POST", f"/api/v1/admin/plugins/{q(plugin_code)}/restore", token=token, body={})
        plugin = get_plugin(token)
    if plugin.get("status") != "enabled":
        log(f"{plugin_code} 当前状态 {plugin.get('status') or 'unknown'}，执行 enable")
        must("POST", f"/api/v1/admin/plugins/{q(plugin_code)}/enable", token=token, body={})
        plugin = get_plugin(token)
    log(f"插件已就绪：{plugin_name} ({plugin_code}) status={plugin.get('status') or 'enabled'}")
    return plugin


def ensure_community(token):
    payload = must("GET", "/api/v1/admin/communities", token=token)
    communities = normalize_items(payload)
    community = next((item for item in communities if item.get("slug") == community_slug), None)
    if not community:
        community = next((item for item in communities if int(item.get("status") or 0) == 1), None)
    if not community and communities:
        community = communities[0]
    if not community or not community.get("id"):
        raise RuntimeError("未找到可用子站")
    resp = api_request("POST", f"/api/v1/admin/communities/{community['id']}/plugins/{q(plugin_code)}/enable", token=token, body={}, allow_failure=True)
    if not resp["ok"]:
        raise RuntimeError(f"启用子站插件失败：{error_message(resp['data'])}")
    log(f"子站已就绪：{community.get('slug') or community['id']}")
    return community


def ensure_category(token, community):
    if category_id_override:
        log(f"使用指定板块：id={category_id_override}")
        return {"id": int(category_id_override), "name": category_name, "content_type": publish_content_type}
    path = f"/api/v1/admin/communities/{community['id']}/categories"
    categories = normalize_items(must("GET", path, token=token))
    category = None
    for item in categories:
        allowed = item.get("allowed_content_types") if isinstance(item.get("allowed_content_types"), list) else []
        if item.get("plugin_code") == category_plugin_code or item.get("content_type") == publish_content_type or publish_content_type in allowed:
            category = item
            break
    if not category:
        if not create_category:
            raise RuntimeError(f"未找到 {content_type} 板块，且 DEVHUB_CREATE_CATEGORY=0")
        log(f"未找到 {content_type} 板块，创建 {category_name}")
        category = must("POST", path, token=token, body={
            "name": category_name,
            "slug": category_slug,
            "type": publish_content_type,
            "content_type": publish_content_type,
            "plugin_code": category_plugin_code,
            "allowed_content_types": [publish_content_type],
            "description": "用于一键验证飞书链接 external_service webhook。",
            "icon": "link",
            "sort_order": 320,
            "visible": True,
            "nav_visible": True,
            "postable": True,
            "status": 1,
        })
    else:
        allowed = set(category.get("allowed_content_types") or [])
        allowed.add(publish_content_type)
        needs_update = (
            category.get("plugin_code") != category_plugin_code
            or category.get("content_type") != publish_content_type
            or not category.get("postable")
            or int(category.get("status") or 0) != 1
            or publish_content_type not in allowed
        )
        if needs_update:
            log(f"更新板块 {category.get('id')} 以绑定 {publish_content_type}")
            body = dict(category)
            body.update({
                "type": publish_content_type,
                "content_type": publish_content_type,
                "plugin_code": category_plugin_code,
                "allowed_content_types": sorted(allowed),
                "visible": category.get("visible") is not False,
                "nav_visible": category.get("nav_visible") is not False,
                "postable": True,
                "status": 1,
            })
            category = must("PUT", f"/api/v1/admin/categories/{category['id']}", token=token, body=body)
    if int(category.get("status") or 0) != 1:
        must("POST", f"/api/v1/admin/categories/{category['id']}/enable", token=token, body={})
    log(f"板块已就绪：{category.get('name') or category.get('id')} id={category.get('id')}")
    return category


def configure_external_service(token):
    payload = {
        "enabled": True,
        "endpoint_url": webhook_endpoint,
        "health_check_path": "/health",
        "timeout_ms": int(env.get("DEVHUB_WEBHOOK_TIMEOUT_MS", "3000")),
        "failure_policy": env.get("DEVHUB_WEBHOOK_FAILURE_POLICY", "warn"),
        "auth_type": env.get("DEVHUB_WEBHOOK_AUTH_TYPE", "none"),
        "token": webhook_token,
        "warning_threshold": int(env.get("DEVHUB_WEBHOOK_WARNING_THRESHOLD", "1")),
        "error_threshold": int(env.get("DEVHUB_WEBHOOK_ERROR_THRESHOLD", "3")),
    }
    saved = must("PUT", f"/api/v1/admin/plugins/{q(plugin_code)}/external-service", token=token, body=payload)
    assert_no_sensitive("external_service 保存响应", saved)
    if payload["auth_type"].lower() == "bearer":
        token_ref = saved.get("token_ref") or ((saved.get("token_secret") or {}).get("ref"))
        expected_ref = f"secret://external_service/{plugin_code}/token"
        if token_ref != expected_ref:
            raise RuntimeError(f"bearer token 未收口为 token_ref，got={token_ref!r} expected={expected_ref!r}")
    log(f"external_service 已配置：{webhook_endpoint}")
    health = api_request("POST", f"/api/v1/admin/plugins/{q(plugin_code)}/external-service/health-check", token=token, body={}, allow_failure=True)
    if not health["ok"]:
        raise RuntimeError(f"external_service health-check 失败：{error_message(health['data'])}。如果后端在 Docker 内，请设置 DEVHUB_WEBHOOK_ENDPOINT=http://172.17.0.1:{webhook_port}")
    data = health["data"]
    assert_no_sensitive("external_service health-check 响应", data)
    log(f"health-check 通过：{data.get('status') or data.get('health_status') or 'ok'}")


def publish_feishu_link(token, community, category):
    title = f"{title_prefix} {int(time.time() * 1000)}"
    payload = {
        "site": community.get("slug"),
        "board": "community",
        "category_id": category.get("id"),
        "content_type": publish_content_type,
        "plugin_code": publish_plugin_code,
        "title": title,
        "summary": "一键脚本发布，用于触发 AfterCreateContent external_service webhook。",
        "content": f"飞书链接 webhook 验证正文。\n\nhttps://example.com/feishu/{int(time.time() * 1000)}",
        "status": "published",
    }
    result = must("POST", "/api/v1/admin/posts", token=token, body=payload)
    topic = result.get("topic") if isinstance(result.get("topic"), dict) else result
    topic_id = topic.get("id") or result.get("id") or result.get("topic_id")
    if not topic_id:
        raise RuntimeError(f"发布成功但未返回 topic id：{json.dumps(result, ensure_ascii=False)}")
    log(f"内容发布成功：topic_id={topic_id} title={title}")
    return topic_id


def list_executions(token, topic_id=None):
    path = f"/api/v1/admin/plugins/{q(plugin_code)}/hooks/executions?service_type=external_service&page=1&page_size=20"
    if topic_id:
        path += f"&content_id={q(topic_id)}"
    payload = must("GET", path, token=token)
    assert_no_sensitive("hook_executions 响应", payload)
    return normalize_items(payload)


def find_execution(token, topic_id, expected_statuses=None, require_error_code=""):
    expected_statuses = set(expected_statuses or [])
    for _ in range(12):
        rows = list_executions(token, topic_id)
        matched = None
        for row in rows:
            if int(row.get("content_id") or 0) == int(topic_id) and row.get("hook_name") == "AfterCreateContent":
                status = str(row.get("status") or "").lower()
                error_code = str(row.get("error_code") or "").lower()
                if expected_statuses and status not in expected_statuses:
                    matched = matched or row
                    continue
                if require_error_code and error_code != require_error_code:
                    matched = matched or row
                    continue
                return row
                break
        if matched:
            log(f"投递仍在进行：status={matched.get('status')}")
        time.sleep(1)
    return None


def run_delivery_scenario(token, community, category, name, mode, expected_statuses, require_error_code=""):
    set_receiver_mode(mode)
    topic_id = publish_feishu_link(token, community, category)
    execution = find_execution(token, topic_id, expected_statuses, require_error_code)
    if not execution:
        raise RuntimeError(f"{name} 未找到期望投递记录：topic_id={topic_id} expected={sorted(expected_statuses)}")
    status = execution.get("status") or ("success" if execution.get("success") else "failed")
    log(f"{name} 投递记录：id={execution.get('id')} status={status} response={execution.get('response_status') or '-'} error={execution.get('error_code') or '-'}")
    return execution


def run_manual_retry_scenario(token, community, category):
    failed = run_delivery_scenario(token, community, category, "manual-retry-source", "500", {"retry_exhausted"})
    set_receiver_mode("ok")
    resp = must("POST", f"/api/v1/admin/plugins/{q(plugin_code)}/hooks/executions/{failed.get('id')}/retry", token=token, body={})
    assert_no_sensitive("manual retry 响应", resp)
    if resp.get("status") != "success":
        raise RuntimeError(f"manual retry 未成功：{json.dumps(resp, ensure_ascii=False)}")
    retry_id = resp.get("retry_record_id")
    rows = list_executions(token)
    retry_row = next((row for row in rows if str(row.get("id")) == str(retry_id)), None)
    if not retry_row or retry_row.get("status") != "success":
        raise RuntimeError(f"manual retry 执行记录未成功：retry_record_id={retry_id}")
    log(f"manual retry 成功：source={failed.get('id')} retry={retry_id}")
    return retry_row


def verify_admin_logs(token):
    payload = must("GET", f"/api/v1/admin/audit-logs?plugin_code={q(plugin_code)}&page=1&page_size=50", token=token)
    assert_no_sensitive("admin_logs 响应", payload)
    rows = normalize_items(payload)
    if not rows:
        raise RuntimeError(f"admin_logs 未找到 plugin_code={plugin_code} 的审计记录")
    log(f"admin_logs 已检查：{len(rows)} 条，未发现敏感明文")


def main():
    try:
        start_webhook_receiver()
        log(f"DevHub origin: {origin}")
        admin_token = login_admin()
        ensure_plugin(admin_token)
        community = ensure_community(admin_token)
        category = ensure_category(admin_token, community)
        configure_external_service(admin_token)
        if webhook_flow == "full":
            run_delivery_scenario(admin_token, community, category, "success", "ok", {"success"})
            run_delivery_scenario(admin_token, community, category, "fail-500", "500", {"retry_exhausted"})
            run_delivery_scenario(admin_token, community, category, "timeout", "timeout", {"retry_exhausted"}, "network_timeout")
            run_manual_retry_scenario(admin_token, community, category)
        elif webhook_flow == "fail":
            run_delivery_scenario(admin_token, community, category, "fail-500", "500", {"retry_exhausted"})
        elif webhook_flow == "timeout":
            run_delivery_scenario(admin_token, community, category, "timeout", "timeout", {"retry_exhausted"}, "network_timeout")
        elif webhook_flow == "retry":
            run_manual_retry_scenario(admin_token, community, category)
        else:
            run_delivery_scenario(admin_token, community, category, "success", webhook_mode, {"success"})
        verify_admin_logs(admin_token)
        if start_receiver:
            log(f"receiver 收到次数：{received_count}")
            if last_received:
                log(f"receiver authorization_present={last_received.get('authorization_present')}")
                log(f"receiver 最后 payload：{json.dumps(last_received['body'], ensure_ascii=False)}")
        log("完成。")
    finally:
        stop_webhook_receiver()


if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        stop_webhook_receiver()
        raise
    except Exception as exc:
        stop_webhook_receiver()
        print(f"[feishu-webhook] 失败：{exc}", file=sys.stderr, flush=True)
        sys.exit(1)
PY

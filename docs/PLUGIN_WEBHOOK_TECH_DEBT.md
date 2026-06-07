# external_service Webhook 技术债记录

[返回文档入口](README.md)

更新时间：2026-05-28

本文记录本地联调 `plugin_a7b0cc04` external_service Webhook 时暴露出的技术债。它不是用户使用手册；面向后续产品、后端、前端和运维收口。

## 本次现场问题

本次联调围绕 `AfterCreateContent -> external_service -> /hooks/content.after_create`，连续出现以下问题：

1. DevHub 后端运行在 Docker 容器中，`endpoint_url=http://127.0.0.1:18081` 会让容器访问自身 loopback，导致：

```text
Post "http://127.0.0.1:18081/hooks/content.after_create": dial tcp 127.0.0.1:18081: connect: connection refused
```

2. 将 endpoint 改为 Docker host gateway `http://172.17.0.1:18081` 后，网络可达，但后端安全校验拒绝非 HTTPS、非 localhost HTTP endpoint：

```text
plugin_external_service_invalid_endpoint
生产环境建议使用 https；仅本地开发允许 http://localhost 或 127.0.0.1
```

3. 用户在插件“全局配置”里保存了 `endpoint_url=http://172.17.0.1:18081`，但投递仍使用 `http://127.0.0.1:18081`。根因是全局 `config_json` 与运行态 `external_service_config` 是两套配置，健康检查和投递只读取 `external_service_config`。

4. 插件包重复安装被阻断：

```text
plugin_package_already_installed
同编码插件已安装，不能重复安装
```

错误本身合理，但页面入口让用户误以为需要“重新安装”才能更新配置或升级。

5. 版本仓库只有“可对比 / 查看版本 / 升级差异 / 提交审批 / 审批执行”的链路，没有直观的“升级”入口，用户难以从错误建议跳转到正确流程。

6. external_service 已经恢复 healthy 后，插件详情顶部仍可能展示历史 Hook 失败摘要，例如旧的 `127.0.0.1` 失败记录，造成“当前配置还是 127”的误解。

## 已做缓解

- 新增 `DEVHUB_EXTERNAL_SERVICE_HTTP_ALLOWLIST`：默认仍要求 HTTPS 或 localhost；只有显式写入 allowlist 的 HTTP origin 才放行。
- `dev.sh` 透传 `DEVHUB_EXTERNAL_SERVICE_HTTP_ALLOWLIST` 到本地 Go 和 Docker Go 进程。
- external_service endpoint 校验失败提示补充 allowlist 建议。
- 插件健康摘要在 external_service 后续 healthy 成功后，不再用已恢复的 non-blocking Hook 历史失败覆盖当前“最近异常”。

本地联调推荐启动方式：

```bash
DEVHUB_EXTERNAL_SERVICE_HTTP_ALLOWLIST=http://172.17.0.1:18081 ./dev.sh restart --no-build
```

运行态配置应写入：

```text
插件详情 -> 运行记录 -> external_service 配置
endpoint_url = http://172.17.0.1:18081
health_check_path = /health
```

不要把 external_service endpoint 只写入插件全局配置。

## 技术债清单

| ID | 优先级 | 问题 | 影响 | 建议 |
| --- | --- | --- | --- | --- |
| WH-01 | P0 | Docker 本地联调下 `127.0.0.1` 语义容易误导 | Webhook 接收端明明在宿主机监听，容器内仍连接失败 | 后台在检测到 Docker / dev 环境时提示 `host.docker.internal` 或 allowlist；一键脚本继续自动给出正确 endpoint |
| WH-02 | P0 | 全局配置与 external_service 运行配置字段重名 | 用户保存了看似正确的 `endpoint_url`，但投递仍读旧运行配置 | 全局配置页对 external_service 字段加醒目标识，或提供“同步到 external_service 运行配置”的明确入口 |
| WH-03 | P0 | endpoint 安全策略缺少可治理的白名单 UI | 当前只能靠环境变量，运维可见性弱 | 后台系统设置增加 external_service HTTP allowlist 只读展示或受控编辑，审计记录变更 |
| WH-04 | P1 | 插件重复安装错误建议不够可操作 | 用户看到“请使用升级流程”，但页面没有直接“升级”按钮 | 错误响应补充具体入口；前端在本地包已安装时给出“查看版本 / 升级差异 / 配置 external_service”快捷操作 |
| WH-05 | P1 | 版本升级链路隐藏在“可对比”后 | 用户无法快速理解“对比 -> 提交审批 -> 审批 -> 执行” | 将 `可对比` 改为操作按钮文案；详情抽屉顶部展示下一步 CTA |
| WH-06 | P1 | 健康摘要和历史执行记录语义混在一起 | 已恢复服务仍显示旧失败，误判当前配置 | 当前已缓解；后续应在 UI 上区分“当前健康”与“历史失败”，并提供时间戳 |
| WH-07 | P1 | external_service 配置入口太深 | 用户容易在全局配置页误操作 | 插件详情概览卡片增加“配置 external_service”直达按钮；Webhook 治理也提供配置入口 |
| WH-08 | P2 | 本地 receiver 工具与实际端口不统一 | `cmd/webhook-mock-receiver` 默认 18090，一键脚本 / 现场服务用 18081 | 文档和脚本统一默认端口，或在 UI / 报错里显示当前 receiver 推荐命令 |
| WH-09 | P2 | allowlist 仅按 origin 匹配，缺少解释型诊断 | 配错端口 / host 时只看到 endpoint invalid 或连接失败 | 健康检查失败详情增加“安全校验失败 / 网络连接失败 / HTTP 状态失败”的分层提示 |

## 建议收口路线

1. P0 先收口配置混淆：全局配置页不要让 `endpoint_url` 看起来像投递运行配置；external_service 配置入口要更直达。
2. P0 保留安全默认值：生产仍默认 HTTPS；HTTP 只允许 localhost 或显式 allowlist。
3. P1 改善升级入口：把“同编码插件已安装”错误直接引导到版本与升级的对应插件行。
4. P1 拆分健康摘要：当前健康、最近成功、最近失败、历史失败计数分别展示，不互相覆盖。
5. P2 统一本地联调工具：receiver 默认端口、脚本提示、文档示例和健康检查推荐命令保持一致。

## 手动验证建议

按仓库规则，自动测试由人工按需执行。建议验证：

```bash
docker run --rm -v "$PWD":/app -w /app golang:1.22-bookworm go test ./internal/service -run TestExternalServiceHTTPAllowlist -count=1
./scripts/test-all.sh
```


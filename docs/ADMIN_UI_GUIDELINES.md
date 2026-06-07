# DevHub 后台 UI 基础设计规范

[返回文档入口](README.md)

更新时间：2026-05-29

本文记录 `v1.8.4-S20` 后台治理型控制台的基础视觉约定。当前目标是统一插件中心、SecretCenter、Webhook 治理和当前生效配置等高频后台页面，不改变任何 API、权限或插件生命周期语义。

## 基础原则

- 后台是治理型控制台，优先清楚、稳定、低噪音。
- 页面按“标题说明 -> 关键状态 -> 筛选操作 -> 表格列表 -> 详情抽屉 -> 技术详情”组织。
- 普通管理员先看结论；manifest、diagnostics、dry-run、raw JSON 等技术信息默认折叠。
- 危险操作必须用 warning / danger 样式并保留确认流程。
- token、secret、Authorization、root key、`encrypted_value` 不在页面、技术详情、日志说明中回显。
- 继续使用 Element Plus，不引入新的大型 UI 框架。

## Design Tokens

后台基础变量位于：

- `web/admin-app/src/styles/admin-tokens.css`
- `web/admin-app/src/styles/admin-layout.css`
- `web/admin-app/src/styles/admin-components.css`

变量覆盖后台背景、面板、边框、主/次文字、主色、成功/警告/危险/信息色、圆角、卡片阴影、页面宽度、页面间距、表格密度、标签高度和抽屉宽度。现有 `styles.css` 仍保留，新增样式只做后台治理页统一收口。

## 通用组件

后台通用组件位于 `web/admin-app/src/components/admin/`：

- `AdminPageHeader`
- `AdminSectionCard`
- `AdminMetricCard`
- `AdminStatusTag`
- `AdminRiskTag`
- `AdminActionBar`
- `AdminEmptyState`
- `AdminDetailDrawer`
- `AdminTechnicalDetails`
- `AdminInlineHint`

`AdminStatusTag` 和 `AdminRiskTag` 复用插件模块现有中文映射，不新增业务枚举。`AdminTechnicalDetails` 默认折叠，并对 `token`、`secret`、`Authorization`、`encrypted_value`、密文和 hash 类字段做前端兜底脱敏。

## 已接入页面

本轮已轻量接入：

- 插件列表页：统一页头、关键指标卡、状态标签。
- 插件详情抽屉：新增结论优先指标区，继续保留旧 `data-testid` 与技术详情默认折叠。
- Webhook 治理页：统一页头和总览指标，当前投递、待重试、熔断、历史 external_service 失败分开展示。
- 系统设置 / 当前生效配置：统一页头、状态指标、脱敏诊断折叠技术详情。
- SecretCenter：列表继续不展示明文，详情抽屉增加统一状态标签和折叠技术详情。
- 插件包治理入口、本地包与预检入口、版本仓库入口：统一页面标题区。

## 边界

本轮只改后台 UI 和文档：

- 不改变 API 语义。
- 不改变插件生命周期。
- 不改变 external_service 投递逻辑。
- 不改变 SecretCenter 安全模型。
- 不开放 blocking Hook。
- 不执行第三方代码。
- 不引入新大型 UI 框架。
- 不回显 token / secret / Authorization / `encrypted_value`。

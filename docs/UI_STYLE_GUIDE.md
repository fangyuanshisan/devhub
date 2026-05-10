# DevHub UI Style Guide

[返回文档大纲](README.md)

更新时间：2026-05-10

## 目的

本文档定义 DevHub 当前前台默认主题、样式 token 使用方式，以及文案 / 数据来源边界。

## 默认主题

- 主题名：`default-blue`
- 主题定位：蓝白现代社区风
- 主题角色：默认主题，不是唯一主题

当前默认主题强调：

- 清爽克制的蓝白界面
- 明确的信息层级
- 适度卡片化
- 面向社区内容浏览与发布

## Token 规则

前台样式 token 统一放在：

- `web/frontend-app/src/styles/tokens.css`

基础全局样式放在：

- `web/frontend-app/src/styles/global.css`

应优先使用 CSS Variables 管理以下能力：

- 颜色
- 圆角
- 阴影
- 字体
- 间距
- 内容宽度

当前 token 分层建议：

- 语义 token：`--color-bg-page`、`--color-brand-primary`、`--color-text-primary`
- 兼容别名：`--bg`、`--primary`、`--text`
- 结构 token：`--layout-content-width`
- 视觉 token：`--radius-*`、`--shadow-*`

## 组件实现规则

- 不在基础组件中硬编码 DevHub 专属颜色。
- 不在基础组件中硬编码 DevHub / LearnKu / PHP / Go / Java / AI 等专属文案。
- 组件内的颜色、圆角、阴影、字体、间距优先读取 token。
- 需要主题色时，优先读取站点配置或 CSS Variables，而不是写死 hex 值。
- 页面结构允许存在默认模块，但应为后续配置化预留替换空间。

## 内容来源边界

### 来自站点配置或默认配置

- 站点名称
- Logo / 缩写
- Slogan / 描述
- 顶部导航项
- 首页模块标题和顺序
- 默认主题色

### 来自业务数据

- 子站列表
- 板块 / 分类
- 标签
- Topic 列表和详情
- 搜索结果
- 公告
- 版主和用户数据
- 通知数和互动状态

### 来自组件内部

- 加载态
- 空状态结构
- 按钮状态
- 交互反馈

说明：组件可以提供默认兜底文案，但不应把品牌和技术方向写死进通用结构。

## 页面与组件分层

### 基础组件

- Button
- Card
- Badge
- Tabs
- Avatar
- Input
- Pagination
- EmptyState

### 社区业务组件

- SiteHeader
- SiteSwitcher
- ContentCard
- ContentFeed
- TagCloud
- BoardList
- AnnouncementCard
- ActiveUserList

### 页面模板

- 首页
- 子站页
- 搜索页
- 内容详情页
- 用户中心页

## 当前仓库约束

- 默认蓝白风继续保留，不做激进改版。
- 主题化仅整理前台结构，不改变数据库结构。
- 不新增复杂后台主题配置。
- 不破坏 Go 动态 SEO 页面输出。
- 不把具体技术内容写死到基础组件。

## 当前落地状态

已落地：

- `tokens.css` 已作为前台主题变量集中入口。
- 前台多个页面已改为共享 `sites.ts` 和 `site-config.ts`，减少重复站点定义。
- 首页、搜索页、子站页、发布页已开始去除部分品牌和技术硬编码。

仍需继续配置化：

- 首页模块顺序与显隐
- Header 顶部导航来源
- 子站专题文案与规则文案
- 页面级空状态和运营提示文案
- 后台管理端的主题 token 分层

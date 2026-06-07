 # DevHub 通用社区程序迭代执行文档

> 用途：本文件用于交给 Codex / AI Agent 执行 DevHub 项目迭代。
>
> 项目名称必须始终保持为 **DevHub**，不得改成 LearnKu、LearnKu Clone、Laravel China 或其他名称。
>
> 本次迭代目标不是做一个单一论坛，而是把 DevHub 升级为一个支持多技术子站的通用社区程序。

---

## 0. 当前执行状态，2026-05-09 更新

> 当前真实项目进度以 [docs/PROJECT_PROGRESS.md](docs/PROJECT_PROGRESS.md) 为准；全部文档导航见 [docs/README.md](docs/README.md)。

### 0.1 当前结论

DevHub 已经从“静态社区首页”推进到“可运行的多子站社区 CMS 骨架”阶段。

当前技术结构已经确定为：

```text
后端：Go + Gin
前台：Astro + Vue Islands
后台：Vue 3 + Element Plus
数据库：MemoryStore / MySQLStore
```

当前只维护两个正式入口：

```text
前台：http://127.0.0.1:8090/
后台：http://127.0.0.1:8090/admin-next
```

旧的原生 HTML / CSS / JS 前台和后台已经移除；`/admin`、`/admin/:site` 仅作为兼容重定向到 `/admin-next`。

### 0.2 已完成

```text
项目结构：
  - 已确认 Go + Gin / Astro / Vue 3 / MySQL 8 技术栈。
  - 已按 internal/domain、store、service、transport/httpapi 拆分后端。
  - 已保留 MemoryStore 与 MySQLStore 两套仓储。
  - 已移除旧原生前台和后台，只保留 Astro 前台与 Vue 后台。

启动脚本：
  - dev.sh 默认固定端口 8090。
  - 支持 memory / mysql 模式。
  - 支持本机无 npm 时用 Docker Node 构建。
  - Go 后端默认统一使用 Docker Go 启动。
  - 需要临时使用宿主机 Go 时，可加 `--local-go`。
  - 支持 GOPROXY 国内代理配置。
  - 支持检测端口占用、复用已有 DevHub 服务。
  - 支持 ./dev.sh --restart 重启已有服务。

数据模型：
  - 已有 sites、boards、posts、comments、tags、notifications 兼容模型。
  - 已有 users、roles、permissions、role_permissions、user_roles、refresh_tokens。
  - 已有 admin_users、admin_roles、admin_settings、admin_logs。
  - Go 内置 schema 已新增 communities、categories、topics、topic_tags。
  - Go 内置 schema 已新增 reactions、favorites、follows、activities。
  - Go 内置 schema 已新增 reports、community_moderators、wiki_pages、wiki_revisions。

后端 API：
  - 已有 /api/v1/health。
  - 已有 sites、boards、posts、comments、search、hot、tags、feed、notifications。
  - 已有 communities、categories、topics、search/topics。
  - 已有 reactions/favorites/follows toggle 基础接口。
  - 已有 admin 登录、概览、内容、评论、站点、用户、角色、标签、设置、日志、通知推送接口。

前台：
  - 已实现总站、子站页、搜索页、标签页、内容详情页。
  - 已支持 /site/:site/ 与 /c/:site/。
  - 登录 / 注册入口已统一到导航右上角。
  - 站点选择菜单已从 hover 改为点击控制，避免鼠标移走消失。
  - 搜索页已支持读取 URL 参数并动态拉取列表。
  - 标签点击已进入搜索筛选。
  - 点赞与评论区已接入基础接口。

后台：
  - 已保留 Vue 3 + Element Plus 后台作为唯一维护后台。
  - 已支持 /admin-next 和 /admin-next/... 深层路由。
  - 已有控制台、内容管理、评论审核、站点管理、用户权限、运营工具、数据统计、系统设置页面。
```

### 0.3 部分完成

```text
数据：
  - db/mysql/001_schema.sql 仍主要是旧兼容表。
  - 新通用社区表已在 internal/store/schema.go 中存在，但需要同步到正式 SQL 或 migration。
  - 后端 MemoryStore 的站点 seed 仍偏 PHP / Go / Java。
  - AI / Frontend 子站在前台 fallback 数据中已补齐，但后端 seed 还需完整覆盖。

前台：
  - 当前前台仍主要消费 sites/posts 兼容 API，并优先尝试 topics API。
  - 还没有完全切换到 communities/topics。
  - 前台发帖、编辑、草稿、Markdown、标签选择、内容类型选择还未完整闭环。

业务：
  - 收藏、关注、动态接口已有基础能力，但前台列表页还未完成。
  - 举报、版主、Wiki 已建模，但完整业务流程和后台 UI 还未完成。
  - 问答采纳最佳答案接口已有基础能力，但前台展示和流程仍需完善。
  - RBAC 已有中间件和权限点，但还需要更多越权测试。
```

### 0.4 下一步优先级

```text
1. 同步 MySQL schema：把 internal/store/schema.go 中的新通用社区表同步到 db/mysql/001_schema.sql 或新增正式 migration。
2. 补齐 seed：让 MemoryStore 和 MySQLStore 都生成 PHP、Go、Java、AI、Frontend 五个子站及完整板块、标签、Topic、评论。
3. 前台 API 统一：逐步从 sites/posts 兼容 API 迁移到 communities/topics。
4. 完成发布流程：发帖、编辑、草稿、Markdown、标签选择、内容类型选择。
5. 搜索增强：完善搜索范围、分类筛选、排序筛选和空状态。
6. 互动闭环：完成收藏列表、关注列表、动态流和通知联动。
7. 问答能力：完成未解决筛选、采纳最佳答案、最佳答案展示。
8. 治理能力：完成举报、版主、精华、置顶、隐藏、评论锁定。
9. 后台 CRUD：完善 communities、categories、topics、reports、moderators、wiki 管理页面。
10. 测试与部署：补测试矩阵、CI、生产配置、备份恢复和回滚文档。
```

### 0.5 当前验收状态

```text
[x] 项目名称保持 DevHub
[x] 前台可通过 / 打开
[x] 后台可通过 /admin-next 打开
[x] 默认端口固定为 8090
[x] 旧原生前台和后台已移除
[x] 子站页面可访问
[x] 搜索页可访问
[x] 内容详情页可访问
[x] 标签点击可进入筛选
[x] 登录注册入口在导航右上角
[x] GET /api/v1/communities 可用
[x] GET /api/v1/topics 可用
[x] GET /api/v1/search/topics 可用
[ ] MySQL SQL 文件与 Go 内置 schema 完全同步
[ ] 后端 seed 完整覆盖 AI / Frontend
[ ] 前台完全切换到通用社区 API
[ ] Wiki、举报、版主完整业务流程完成
```

---

## 1. 背景与目标

DevHub 当前已经有前端页面雏形，方向是参考 LearnKu 的社区结构，但不是照搬页面，而是抽象出一个可复用的通用技术社区系统。

LearnKu 的核心不是单纯帖子列表，而是：

```text
总站
  ├── 多个技术子站
  ├── 每个子站有自己的导航、板块、标签和内容
  ├── 内容类型包括帖子、问答、Wiki、文档、招聘、项目、等
  ├── 用户通过评论、点赞、收藏、关注、动态参与社区
  └── 通过版主、举报、声望、精华、置顶等机制治理内容质量
```

DevHub 的目标是形成类似结构：

```text
DevHub 总站
  ├── PHP 子站
  ├── Go 子站
  ├── Java 子站
  ├── AI 子站
  ├── 前端子站
  └── 其他可扩展子站
```

每个子站应该拥有：

```text
社区
问答中心
开源项目

招聘内推
Wiki
文档
标签合集
搜索
动态
```

---

## 2. 本次迭代总目标

本次迭代需要完成 DevHub 从“静态社区首页”到“可运行的社区程序骨架”的升级。

### 2.1 必须完成

1. 保持现有 DevHub 页面视觉方向，不要大面积推翻重写。
2. 建立多子站数据模型。
3. 建立板块 / 分类数据模型。
4. 建立内容主题 Topic 数据模型。
5. 建立评论、标签、点赞、收藏基础模型。
6. 支持首页聚合不同子站内容。
7. 支持进入某个子站后，只展示当前子站内容。
8. 支持搜索框选择搜索范围：总站 / 当前子站 / 指定子站。
9. 支持帖子列表筛选：最新、热门、精华、未解决。
10. 支持子站右侧标签合集展示。
11. 提供基础后台管理能力的数据结构和接口。
12. 提供 Demo Seed 数据，便于页面验证。

### 2.2 不要求本次完成

以下能力可以先预留结构，不要求完整实现：

1. AI 摘要真实调用。
2. 推荐算法。
3. 复杂积分等级系统。
4. 完整后台 UI。
5. Wiki 多人协作审批流。
6. 私信系统。
7. 复杂通知推送。
8. 商业化广告结算系统。

---

## 3. 技术执行原则

Codex 执行时需要遵守以下规则。

### 3.1 项目命名

所有页面、变量、文案中，网站名称统一使用：

```text
DevHub
```

禁止出现：

```text
LearnKu Clone
LearnKu
Laravel China
```

除非是在注释中说明参考来源，否则不要使用这些名称。

### 3.2 优先基于现有代码修改

不要一上来全量重构。

执行顺序：

```text
1. 先阅读当前项目目录
2. 判断当前使用的技术栈
3. 找到现有页面入口、路由、组件和样式
4. 在现有结构上补充功能
5. 只有在当前结构无法运行时，才进行局部重构
```

### 3.3 前后端分层

如果当前项目已有后端框架，则优先遵守当前框架。

如果是 Laravel / Hyperf / ThinkPHP 类项目，优先生成：

```text
migration
model
controller
request
resource
service
seeder
route
test
```

如果当前只有前端页面，则先实现：

```text
mock 数据层
页面状态管理
API 接口约定文档
后端目录骨架建议
```

### 3.4 不要破坏当前页面

已有可运行页面必须继续可运行。

每次修改后需要保证：

```text
页面能打开
导航能点击
子站能切换
搜索框正常显示
列表不为空
样式不严重错位
```

---

## 4. DevHub 信息架构

### 4.1 总站结构

总站负责聚合所有子站内容。

总站页面包含：

```text
顶部导航
子站切换区
搜索框
全站热门
最新内容
推荐子站
动态流
右侧标签 / 推荐内容 / 公告
```

总站内容来源：

```text
所有启用状态的子站
所有公开状态的帖子
所有公开状态的问答
所有公开状态的项目
所有公开状态的 AI 作品
```

### 4.2 子站结构

每个子站是一个独立社区空间。

子站页面包含：

```text
子站 Logo
子站名称
子站简介
子站导航
发帖按钮
内容列表
排序筛选
标签合集
版主信息
社区规则
```

子站导航固定为：

```text
社区
问答中心
开源项目

招聘内推
Wiki
文档
```

### 4.3 内容类型

Topic 作为通用内容表，使用 `content_type` 区分不同类型。

```text
article      普通帖子 / 分享
question     问答
project      开源项目
ai_work      
job          招聘内推
wiki         Wiki索引或Wiki关联内容
 doc         文档
news         周刊 / 公告
```

注意：Wiki 后续应该独立成 `wiki_pages`，但 Topic 可以作为聚合列表入口。

---

## 5. 页面功能迭代

## 5.1 顶部导航

### 目标

顶部导航需要体现 DevHub 的多子站结构。

### 功能要求

1. DevHub Logo 点击返回总站首页。
2. 子站切换入口和“我的动态 / 全站热门 / 搜索”等一级导航保持同层级。
3. 当前子站高亮显示。
4. 点击 PHP / Go / Java 等子站后，页面内容切换为对应子站。
5. 不要把子站切换放得太低，应该是全局级入口。

### 验收标准

```text
点击 DevHub Logo → 回到总站首页
点击 PHP 子站 → 进入 PHP 子站首页
点击 Go 子站 → 进入 Go 子站首页
点击 Java 子站 → 进入 Java 子站首页
当前子站导航状态明显
```

---

## 5.2 搜索框

### 目标

搜索框需要支持总站搜索和子站搜索。

### 功能要求

搜索框左侧增加子站范围选择。

可选项：

```text
全站
当前子站
PHP
Go
Java
AI
前端
```

行为规则：

```text
用户在总站搜索：默认范围为全站
用户在 PHP 子站搜索：默认范围为当前子站 PHP
用户手动选择 Go：只搜索 Go 子站
```

搜索结果页左侧需要增加分类筛选：

```text
全部
社区
问答中心
开源项目

招聘内推
Wiki
文档
```

### 验收标准

```text
搜索框能输入关键词
搜索范围下拉能选择
搜索结果能展示范围说明
搜索结果左侧有分类筛选
切换分类后列表内容变化或 mock 数据变化
```

---

## 5.3 内容列表

### 目标

内容列表是社区核心页面。

### 列表字段

每条内容至少展示：

```text
标题
内容类型
所属子站
所属板块
作者
发布时间
最后活跃时间
评论数
点赞数
浏览数
标签
是否精华
是否置顶
是否已解决
```

### 排序

支持：

```text
最新
活跃
热门
精华
未解决
```

### 热门分计算建议

早期可以使用简单公式：

```text
hot_score = view_count * 1 + comment_count * 5 + like_count * 3 + favorite_count * 4
```

后续可以引入时间衰减。

### 验收标准

```text
列表正常渲染
不同内容类型有明显标识
点击排序后列表顺序变化
点击标签后进入标签聚合页或筛选列表
```

---

## 5.4 子站右侧标签合集

### 目标

每个子站右侧展示当前子站的热门标签。

### 功能要求

标签展示字段：

```text
标签名
内容数量
关注数量，可选
热度，可选
```

排序规则：

```text
优先按 topic_count 降序
其次按 follow_count 降序
```

### 验收标准

```text
总站显示全站热门标签
PHP 子站只显示 PHP 标签
Go 子站只显示 Go 标签
Java 子站只显示 Java 标签
```

---

## 5.5 帖子详情页

### 目标

帖子详情页需要支持社区内容阅读和互动。

### 页面字段

```text
标题
作者信息
子站
分类
标签
发布时间
更新时间
浏览数
点赞数
收藏数
评论数
正文
评论区
相关内容
```

### 操作功能

```text
点赞
收藏
评论
回复评论
举报
编辑，作者或管理员可见
删除，作者或管理员可见
采纳答案，问答类型可见
```

### AI 摘要预留

页面可以预留 AI 摘要区域，但本阶段可以先使用 mock 字段：

```text
ai_summary
```

没有摘要时不显示。

---

## 6. 后端数据模型设计

以下为推荐表结构。请 Codex 根据当前项目 ORM / migration 规范生成。

---

## 6.1 communities 子站表

```sql
CREATE TABLE communities (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  name VARCHAR(64) NOT NULL COMMENT '子站名称，如 PHP、Go、Java',
  slug VARCHAR(64) NOT NULL UNIQUE COMMENT '子站标识，如 php、go、java',
  logo VARCHAR(255) NULL,
  description VARCHAR(500) NULL,
  sort_order INT NOT NULL DEFAULT 0,
  status TINYINT NOT NULL DEFAULT 1 COMMENT '1启用 0禁用',
  created_at TIMESTAMP NULL,
  updated_at TIMESTAMP NULL,
  deleted_at TIMESTAMP NULL
);
```

---

## 6.2 categories 板块表

```sql
CREATE TABLE categories (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  community_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '0表示全站通用板块',
  name VARCHAR(64) NOT NULL,
  slug VARCHAR(64) NOT NULL,
  type VARCHAR(32) NOT NULL COMMENT 'article/question/project/ai_work/job/wiki/doc/news',
  description VARCHAR(500) NULL,
  icon VARCHAR(100) NULL,
  sort_order INT NOT NULL DEFAULT 0,
  status TINYINT NOT NULL DEFAULT 1,
  created_at TIMESTAMP NULL,
  updated_at TIMESTAMP NULL,
  deleted_at TIMESTAMP NULL,
  INDEX idx_community_type (community_id, type),
  INDEX idx_slug (slug)
);
```

---

## 6.3 topics 内容主题表

```sql
CREATE TABLE topics (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  community_id BIGINT UNSIGNED NOT NULL,
  category_id BIGINT UNSIGNED NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  title VARCHAR(200) NOT NULL,
  slug VARCHAR(220) NULL,
  content_type VARCHAR(32) NOT NULL DEFAULT 'article',
  summary VARCHAR(500) NULL,
  content MEDIUMTEXT NOT NULL,
  ai_summary TEXT NULL,
  cover_image VARCHAR(255) NULL,
  status TINYINT NOT NULL DEFAULT 1 COMMENT '1正常 0隐藏 2审核中 3已删除',
  is_pinned TINYINT NOT NULL DEFAULT 0,
  is_featured TINYINT NOT NULL DEFAULT 0,
  is_solved TINYINT NOT NULL DEFAULT 0,
  best_comment_id BIGINT UNSIGNED NULL,
  view_count INT UNSIGNED NOT NULL DEFAULT 0,
  comment_count INT UNSIGNED NOT NULL DEFAULT 0,
  like_count INT UNSIGNED NOT NULL DEFAULT 0,
  favorite_count INT UNSIGNED NOT NULL DEFAULT 0,
  hot_score INT UNSIGNED NOT NULL DEFAULT 0,
  last_active_at TIMESTAMP NULL,
  created_at TIMESTAMP NULL,
  updated_at TIMESTAMP NULL,
  deleted_at TIMESTAMP NULL,
  INDEX idx_community_type_status (community_id, content_type, status),
  INDEX idx_category_status (category_id, status),
  INDEX idx_hot_score (hot_score),
  INDEX idx_last_active_at (last_active_at),
  FULLTEXT KEY ft_title_content (title, content)
);
```

---

## 6.4 comments 评论表

```sql
CREATE TABLE comments (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  topic_id BIGINT UNSIGNED NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  parent_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  reply_user_id BIGINT UNSIGNED NULL,
  content TEXT NOT NULL,
  like_count INT UNSIGNED NOT NULL DEFAULT 0,
  status TINYINT NOT NULL DEFAULT 1 COMMENT '1正常 0隐藏 2审核中',
  created_at TIMESTAMP NULL,
  updated_at TIMESTAMP NULL,
  deleted_at TIMESTAMP NULL,
  INDEX idx_topic_status (topic_id, status),
  INDEX idx_parent_id (parent_id)
);
```

---

## 6.5 tags 标签表

```sql
CREATE TABLE tags (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  community_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  name VARCHAR(64) NOT NULL,
  slug VARCHAR(64) NOT NULL,
  description VARCHAR(500) NULL,
  topic_count INT UNSIGNED NOT NULL DEFAULT 0,
  follow_count INT UNSIGNED NOT NULL DEFAULT 0,
  status TINYINT NOT NULL DEFAULT 1,
  created_at TIMESTAMP NULL,
  updated_at TIMESTAMP NULL,
  deleted_at TIMESTAMP NULL,
  UNIQUE KEY uk_community_slug (community_id, slug),
  INDEX idx_topic_count (topic_count)
);
```

---

## 6.6 topic_tags 主题标签关联表

```sql
CREATE TABLE topic_tags (
  topic_id BIGINT UNSIGNED NOT NULL,
  tag_id BIGINT UNSIGNED NOT NULL,
  PRIMARY KEY (topic_id, tag_id),
  INDEX idx_tag_id (tag_id)
);
```

---

## 6.7 reactions 点赞表

```sql
CREATE TABLE reactions (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  target_type VARCHAR(32) NOT NULL COMMENT 'topic/comment/wiki',
  target_id BIGINT UNSIGNED NOT NULL,
  reaction_type VARCHAR(32) NOT NULL DEFAULT 'like',
  created_at TIMESTAMP NULL,
  updated_at TIMESTAMP NULL,
  UNIQUE KEY uk_user_target_reaction (user_id, target_type, target_id, reaction_type),
  INDEX idx_target (target_type, target_id)
);
```

---

## 6.8 favorites 收藏表

```sql
CREATE TABLE favorites (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  target_type VARCHAR(32) NOT NULL COMMENT 'topic/wiki/project',
  target_id BIGINT UNSIGNED NOT NULL,
  created_at TIMESTAMP NULL,
  UNIQUE KEY uk_user_target (user_id, target_type, target_id),
  INDEX idx_target (target_type, target_id)
);
```

---

## 6.9 follows 关注表

```sql
CREATE TABLE follows (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  target_type VARCHAR(32) NOT NULL COMMENT 'user/topic/community/tag',
  target_id BIGINT UNSIGNED NOT NULL,
  created_at TIMESTAMP NULL,
  UNIQUE KEY uk_user_target (user_id, target_type, target_id),
  INDEX idx_target (target_type, target_id)
);
```

---

## 6.10 activities 动态表

```sql
CREATE TABLE activities (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  community_id BIGINT UNSIGNED NULL,
  action VARCHAR(64) NOT NULL COMMENT 'created_topic/commented/liked/followed/favorited',
  target_type VARCHAR(32) NOT NULL,
  target_id BIGINT UNSIGNED NOT NULL,
  remark VARCHAR(500) NULL,
  created_at TIMESTAMP NULL,
  INDEX idx_user_created (user_id, created_at),
  INDEX idx_community_created (community_id, created_at),
  INDEX idx_target (target_type, target_id)
);
```

---

## 6.11 reports 举报表

```sql
CREATE TABLE reports (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  reporter_id BIGINT UNSIGNED NOT NULL,
  target_type VARCHAR(32) NOT NULL COMMENT 'topic/comment/user/wiki',
  target_id BIGINT UNSIGNED NOT NULL,
  reason_type VARCHAR(64) NOT NULL,
  reason_text VARCHAR(1000) NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'pending' COMMENT 'pending/accepted/rejected',
  handled_by BIGINT UNSIGNED NULL,
  handled_at TIMESTAMP NULL,
  created_at TIMESTAMP NULL,
  updated_at TIMESTAMP NULL,
  INDEX idx_status (status),
  INDEX idx_target (target_type, target_id)
);
```

---

## 6.12 community_moderators 子站版主表

```sql
CREATE TABLE community_moderators (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  community_id BIGINT UNSIGNED NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  role VARCHAR(32) NOT NULL DEFAULT 'moderator' COMMENT 'moderator/senior_moderator',
  status TINYINT NOT NULL DEFAULT 1,
  created_at TIMESTAMP NULL,
  updated_at TIMESTAMP NULL,
  UNIQUE KEY uk_community_user (community_id, user_id)
);
```

---

## 6.13 wiki_pages Wiki 表，预留

```sql
CREATE TABLE wiki_pages (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  community_id BIGINT UNSIGNED NOT NULL,
  category_id BIGINT UNSIGNED NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  title VARCHAR(200) NOT NULL,
  slug VARCHAR(220) NULL,
  summary VARCHAR(500) NULL,
  content MEDIUMTEXT NOT NULL,
  status TINYINT NOT NULL DEFAULT 1,
  view_count INT UNSIGNED NOT NULL DEFAULT 0,
  like_count INT UNSIGNED NOT NULL DEFAULT 0,
  created_at TIMESTAMP NULL,
  updated_at TIMESTAMP NULL,
  deleted_at TIMESTAMP NULL,
  INDEX idx_community_status (community_id, status)
);
```

---

## 6.14 wiki_revisions Wiki 版本表，预留

```sql
CREATE TABLE wiki_revisions (
  id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  wiki_page_id BIGINT UNSIGNED NOT NULL,
  editor_id BIGINT UNSIGNED NOT NULL,
  title VARCHAR(200) NOT NULL,
  content MEDIUMTEXT NOT NULL,
  change_note VARCHAR(500) NULL,
  created_at TIMESTAMP NULL,
  INDEX idx_wiki_page_id (wiki_page_id)
);
```

---

## 7. API 设计建议

根据当前项目路由风格实现即可。以下为建议接口。

---

## 7.1 子站接口

```http
GET /api/communities
GET /api/communities/{slug}
GET /api/communities/{slug}/home
GET /api/communities/{slug}/tags
GET /api/communities/{slug}/categories
```

### 返回重点

```json
{
  "id": 1,
  "name": "PHP",
  "slug": "php",
  "logo": "",
  "description": "PHP 开发者社区",
  "categories": [],
  "hot_tags": [],
  "stats": {
    "topic_count": 120,
    "user_count": 30
  }
}
```

---

## 7.2 内容接口

```http
GET /api/topics
GET /api/topics/{id}
POST /api/topics
PUT /api/topics/{id}
DELETE /api/topics/{id}
```

### 查询参数

```text
community_slug
category_slug
content_type
sort: latest / active / hot / featured
status
is_solved
tag
keyword
page
page_size
```

---

## 7.3 评论接口

```http
GET /api/topics/{topic_id}/comments
POST /api/topics/{topic_id}/comments
PUT /api/comments/{id}
DELETE /api/comments/{id}
```

---

## 7.4 互动接口

```http
POST /api/reactions/toggle
POST /api/favorites/toggle
POST /api/follows/toggle
POST /api/reports
```

### Toggle 请求示例

```json
{
  "target_type": "topic",
  "target_id": 1
}
```

---

## 7.5 搜索接口

```http
GET /api/search
```

### 参数

```text
keyword
scope: all / current / community
community_slug
content_type
sort
page
page_size
```

### 行为

```text
scope = all：搜索全站
scope = current：由前端传当前 community_slug
scope = community：搜索指定 community_slug
```

---

## 8. Demo 数据要求

Codex 需要创建 Seeder / mock 数据，让页面可以立即验证。

### 8.1 子站数据

```text
DevHub 总站不作为 community 记录，只作为聚合入口。

PHP
Go
Java
AI
Frontend
```

### 8.2 每个子站默认板块

```text
社区
问答中心
开源项目

招聘内推
Wiki
文档
```

### 8.3 标签数据示例

PHP 子站：

```text
Laravel
Hyperf
Composer
Swoole
Redis
MySQL
性能优化
代码规范
```

Go 子站：

```text
Gin
Gorm
微服务
并发
gRPC
Docker
性能优化
```

Java 子站：

```text
Spring Boot
MyBatis
JVM
Redis
MySQL
微服务
消息队列
```

AI 子站：

```text
AI Agent
Prompt
RAG
OpenAI
Claude
Codex
工作流
```

Frontend 子站：

```text
Vue
React
TypeScript
Tailwind CSS
Vite
Next.js
性能优化
```

### 8.4 Topic 示例

每个子站至少生成：

```text
3 条社区帖子
3 条问答
2 条开源项目
2 条 
2 条招聘内推
2 条 Wiki
2 条文档
```

每条内容需要随机填充：

```text
view_count
comment_count
like_count
favorite_count
hot_score
is_featured
is_pinned
is_solved
last_active_at
```

---

## 9. 前端页面改造任务

Codex 需要根据现有页面结构实现以下页面或组件。

### 9.1 页面

```text
/                      DevHub 总站首页
/c/:communitySlug       子站首页
/c/:communitySlug/:type 子站板块页
/topics/:id             帖子详情页
/search                 搜索结果页
/tags/:slug             标签聚合页
```

如果当前项目不是以上路由风格，可以按照现有路由系统适配。

---

## 9.2 组件

```text
SiteHeader              顶部导航
CommunitySwitcher       子站切换
SearchBox               带范围选择的搜索框
CommunityNav            子站板块导航
TopicList               内容列表
TopicCard               内容卡片
TagCloud                标签合集
RightSidebar            右侧边栏
CommunityStats          子站统计
ActivityFeed            动态流
TopicDetail             内容详情
CommentList             评论列表
```

---

## 10. 权限规则

### 10.1 普通用户

```text
浏览公开内容
发帖
评论
点赞
收藏
关注
举报
编辑自己的内容
删除自己的内容，软删除
```

### 10.2 版主

版主只管理自己负责的子站。

```text
设置精华
设置置顶
隐藏内容
处理举报
锁定评论
编辑分类
```

### 10.3 管理员

```text
管理所有子站
管理所有分类
管理所有内容
管理所有用户
管理版主
处理全站举报
```

---

## 11. 业务规则

## 11.1 首页聚合规则

总站首页展示所有子站内容。

```text
最新内容：按 created_at 倒序
热门内容：按 hot_score 倒序
活跃内容：按 last_active_at 倒序
推荐内容：is_featured = 1
```

## 11.2 子站内容规则

进入子站后，所有列表默认加上：

```text
community_id = 当前子站 ID
```

## 11.3 问答规则

当 `content_type = question` 时：

```text
显示是否已解决
允许提问者采纳最佳答案
未解决筛选只展示 is_solved = 0
采纳后 is_solved = 1，并写入 best_comment_id
```

## 11.4 标签规则

```text
标签可属于全站，也可属于某个子站
子站页面优先展示当前子站标签
全站页面展示所有子站热门标签
创建内容时最多选择 5 个标签
标签不存在时，早期可以不允许自动创建，后续再做
```

## 11.5 热度更新规则

当以下事件发生时，需要更新 topic 统计：

```text
新增评论：comment_count + 1，更新 last_active_at
点赞：like_count + 1 或 -1
收藏：favorite_count + 1 或 -1
浏览：view_count + 1
```

然后重新计算：

```text
hot_score = view_count + comment_count * 5 + like_count * 3 + favorite_count * 4
```

---

## 12. 后台管理预留

本阶段可以不做完整后台页面，但后端结构需要支持。

后台模块包括：

```text
子站管理
分类管理
标签管理
帖子管理
评论管理
举报管理
版主管理
Wiki管理
广告位管理，预留
友情链接管理，预留
```

---

## 13. 测试与验收

Codex 修改完成后，需要保证以下验收项通过。

### 13.1 页面验收

```text
首页能正常打开
子站能正常切换
搜索框能显示范围选择
搜索结果页能显示分类筛选
子站右侧能显示标签合集
帖子列表不为空
帖子详情页能打开
评论区能展示
```

### 13.2 数据验收

```text
communities 有 PHP / Go / Java / AI / Frontend 数据
每个 community 有默认 categories
每个 community 有 tags
每个 community 有 topics
topic 和 tag 有关联
topic 有统计字段
```

### 13.3 接口验收

```text
GET /api/communities 可用
GET /api/topics 可用
GET /api/topics?community_slug=php 可用
GET /api/search?keyword=xxx 可用
GET /api/communities/php/tags 可用
```

### 13.4 兼容验收

```text
不要改坏已有页面
不要删除已有核心组件
不要把 DevHub 改名
不要引入无法安装的大型依赖
不要因为 mock 数据缺失导致空白页
```

---

## 14. 建议的迭代顺序

Codex 应该按照以下顺序执行。

### 第一步：检查项目

```text
阅读 package.json / composer.json
确认前端框架
确认后端框架
确认路由结构
确认页面入口
确认数据库配置
```

### 第二步：补数据模型

```text
communities
categories
topics
comments
tags
topic_tags
reactions
favorites
follows
activities
reports
```

### 第三步：补 Seed 数据

```text
创建 PHP / Go / Java / AI / Frontend 子站
创建默认板块
创建默认标签
创建 Demo Topics
创建 Demo Comments
```

### 第四步：补 API

```text
communities API
topics API
comments API
search API
tags API
interactions API
```

### 第五步：接前端页面

```text
首页接聚合数据
子站页接 community 数据
列表接 topics 数据
搜索页接 search 数据
右侧标签接 tags 数据
详情页接 topic detail 数据
```

### 第六步：优化体验

```text
修复页面间距
修复导航居中
修复空数据状态
修复移动端基础适配
修复 loading 状态
修复错误提示
```

---

## 15. Codex 执行提示词

可以直接把下面这段作为 Codex 的任务输入。

```text
请基于当前仓库实现 DevHub 通用社区程序的第一阶段迭代。

注意：项目名称必须始终保持为 DevHub，不要改成 LearnKu 或其他名称。

目标：把当前 DevHub 页面升级为支持多技术子站的社区程序骨架。请先阅读当前项目结构，判断技术栈和路由方式，然后在现有代码基础上增量实现，不要无理由全量重构。

需要实现：
1. 多子站模型：PHP、Go、Java、AI、Frontend。
2. 每个子站包含固定板块：社区、问答中心、开源项目、、招聘内推、Wiki、文档。
3. 主题内容模型 Topic，支持 article、question、project、ai_work、job、wiki、doc 等 content_type。
4. 标签系统，子站右侧展示当前子站热门标签合集。
5. 搜索框增加搜索范围选择：全站、当前子站、指定子站。
6. 搜索结果页左侧增加分类筛选。
7. 首页展示全站聚合内容，进入子站后只展示当前子站内容。
8. 帖子列表支持最新、活跃、热门、精华、未解决筛选。
9. 帖子详情页支持基础信息、正文、标签、评论区、点赞、收藏、举报入口。
10. 生成足够 Demo 数据，确保页面打开不是空白。

如果当前项目已有后端，请生成对应 migration、model、controller、route、seeder、request/resource/service，并接入前端。

如果当前项目暂时只有前端，请先用 mock 数据实现完整页面交互，同时补充 API 约定文件，方便后续后端实现。

修改完成后请检查：
- 页面能运行
- 首页能打开
- 子站能切换
- 搜索框正常显示
- 子站标签合集正常显示
- 帖子列表不为空
- 帖子详情页能打开
- DevHub 名称没有被替换
```

---

## 16. 本阶段最终交付物

当前交付状态：

```text
[x] 1. 可运行的 DevHub 页面
[x] 2. 多子站数据结构
[x] 3. 板块 / 标签 / 内容 Demo 数据，前台 fallback 已覆盖 AI / Frontend，后端 seed 仍需补齐
[x] 4. 基础 API 或 mock API，兼容 API 与通用社区 API 已并存
[x] 5. 搜索范围选择，搜索页已支持 URL 参数和动态筛选
[x] 6. 子站内容筛选
[x] 7. 标签合集展示
[x] 8. 帖子详情页
[x] 9. 评论展示
[x] 10. README 或变更说明，已补 README、docs/README.md、docs/PROJECT_PROGRESS.md
[ ] 11. MySQL SQL 文件与 Go 内置 schema 完全同步
[ ] 12. 前台完全切换到 communities/topics 通用社区 API
[ ] 13. Wiki、举报、版主完整业务流程完成
```

---

## 17. 后续阶段规划

### 第二阶段

```text
登录注册
发帖编辑器
Markdown 支持
评论发布
点赞收藏真实写入
问答采纳最佳答案
通知系统
用户主页
```

### 第三阶段

```text
Wiki 独立模块
Wiki 版本记录
文档中心
开源项目库
库
动态流
用户关注
用户声望
```

### 第四阶段

```text
后台管理
举报处理
版主系统
精华置顶
内容审核
广告位
赞助商
友情链接
数据统计
```

### 第五阶段

```text
AI 摘要
AI 自动打标签
相似问题推荐
个性化 Feed
搜索优化
推荐算法
```

---

## 18. 重点提醒

1. DevHub 是通用社区程序，不是 Laravel 专属社区。
2. 多子站是一级核心能力。
3. 帖子、问答、项目、、招聘、Wiki、文档都应该能被统一聚合。
4. 子站进入后必须只看当前子站内容。
5. 搜索必须有范围感。
6. 标签必须跟子站关联。
7. 右侧标签合集是当前页面的重要组成部分。
8. 本阶段优先保证可运行和结构正确，不追求复杂算法。
9. 所有功能都要围绕后续可扩展来设计。
10. 不要破坏当前已有页面。

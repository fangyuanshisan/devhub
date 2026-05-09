# DevHub Web

[返回文档大纲](../docs/README.md)

`web/` 目录只保留当前维护中的前台、后台源码与构建产物。

## 目录结构

```text
web/
├── frontend-app/     # Astro + Vue Islands 前台源码
├── frontend/         # Astro 构建产物，由 Go 服务托管
├── admin-app/        # Vue 3 + Element Plus 后台源码
├── admin-vue/        # Vue 后台构建产物，由 Go 服务托管
└── README.md
```

旧的原生 HTML / CSS / JS 前台与后台已经移除，包括旧 `web/admin`、`web/assets`、`web/pages`、`web/api`、`web/components`、`web/utils` 等目录。

## 前台

源码目录：

```text
web/frontend-app
```

构建命令：

```bash
cd web/frontend-app
npm install
npm run build
```

构建产物输出到：

```text
web/frontend
```

Go 服务托管入口：

```text
/
/site/php/
/c/php/
/search/
/topics/new/
/c/:site/topics/new/
/topics/:id/        Go 动态输出 SEO HTML
/posts/:id/         兼容入口，301 跳转到 /topics/:id/
/tags/:tag/
/me/favorites
/me/follows
/me/activities
/notifications
/me/notifications
```

当前前台用户中心页面已经提供“我的收藏”“我的关注”“我的动态”和“通知中心”。`/topics/:id/` 的评论区、回复和问答采纳由 Go 动态 SEO HTML 内的运行时脚本加载，不需要重新构建即可读取新评论和最佳答案状态。

## 后台

源码目录：

```text
web/admin-app
```

构建命令：

```bash
cd web/admin-app
npm install
npm run build
```

构建产物输出到：

```text
web/admin-vue
```

Go 服务托管入口：

```text
/admin-next
/admin-next?site=php
```

旧入口仍保留为重定向：

```text
/admin      -> /admin-next
/admin/php  -> /admin-next?site=php
```

## 推荐启动方式

优先使用仓库根目录的一键脚本：

```bash
./dev.sh
./dev.sh --mysql
DOCKER="sudo docker" ./dev.sh --no-build
```

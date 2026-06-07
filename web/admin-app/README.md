# DevHub Admin

[返回文档大纲](../../docs/README.md)

Vue 3 + Vite + Element Plus 后台，是 DevHub 当前唯一维护的后台前端。

## 技术栈

- Vue 3
- Vite
- Element Plus
- Pinia
- Vue Router
- Axios
- Toast UI Editor

## 开发

```bash
cd web/admin-app
npm install
npm run dev
```

开发服务会代理 `/api` 到 `http://127.0.0.1:8090`。请先在仓库根目录启动后端：

```bash
./dev.sh
```

## 构建

```bash
cd web/admin-app
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

## E2E

后台 Playwright E2E 优先使用仓库根目录的 Compose runner：

```bash
docker compose run --rm admin-e2e npx playwright test tests/e2e/plugin-governance.spec.js
```

默认 baseURL 为 `http://devhub:8090`。Compose 会等待 `devhub` 的 `/api/v1/health` 变为 healthy；如 E2E 失败，可运行 `./scripts/check-frontend.sh --admin-only` 获取额外诊断。

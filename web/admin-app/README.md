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

# GoGopher Arch Web

React + TypeScript frontend for the Capability/Evidence learning loop.

## Development

```bash
npm install
npm run dev
```

Vite 默认把 `/api` 代理到 `http://localhost:8080`。从仓库根目录执行 `./scripts/dev.sh backend` 可通过 loopback development overlay 启动 Gateway、Sandbox 和 PostgreSQL。

## Verification

```bash
npm test -- --run
npm run build
npm run e2e:compose
```

客户端只展示 Learning API 返回的 Activity、Attempt、Execution、Evidence、Snapshot 和 queue state，不复制 held-out content 或服务端掌握判定。

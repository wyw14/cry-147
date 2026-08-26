基于 Go 实现的 CellForge 项目，一款锂电芯化成、分容、静置与安全监护联控服务，统一维护设备会话、工步状态和隔离结果。

# CellForge

CellForge 运行一个本地 HTTP 服务，提供化成批次、充放电设备、保护联锁和异常事件的运营页面。服务使用文件事件日志和原子快照持久化，不依赖外部数据库。

## 运行

使用固定 Go 工具链执行：

```text
go run -mod=vendor ./cmd/cellforge
```

默认监听 `127.0.0.1:21247`。可通过 `CELLFORGE_ADDR`、`CELLFORGE_DATA`、`CELLFORGE_WEB` 和 `CELLFORGE_THERMAL_URL` 配置监听地址、数据目录、页面目录及温箱入口。

## 页面

- `/operations`
- `/equipment`
- `/interlocks`
- `/incidents`

健康探测位于 `/healthz`，对应业务数据由 `/api/operations`、`/api/equipment`、`/api/interlocks` 和 `/api/incidents` 提供。

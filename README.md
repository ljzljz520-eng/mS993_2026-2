# 香薰产品维护台

这是一个单一 Go module 的纯后端示例，使用 Go 1.25.13 和固定内存 fixture 模拟香薰商品维护台。商品分为蜡烛、精油、扩香石和礼盒，提供商品浏览页、价格调整、图片地址维护、库存变更、操作日志和任务处理入口。视图交互是由 Go HTTP 入口直接提供的 HTML/CSS/JS，不需要前端包。

## 运行

```bash
go run ./cmd/aroma
```

打开 `http://localhost:8080/` 可浏览固定商品。接口包括：

- `GET /api/products`
- `POST /api/products/{id}/price`，JSON：`{"price":"139.00"}`
- `POST /api/products/{id}/image`，JSON：`{"imageUrl":"/uploads/new.png"}`
- `POST /api/products/{id}/stock`，JSON：`{"delta":2,"reason":"restock"}`
- `GET /api/logs`
- `GET /api/tasks`
- `POST /api/tasks/{taskID}/run`，JSON：`{"worker":"worker-a"}`

## 业务链路

```bash
go test -count=1 ./...
```

业务链路包含两个固定 worker 同时领取 `task-stock-refresh-001` 的场景。当前任务领取器把读取状态与完成标记分成两个操作，因此两个 worker 都可能执行库存副作用；对应验收测试会稳定失败，作为重复消费 bug 的回归信号。正确结果应是库存只增加一次，并由另一个 worker 返回 `already-processed`。

## 打包

```bash
./build_benzhi_docker.sh aroma-maintenance linux/amd64
./build_benzhi_docker.sh aroma-maintenance linux/arm64
```

`benzhi.Dockerfile` 与 `build_benzhi_docker.sh` 会由 `package-go-project` skill 的 GoLabel synthetic-project 模式生成并同步回根目录。

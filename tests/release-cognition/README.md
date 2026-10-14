# 认知实际来源验证

`verify-wal.ps1` 的公开来源、ref、scope、日期、时区和已确认的 strength/decay/initial-confidence 全部显式传入。独立 server 实际初始化库后关闭，生产 EventRecorder writer 保持 WAL 写入，另一个 snapshot 进程固定只读视图；保留空库、并发、完整采集和真实 Sleep 后四个实际 generation。脚本终态要求真实主题超过 20 个事件，独立核对全部主题的最高 attention 排序；样本不足退出失败，完整副本和本地诊断保留。

Workbench 不要求每个主题都超过 20 事件。该条件是此处分页正路径的样本前置，不能通过额外提交、修改 source ts/主题/权重或省略实际失败满足。先前实际不可变 generation 可以用于独立阶段 UI 核查，必须标记其对应阶段与限制，不能代表最新状态。

生产验收工具：

```powershell
go build -mod=readonly -o .verification/actual-public-cognition.exe ./cmd/workbench-snapshot/verify
./.verification/actual-public-cognition.exe --mode inspect --source <明确的实际源或本次已发布generation副本> --root <对应发布根> --output <自身.verification下安全输出目录>
```

inspect 使用 SQLite `mode=ro/query_only` 与原生产 Reader，逐项比较每个实际 event_id 的六字段、每页行数/总计/末尾空页，逐 event_id 核对真实 origin_key 与实际 commit.ts，检查原始公开 subject 没有出现在安全 DTO 中，直接 SQL 导出四路由恰好十值。`admission-safe.json`、`privacy-safe.json`、`origin-safe.json` 与计数均为安全聚合；SQL 期望和六字段文件不包含 raw、observation、源 payload 或底层异常。

本次 `.verification` 保留的原始公开 SQLite、Git checkout、Sleep 本地输出及进程日志不进入 Git 或 ZIP。正式 V4、实际 native/HTTP/DOM、前端完整浏览器及发布包验收分别由对应原 owner 汇报。

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

`-CrossDay` 使用实际上海日期，要求显式 `SleepDay` 为昨天；`-OutputDirectory` 仅允许当前工作树 `.verification` 内的新库。按明确批准的顺序完整采集选定公开 ref，再发布 `preSleep-local`，执行原生产 Sleep，最后发布 `local`（main）与 `closed-local`。closed 和 main 均包含完整已采今日事件，保存同一真实终态；14 日闭日表示 Sleep 的事件选择窗口，不能将 closed 当作 14 日纯截面。每个阶段在自身 safe 目录提供 SQL expected 与生产 Reader 派生结果，历史只继承实际发布状态。

跨日 inspector 使用 `--minimum-today 2 --closed-day 2026-10-14`，要求实际上海今日至少两个事件、昨天非空、唯一 Sleep 的开始及完成时间均为今日且与 Reader 全等。preSleep 使用 `--minimum-today 2`。传入 `--repository` 与 `--ref` 后，独立逐条核对实际可达 Git commit、SQL origin_key、event_id 和原始 commit.ts，确保选定人口没有遗漏或重复；四路由十值直接与生产 Reader 比较，两主题全部标量、六字段、全部分页和每主题至少两个实际历史点保持强断言。

Git mapper 的 `tokens=0` 保留其生产结果，不能据此推断模型 token usage。该公开生产重放没有模型、native 或 judge 调用。

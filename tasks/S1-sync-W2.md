# S1 同步 W2：本地 Agent 库存阶段

2026-10-10；分支 `s1-local-agents-1010`，基线 `20437708e43201e352d6c6926902e1363fd2ad3e`，工作树为同名 Orca 工作树。先读 AGENTS、HANDOFF、STATUS、S1-sync-local-agent-plan 与 SPEC §8.1。客户端为本机原生 Orca Codex；有效模型 `gpt-6.1-sol` 由主控 worker-read projection 明确观察后通过本 Dispatch 告知，Worker 没有读取私有配置或会话历史来反查。

## 已完成和接口交接

- 安装库存独立呈现 installed/configured/startable 与 SPEC 八项 native capability；PATH 发现不代表原生 discover。Fresh CLI 探测确认 Codex 0.162.0、Claude 2.1.238、OpenCode 1.18.35、Pi 1.0.1、Grok 1.0.34、Kimi 2.1.1、Qwen 0.24.6、Cursor GUI 3.23.23。额外 Gemini、cursor-agent 未在 PATH 找到；agent.exe 是 Grok 别名。所有 configured/startable/八项 native 为 unknown / NOT_RUN。
- `agents.NewInventory` PATH-only 初始化；显式 `RefreshCLI` 只允许固定 version/help，5 秒单命令期限、64KiB 保留输出、拥有的进程清理、固定原因码。缓存不会泄漏路径/帮助/原始诊断，不读凭据或会话目录。`Snapshot` 和 app `ListLocalAgents` 不启动命令，未连接 provider 明确 unsupported。取消保留已完成部分并标记未运行或不完整，不制造版本/能力成功。
- W0 已发布保护人类会话的 `GET /local-agents`、OpenAPI/生成客户端和 `InventoryProvider` / `LocalAgentInventory`；W2 通过普通 merge 消费精确 W0 `bb912efd6f28f9f3909d87ea9fc1f4acb04c9098`，不修改他轨文件。W0 负责入口组装，建议启动 refresh 总期限 10–15 秒，查询仍只读缓存；W3 消费同一契约。
- [探测文档](../docs/probes/local-agents.md) 保留旧清点与首次工具错误，记录 Codex app-server/Pi RPC 的本机版本证据、原生接续/送消息/取消/事件流可行路线及具体前置授权要求；没有造 scheduler、Mission、Swarm 或假原生会话。

## 分阶段提交

| 提交 | 内容 |
|---|---|
| `1fe4269` | 独立安装与能力 domain DTO，已 push，供 W0 契约依赖 |
| `331c4c9` | 安全库存适配器、Windows/Unix 探测边界、实测与公开原生接口调查，已 push |
| `763e183df35951a3ba919f593a51168d22fee522` | 消费 W0 契约后提供缓存查询 service、部分探测结果语义，已 push |

本报告提交另包含启动中断边界测试与部分探测明确标记；最终准确 SHA 通过 Orca Handoff 报告，不将文档 SHA 自写入同一提交。

## 实际验证

| 命令 | 结果 |
|---|---|
| `go test ./internal/adapters/agents ./internal/workspace/... ./internal/adapters/httpapi` | PASS；拒绝非 version/help 的 exec/resume/sessions/doctor、命令注入路径；实际临时 helper 超时退出，Windows 带空格 CMD 包装成功且取消后后代未写遗留标记；启动中断停止后续探测，缓存保留部分与未运行差别。HTTP 受保护库存测试由 W0 提供；不是假 native session 验收 |
| `go vet ./internal/adapters/agents ./internal/workspace/...` | PASS |
| `go build ./...` | PASS，包含合入 W0 公共端口后的 Go 项目 |
| `GOOS=linux go build ./internal/adapters/agents ./internal/workspace/...` | PASS（交叉编译；不是 Linux 运行时能力实测） |
| `ASTROCYTE_TEST_LOCAL_AGENT_CLI=1 go test ./internal/adapters/agents -run '^TestInventoryLiveCLI$' -count=1 -v` | PASS；首次适配器 fresh8 CLI version/help 4.54 秒；最终显式部分结果语义实测 5.10 秒（Go 命令总 5.454 秒），8 个版本/帮助通过、2 入口未找到，无模型或原生会话操作 |
| `git diff --check` | PASS |

首个探测失败保留：Python CMD 错误套引号使六个包装脚本 exit1；修正后的打印又遇 GBK UnicodeEncodeError、脚本 exit1。UTF-8/正确 argv 后全8 CLI 探测通过，原失败属于探测器；不得解释为 Agent 原生运行时成功或失败。默认 Go 测试跳过 opt-in CLI 实测；上述显式命令是单独 nonsecret 运行，没有将跳过算通过。

## 尚未授权、未执行和真实限制

用户仍未回答“完整 start/resume/send/stop/observe”或“先库存”、Agent 主动引用文件或具体项目根；W2 尚无可按其授权开发的本地项目/活动绑定范围。没有冻结绑定契约或 SQLite 迁移，没有读取私有历史/auth 配置、扫描整盘、控制既有会话、调用应用模型或真实媒体、运行浏览器验收。旧 UNKNOWN 模型请求未重发。

因此库存阶段已可集成，原 TASK 中项目/活动绑定与八项原生能力验收仍 NOT_RUN，不能据此声称完整本地接入或 S1 完成。未来原生操作需选定明确项目/读取根与目标原生 ID、确认占用交接、工具/外发规则及主控单场运行安排；安装 CLI 与官方协议文档不能证明 native readiness。本次 Windows 临时进程树测试不等于任意原生 daemon 或既有用户会话控制已验证；Linux 仅交叉编译。完整 pnpm check/build/E2E 和真实页面/API/SQLite 重启验收归主控/W0/W3 后续集成，本 Worker 未冒充已执行。

SQLite `adapter.go`、迁移、HTTP、组装、前端、全局 Agent 配置均由既有所有者维护；本轨没有编辑这些业务文件。已向主控和 W0 发送精确源码 SHA、命令结果、启动总期限/部分缓存语义及待答限制，等主控安排后续范围。

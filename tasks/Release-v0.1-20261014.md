# Astrocyte 合并产品与 release 包

2026-10-14 用户确认：继续网络中断任务，尽可能提高分工效率，今晚交付第一版产品及 release 包；合并 Astrocyte 新标准整改与 Workbench / Works Relay 两条线。当前环境为 Windows x64，先准备 Windows x64 ZIP；具体发布渠道问题仍待回复，包的构建与验收继续。最终交付保留普通 Git 历史并 push，不改正式 V4 运行树、锁定协议、实验数据和日程。

## 目标与验收

交付一个既有 Go 服务与既有 web 构成的合并产品，保留资料导入 / 人工沉淀、项目权限 / 进度、Workbench 认知查询、Works Relay 人工派发与 L1 派生会话事件。原始会话、凭据、正式库和用户数据不进入 release 包。未授权 Mission 扩展、用户自由工具、真实用户作品范围、活性窗口及 V4 冻结继续遵守当前决定；对应状态真实可见。

发布必须通过精确源码的适用检查、构建、完整默认浏览器验证和实际受控核心路径；原生 Codex 使用 Agent 创建的公开 Git 项目，文件 / 命令工具关闭，经真实控制器、SQLite、HTTP 与自产会话逐 event_id 核查。Windows ZIP 在新目录解压后实际启动，验证静态页面、session bootstrap、实际读写与重启持久化。包中不包含开发 node_modules、私人配置、数据库、测试原件或正式 V4 自动任务；必要可选能力必须明确依赖和配置。

## 长期开发分工

| 轨道 | 固定工作树 / 分支 | 独占 write_paths | 当前结果 |
|---|---|---|---|
| B 基础后端与导入整改 | standards-backend / standards-backend | 原 B 范围，排除下述 R / H 所有路径；保留 importer / 普通 agents / native events parser / tests/s1 / scripts/*.mjs / dependencies.lock.json；go.mod/go.sum 单归 B | 保留所有未提交修改，继续实际来源、成熟 parser、失败传播、测试与 SOURCE / REPORT |
| F 合并前端 | standards-frontend / standards-frontend | web/src/**、web/e2e/**、docs/standards/frontend.md | 保留已有整改，普通消费 Workbench / Relay 前端精确交付并返修真实 UI 与默认完整 E2E；旧 Relay F 不另派同域人员 |
| R 原生 Relay 运行时 | works-relay-runtime / songconmaisaix31-design/works-relay-runtime | internal/workspace/{domain,app}/relay*.go、internal/adapters/{sqlite,agents}/relay*.go、internal/adapters/relayworktree/**、agents/native.go/native_process.go、docs/access/runtime.md | 延续原 L1 当前修改与公开固定任务 / 来源 / 生命周期验证；普通原生控制器的可选观察端口归 R |
| H L1 采集与 HTTP | works-relay-http / songconmaisaix31-design/works-relay-http | internal/adapters/events/relay*.go、cmd/server/relay*.go、httpapi/relay*.go、workbench/app/relay*.go、sqlite/access_session*.go、contracts/openapi.yaml 与 relay examples、docs/access/http.md；cmd/server/main.go、httpapi/server.go 仅必要组装胶水 | 延续原 L1，真实 controller / SQLite / HTTP / event_id / 三分钟链路；依赖变更只 Handoff 给 B，config.go 和其他后端领域修改退 B |
| P 发布包 | release-package-20261014 / 独立同名分支 | 只写 release/**、docs/release/**；输出自身 .verification/ | 并行制作构建 / 启动 / 解压验证 / 说明；不修改业务或依赖，不接入正式数据；最终只对合并后的精确 SOURCE 打包 |

主控只维护本计划、状态、决定、验收与最终 Git。Orca 已在 20:08 重启，主控自身当前句柄经实际屏幕核实为 `term_494acbb9-67a5-4360-82dc-8514ce124990`，主整改 Run `run_7b2ce96bcd13` 已绑定 generation 4。原 B / F / R / H 终端当前不在运行时清单，旧 B / F 已由恢复登记为 abandoned / failed；已确认旧执行进程 PID 不再存在。恢复继续使用原工作树 / 分支 / Task，先核实实际终端身份及启动，保留全部未提交修改与第一失败；不因联系缺失单独推断退出，不重派尚在执行的写入者。

各轨负责实现、适用验证、报告及返修，SOURCE 和独立 REPORT 追加提交并 push。全部领域轨完成后派唯一集成人员，在独立 release 集成分支普通 --no-ff 合并精确交付，只补必要导入 / 路由 / 类型 / 配置；领域问题交原轨。统一完成 check / build / E2E 与 release 解压实际验收后，再记录精确分支 / SHA / 包路径 / 验证与剩余限制。未接受的 SOURCE 不进入发布。

## 已确定输入

- 正式验证根：validation/v1-20261012，仅计划 / 记录与既有 V4；不在本树合入 release 业务。
- Workbench 已验工程 SOURCE：5dd4b729c2b96d8538f0b5f91ed8fc2678fbf91b；业务线当前 workbench/v0.2-access / 8f94d502763239f2c3e1e6167ed390830459ff45。
- L1 R / H / F 当前分支 HEAD：139ce8090b868ee2764a13b0517fec6dee079424、37151e7a31092d3b267076a5bfa58ccb53f893e4、65ba685ea23e796c1810ff5ec13280dd0dda8f93；R 尚有未提交修改，均须原轨终稿与实际验收。
- 新标准 B 当前 HEAD：e1d0f28f75a3751da5fa39b7d69f9e0ac4763f39，存在未提交修改；F 当前 HEAD：e5b8d09，存在未提交修改。
- P Python SOURCE：a1089f94a13306e3cd250ed8facbb438d6e75a19；REPORT：91e3e59dcfc8024cd2e4955ac72f627e65c18e71，远端一致，主控独立 29 + 25 测试、36 文件规范 / 14 冻结 AST PASS。Python 实验其余 28 项仍 NOT_RUN。

现有 17321 部署保持独立，release 使用自身配置、数据目录与端口验收。Procmon 与两日自动快照按真实条件继续登记，发布包可用性检查不将其填为通过。

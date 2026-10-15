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

## 20:27 实际派发与进度

主控实际句柄 `term_494acbb9-67a5-4360-82dc-8514ce124990`；标准整改 / 发布包沿 `run_7b2ce96bcd13`，Relay 沿原 `run_d577f0d3fee1`。两个 Run 由同一主控在各自绑定后处理邮箱；每次消费须在切换前处理全部消息并 ACK 当前 Delivery，旧 generation 的 Delivery 不再次 ACK。

- B：原 Task `task_7b4e93c1a3c1` / 当前 Dispatch `ctx_7217777c7bf5` / 终端 `term_dcbad8eb-852c-4bcf-8040-08f32188e3ef`，ready / turnStart observed，已接受互斥的新范围。
- F：原 Task `task_7cbe30017760` / 当前 Dispatch `ctx_5bac511ada08` / 终端 `term_f9a085aa-aff2-47bd-bc5d-cdc3b87aab52`，ready / turnStart observed。阶段 SOURCE `f534298` / REPORT `521ed5b` 已 push，typecheck / lint / build / 112 Vitest PASS；默认浏览器阶段 PARTIAL。已普通消费 Workbench `67c5713` 与 Relay `6f61d87`，前端继续迁移新增旧模拟测试及稳定源码真实验证。主控已用编辑工具处理三份 Markdown 合并冲突并保留两线历史。
- R：原 Task `task_a51fab762c01` / 当前 Dispatch `ctx_b4fcb141ca00` / 终端 `term_e0e48c51-2379-4c5e-b0e0-9206b29fcffe`，ready / turnStart observed。首次 `ctx_99b2b8005248` bootstrap 明确 account routing timeout / 返回 shell，按实际退出证据精确 stop 后恢复，原件保留。实际 Codex 0.162.0 新 thread 启动 PASS，完整生命周期继续。
- H：原 Task `task_cfb7c9bde07d` / 当前 Dispatch `ctx_312d983772a9` / 终端 `term_7a9d814a-d487-4dc3-9708-f6bdc5e1479d`，ready / turnStart observed；本次独立实际服务 API 17342，空 registry 基线与自行持有的公开项目已创建，后续原生链路与 F DOM 继续。
- P 发布包：Task `task_76c43934b15c` / Dispatch `ctx_6d251e04179d` / 终端 `term_e67c7493-cef5-439e-9687-3197880820d7`，ready / turnStart observed；工作树 release-package-20261014，分支 `songconmaisaix31-design/release-package-20261014`。基线 `9d6f453090f1eaaf635c4b8767cc24820e6a1f32`，实际模型 GPT-6.1-Sol high fast。

主控独立核对 Workbench `5dd4b729` 与 B `e1d0f28` 的核心引擎 / schema / 迁移没有差异；原实验 Python 与文档属于不同父历史，组合仅进入新 release 分支。H 曾在隔离验证目录复制两份 config/auth，收到禁止复制指令后已删除本次副本并确认原文件不变；后续只由官方 CLI 自行鉴权，原始秘密不进入输出、Git、发布包或产品字段。最终认知成功覆盖使用本次自有 DB 与既有只读快照入口，不能读取正式库或旧私密快照替代。

# 产品模块执行与验收

## 责任与事实源

2026-10-14，M-product 当前 Task `task_b0197ca7ebdf` / Dispatch `ctx_b1cd2e0a39aa` / 终端 `term_580c19de-0c3d-49db-805e-8637184b9301`，子 Run `run_dbf1a68bc692`。工作树 `module-product-20261014`，分支 `songconmaisaix31-design/module-product-20261014`，接管基线 `dd60de11bca85662f5bb4a5b281cc7b42e6ffd2f`，接管时工作区无改动。

事实源为 [AGENTS.md](../../AGENTS.md)、[HANDOFF.md](../../HANDOFF.md)、[STATUS.md](../../STATUS.md)、[SPEC](../SPEC.md)、[TASKS](../TASKS.md)、[QUESTIONS](../QUESTIONS.md) 和 [本次发布任务](../../tasks/Release-v0.1-20261014.md)。本模块只修改此文件，负责资料导入、人工沉淀、项目权限、进度与全部既有 Web；实现、测试和返修交原 B/F。当前任务限定六组既有合并产品功能，历史 S0–S6 规划不扩大本次实现授权。

## 一页执行计划

| 开发人员 | 固定工作树 / 分支 | 当前原 Dispatch | 写入范围与下一步 |
|---|---|---|---|
| B | `standards-backend` / `standards-backend` | `task_7b4e93c1a3c1` / `ctx_7217777c7bf5` / `term_dcbad8eb-852c-4bcf-8040-08f32188e3ef` | 原 `internal/**` 与 `cmd`，排除 R/H/Q 域、全部 Python 和冻结核心；`tests/s1`、`tests/attention-engine`、`scripts/*.mjs`、`docs/standards/backend.md`；唯一 `go.mod/go.sum/dependencies.lock.json` 与 `cmd/server/config.go` owner。处理真实视频定位、实际模型/控制器测试、默认 Go 检查、Reader 发布路径与真实进度入口 |
| F | `standards-frontend` / `standards-frontend` | `task_7cbe30017760` / `ctx_5bac511ada08` / `term_f9a085aa-aff2-47bd-bc5d-cdc3b87aab52` | 唯一 `web/src/**`、`web/e2e/**`、`docs/standards/frontend.md` owner。消费精确公开契约与 B/R/H/Q 实际入口，完成未知状态持久化、全部既有 Web 和稳定源码默认完整 E2E |

1. 原 B/F 活跃 Dispatch 继续原 Run，本模块已直接发送接管与报告入口，读取回执待核实。每阶段源码 SOURCE 与独立 REPORT 普通提交并 push，首 RED / UNKNOWN / NOT_RUN 完整保留。
2. 依赖问题由 B 单独处理；跨原生问题交 M-relay `ctx_e6328a370d56` / `run_8bbda1c3974c`，认知与发布问题交 M-release `ctx_376d222c774b` / `run_ab36c13bc549`。管理人员不代写业务。
3. 总负责人接收原 worker_done 并交接已核实终端后，才在本子 Run 复用同一终端、工作树和分支创建后续 Task；不重新绑定总负责人 Run，不增加同写域人员。
4. 所有实际功能与适用检查通过后，向 M-release 提交精确 SOURCE/REPORT 及覆盖证据供唯一 I 普通合并；最终组合 check/build/默认 E2E、ZIP 新目录启动/读写/重启验收继续由原开发人员支持。模块阶段通过不结算本 Task，最终 GitHub Release 由总负责人接受具体结果后发布。

## 核心路径与当前结论

| 范围 | 必需验收 | 当前结论与返修 |
|---|---|---|
| 资料导入 | 公开论文真实 PDF/官方元数据；视频 summarize 原输出与真实时间定位；来源版本、去重/显式刷新、重启一致，零自动 Mission | B 已报告论文定向 PASS；视频 `BV1PReT6EEqR` 正文 32908 bytes 与重启 PASS，segments/spans 为 0，时间证据未通过 |
| 人工沉淀 | 固定输入版本，内容/主题/项目关联，待查问题、四维 unknown、later/reject/revise 各自语义、人工与机器关注分开，实际 HTTP/SQLite/UI 重载 | B 反馈定向 PASS；本次完整生产浏览器覆盖仍待 F；缺失证据不补造，不因历史测试通过填本次 PASS |
| 项目权限 | A 默认、B/C 显式许可、模型处理独立许可、人类会话/token 分离、撤销后拒绝新动作、旧 revision 冲突，真实持久化 | B 真实 token/revoke 定向 PASS；R CreateIntent revision 修复已交精确 SOURCE，H/F 组合实际复验待完成 |
| 项目进度 | 人工记录与真实获准项目 infer 均可用；依据引用独立核对、HTTP/SQLite/重启；未授权拒绝不能代替成功 | B 报告 README 单次生产 Codex infer PASS，证据位于 B `.verification/progress-infer-readme-first.log`；F 消费生产路径与实际 UI 验收待完成 |
| Workbench Web | Q 本次独立公开生产来源，至少 2 topic、同 topic 超过 20 真实 event 分页、2 实测 history points；独立 SQLexpected 与 UI 一致 | Q 时间阻塞已获总负责人明确历史生产重放裁决；具体快照根/SQLexpected 待 Q 交付，保留历史重放范围标签 |
| Relay 与 Today Web | 唯一 F 在同 API17342/UI17343 两 viewport 实际 DOM ready 后，H 注册并开始计时；真实 confirm/send/read/stop/release、409/unknown/换 Agent/三值状态及逐 event_id | 原空 registry/session200 基线已报 ready；R 修复后需 H 重编译/重启及 F 新 ready，旧首观察窗口保留，不重新标记成功 |
| 合并产品与发布支持 | 精确源码 `pnpm check`、`pnpm build`、`pnpm test:e2e` 与适用真实路径；Windows x64 ZIP 新目录启动/session bootstrap/读写/重启 | B 默认全 Go 仍 RED 网络 TLS/旧替代实现，F 默认完整 E2E PARTIAL；最终组合与包核心路径 NOT_RUN |

纯函数显示测试可以使用明确参数输入，外部网络/API/模型/控制器替代实现不能计为实际功能验收。不得新增 skip、删除已有成功断言或以静态页代替核心路径。资源限制可使用 `GOFLAGS=-p=1`、`GOMAXPROCS=2`、`NODE_OPTIONS=--max-old-space-size=512` 与 Playwright workers=1；稳定源码再运行完整默认套件。

## 精确交付与已确认协调

- B 已报告阶段 SOURCE `7f58adecc41070fbec4f88962cfb6fb1723679c6` / REPORT `a3f093ddd3488df9c71f2ad295724dba06e92efb` 已普通 push；前阶段 SOURCE `2644083960649d7c50262ec69fff730b411241a2` / REPORT `6a4415faaa1e092ae1eab42cffed4f3ffa090434` 保留。本模块首次远端核实遇到 `SSL_ERROR_SYSCALL`，尚未独立确认远端当前 HEAD。
- F 已报告 SOURCE `f534298` / REPORT `521ed5bd7f49156163bc8f946931b50b23349a8d` 已 push；typecheck/lint/build/112 Vitest PASS，完整 E2E PARTIAL。普通合并 Workbench 到 `67c5713`、Relay 到 `6f61d87`；新增生产迁移与真实窗口仍进行。
- R 本次权限修复 SOURCE `ce3a519efa5317e642d7488a776b850eebca1093` 已由 M-relay 报告普通 push/远端一致。CreateIntent 保存真实 settings revision/A-B-C；旧 intent 权限变化阻断，替换确认 revision 不能绕过。F 前端比较必须保留，H 正消费并重启固定 API。
- 总负责人 `msg_ec33426ae20d` 临时授权 B 在 F 工作树只用文件编辑工具处理 H 普通 merge 的 `go.mod` 冲突、只 stage 此文件，F 完成普通 merge；依赖 union 保留双方必要库。开始/结束通知 F，完成后临时写域撤回，`go.sum` 如需变化先 Handoff 本模块核实。
- 总负责人 `msg_5c5fa8031f79` 授权原 B 整改 `internal/adapters/sqlite/workbench.go` 的个人正式路径常量。保护检查针对显式配置源与 owned snapshot，保留源只读/immutable/query_only、同文件与 alias 拒绝；默认服务不访问个人正式路径。实际自有 DB、HTTP、alias 和发布 binary 扫描须通过；不修改冻结领域或现有部署。

## 隐私、冻结与剩余限制

原始来源、认证文件、模型输出、流与错误原件只保留本机专门证据目录，不进 HTTP/Git/包/stdout；文档只写安全派生和本机指针。浏览器未知派发锁仅保存公开操作身份与 unknown 状态，跨组件重开/刷新保留，禁止自动重试。

旧正式 V4、冻结 schema/engine/mapper/权重/日程与 17321 不变；真实用户作品、自由工具、未定 Mission 和活性窗口不扩大。已授权公开固定项目实际模型/HTTP 操作沿当前许可，原生文件/命令工具关闭。Procmon、两次真实自动快照、人评及七日科研结果继续按真实状态记录，不能由 release 可用性验收替代。

当前未执行最终组合核心路径与 ZIP 验收，未自行发布 Release。模块文档适用检查为相对链接存在性和 `git diff --check`；业务检查结论以同源实际运行结果更新。

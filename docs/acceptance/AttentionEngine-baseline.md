# Attention Engine 接管基线

分支 `attention-engine/v1-20261011`，业务基线 `dab8fea`，接管文档 `0e14c7d9d4f3c43bf473a25e0f6290e81f14be9a`。客户端 Codex 主控；原八轨 OpenCode / DeepSeek V4 Pro（终端观察）。

2026-10-11 21:43，`GOFLAGS=-p=1 pnpm check` 首场 FAIL：`TestPaperDOMInstalledPinAndIsolation` 无法解析已锁定依赖 `entities/lib/index.js`（MODULE_NOT_FOUND）；其余已运行 Go 包通过，尚未运行后续契约/TS等阶段。没有业务源码变更。

执行 `pnpm install --frozen-lockfile` 和同锁 `pnpm install --frozen-lockfile --force`（均 exit0）。随后 `GOFLAGS=-p=1 GOMAXPROCS=2 pnpm check` 的 importer 包通过（6.244 秒），Go、依赖核验、架构检查和 OpenAPI lint 通过；全场在 `validate-examples` 因 Node `ZoneAllocationFailed/OutOfMemory` exit134 再次 FAIL，后续 TS/lint/Vitest/API 阶段未运行。限制 `NODE_OPTIONS=--max-old-space-size=256` 后单独执行 `node scripts/validate-examples.mjs`，241 个示例 PASS。两次整场 FAIL 均保留，定向恢复不等于整场通过。

W0 恢复时默认并行 `go build ./...` 遇到 Windows paging file too small；已要求各轨单包编译并限制进程并发，未修改分页文件或系统设置。

第三场 `GOFLAGS=-p=1 GOMAXPROCS=2 NODE_OPTIONS=--max-old-space-size=256 pnpm check`：Go/vet、依赖、架构、契约及 S1 API 14/14 PASS；TypeScript 因 256 MB 堆上限 exit134 FAIL，后续 lint/Vitest 未运行。在 `web` 工作目录以 512 MB 堆单独执行 `node node_modules/typescript/bin/tsc -b` PASS；整场第三次失败仍保留。

后续单独执行 `NODE_OPTIONS=--max-old-space-size=512 pnpm --dir web lint` PASS；同堆限制 `pnpm --dir web exec vitest run --maxWorkers=1` 12 文件 / 83 测试 PASS。root 尚无完整单场 `pnpm check` PASS；这些组件复验不合成一场新绿记录。

root 使用 Node SQLite `readOnly:true` 核验固定个人库 `C:/Users/DW/AppData/Roaming/astrocyte/state.sqlite`：缓存 partial，严格匹配 Astrocyte 根存在，native contributors=[]。确认 W1 所选缓存回放 0 eligible 是真实结果，未扩大原始会话读取或修改个人库。

本次读取 W7 猜测的参考图路径 `C:/Users/DW/Downloads/个人作品蜂群接入后.png` 失败（文件不存在）。已要求复用现有看板及任务明确的五能力，不能声称对该图完成精确复刻。

## 波 A 独立检查（2026-10-11 21:58）

W2 `acb4561` 的 30 天 fixture，由 root 以独立 Node 公式重算：30 点 A/M/confidence 最大绝对误差为 0，10 点触底、5 点触顶。相同工作树执行 `GOFLAGS=-p=1 GOMAXPROCS=2 go test ./internal/attention/domain ./tests/attention-engine` PASS。该 fixture 使用显式 λ=0.1，仅验证纯数学；尚未覆盖事件持久化、重复消费、主题生命周期及冷重启，M2 未完整通过。fixture 的原始一次性生成使用 Python（Worker 如实记录）；用户在 QUESTIONS 后续已允许 Python，不新增 Python 项目依赖或默认验收步骤。

## M0 P1 依赖复验（2026-10-11 22:17）

SOURCE `47f47f38d7ec0c5c3cbf30a757ed2c23e8c774f0` 普通合入集成 `2268f4920973128068f46a182dfddfb040a1dbc2`，root 运行 `GOFLAGS=-p=1 GOMAXPROCS=2 go test -count=1 ./tests/attention-engine ./internal/attention/domain ./internal/attention/app ./internal/adapters/sqlite ./internal/adapters/httpapi`，五包 PASS（M0 harness 0.887s、SQLite 4.958s、HTTP 0.565s）。范围为 schema/迁移/持久化基础和既有 HTTP，未组装人工关口仍501，其他轨应用未合，不宣称完整 M1–M6/W7。

W0 独立报告47f的 `pnpm check`/`pnpm build` PASS（Node堆512、Go单并发）；不是root整场复验。旧 Windows paging failure 原始完整stderr文件未留存，仅已知错误片段保留，不补造日志路径。

## P2 组装基础复验（2026-10-11 22:30）

SOURCE `5d356370c4157e74d96388afd098d5fea14dc799` 普通合入并 push 为 `05d853605256e6540d9ef43f63eeb6530b7fe0a9`（含已授权 W7 稳定 Go 阶段）。root 五包 `go test -count=1` PASS：M0/W7 harness、httpapi、sqlite、workspace/app、worksscan。实际 HTTP 用生产 Works 服务/SQLite及合成临时目录输入验证边界与重启；请求生命周期/sink/真实 UI 仍待后续 owner 阶段。

该精确集成源 root `GOFLAGS=-p=1 GOMAXPROCS=2 NODE_OPTIONS=--max-old-space-size=512 pnpm check` 单场 PASS：Go 格式/vet/全包测试、依赖、21包架构、275契约示例与生成漂移、S1 API14/14、TS/lint、Vitest13文件/90测试、diff。完整输出 `%TEMP%/astrocyte-root-p2-check-20261011-2227.log`。前三场失败保持，不改写历史；此场不覆盖尚未合入的动力学/Memory/MCP/Sleep实际服务，也不算完整M1–M6/W7完成。P2本轮root build/E2E尚未执行。

## M2 实际应用逐点复算（2026-10-11 22:39）

W2 SOURCE `23bc4d718776bd8d47b8f6d4d022f1030ad3794c` 的 `M2_application_state.json` 使用 W1事件/Signal、实际SQLite和两次子进程重载（第10/20天）。root 以独立 Node 公式从 initial + signals 重算全部30点，未调用 Worker 的 `M2_reference.mjs`：A、momentum、confidence最大绝对误差均0，updated_at/last_active逐点一致，10点floor、5点ceiling。初始A=.05/C=.5/M=0、lambda=.1/day、成功confidence增量=.1均为该合成流显式参数，不是产品默认。

root NodeSQLite readOnly + PRAGMA query_only=ON 核对留存 `%TEMP%/astrocyte-m2-01d52594-7909-4630-9aad-ad4f67212631.sqlite`，integrity_check=ok，9事件/9信号/1节点；全部Signal及原始BLOB存在、最终StoredAttentionNode JSON与artifact逐字段一致。最终A=.05、M=-.000008972512673807757、C=.5，updated_at=1769731200000、last_active=1768953600000。这是任务书M2合成应用验收，不是真实30天行为。W2最终 `2c148c1034ec28b5d787309f1320198706d97e4a` 再补已有receipt复用和Sleep出生时间顺序；HTTP/MCP/Sleep集成验收仍独立。

## M4 独立排序与留存库核对（2026-10-11 22:51）

W4 SOURCE `82e4340e50943d355ecd81a2a47d65b7e3929f5a` / REPORT `a6f5c6708f77cf360990b2e613e9a5dda907e325` 已结算。root 从 `%TEMP%/astrocyte-m4-6b799aaebe6c468896492f068d317ad6` 的实际 retrieval.json 独立 Node 重算 relevance × exp(-lambdaR × ageSeconds) × Attention × AccessBoost，两场四个得分最大绝对误差0。lambdaR=0 排序 a,b（1.8/.225），lambdaR=1.5 排序 b,a（.000027767205919502902/.000011059582235990778）；返回项仅 p1 confirmed，年龄由明确 created_at 推导。这些参数为合成验收输入，不决定产品时间基准或默认值。

root Node SQLite readOnly + query_only 核对两份留存库 integrity=ok；排序库的节点 JSON/A/updated_at/version 与检索前后 artifact 完全一致。版本库保留内容 one/two 的两条不可变固化记录；human 引用仍固定 crystallized_version=1 / knowledge_version=3，当前知识 version=6，不随新版本漂移。该检查不读取个人源、不写库；sqlite-vec CUT 后实际模式为 topic_keyword，不宣称向量相似度或线上检索质量。

## 后续 owner 集成检查限制（2026-10-11 22:51）

W0 组装源 `271d18cc4d7853bd1a9bc96875b8875733889837` 的实际人工经验 HTTP 重复/冷重启与已提交 MCP 动力学 pending 定向 PASS，尚未 root 最终合入。其完整检查首次在 gofmt 的320条绝对路径 Windows argv 触发 ENAMETOOLONG，原 owner 正补有界相对路径批次；第二场通过 Go/vet/测试/modverify 后 summarize 固定0.25.1版本探测 ETIMEDOUT，后续阶段 NOT_RUN。同命令单独测量2.718秒并 PASS，原60秒超时未改；环境争用只是假设，不能合成完整整场绿色。

## M6 推荐输入首次实际集成 RED（2026-10-11 22:57）

root 独立读取 W0 实际 `%TEMP%/astrocyte-M6-W0-20261011/M6-week-result.json`：七日焦点 research→implementation，但最终两推荐的 inputs.Attention=0/score=0，顺序恰为字母序。此前 W0 测试只有顺序断言并报告PASS；主控不接受该推荐结论。根因是 W4 KnowledgeAttention 按topic查询后解码为 flat domain.AttentionNode，实际 W2 SaveNode 存储 app.StoredAttentionNode 的 nested node，解码无报错但注意力/关联为空。不是 id/topic 查询条件错误；root 首假设已被实际 SQL 排除。

root 保存原270812字节输出为 `M6-week-first-zero-attention.json`（未覆盖既有文件），W0另保存 `M6-first-zero-ranking.json`。原owner W4新限定Task修wrapper/legacy-flat读取并覆盖实际W2节点+query_only回归；W0加强 inputs.Attention=实际图A、非零得分、独立公式复算。旧M4 flat fixture数学核对保持局部PASS，不替代本次正式写入格式的实际跨轨验收。

root另只读核对W7最终留存 `%TEMP%/astrocyte-s1-aijSTJ/data/state.sqlite` integrity=ok，两条pending/raw BLOB分别manual分组与machine同步，字节/observation与独立artifact完全一致，Signal和节点0；10条works receipt均accepted且input/result为BLOB。最终三尺寸报告writes/errors均空。看过390截图，作品/未知状态/手册区可辨认；浮动页头在长截图中叠加保持原限制，不宣称用户主观视觉验收。
